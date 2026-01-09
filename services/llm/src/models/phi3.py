import torch
from transformers import AutoModelForCausalLM, AutoTokenizer, BitsAndBytesConfig
from typing import Optional, List, AsyncGenerator
import asyncio

from ..utils.logger import logger


class Phi3Model:
    def __init__(
        self,
        model_name: str = "microsoft/Phi-3-mini-4k-instruct",
        cache_dir: str = "./model_cache",
        load_in_8bit: bool = True,
        load_in_4bit: bool = False,
        device_map: str = "auto"
    ):
        self.model_name = model_name
        self.cache_dir = cache_dir
        self.load_in_8bit = load_in_8bit
        self.load_in_4bit = load_in_4bit
        self.device_map = device_map

        self.model = None
        self.tokenizer = None
        self._lock = asyncio.Lock()

    async def load(self):
        logger.info(f"Loading model: {self.model_name}")

        quantization_config = None
        if self.load_in_4bit:
            quantization_config = BitsAndBytesConfig(
                load_in_4bit=True,
                bnb_4bit_compute_dtype=torch.float16,
                bnb_4bit_use_double_quant=True,
                bnb_4bit_quant_type="nf4"
            )
        elif self.load_in_8bit:
            quantization_config = BitsAndBytesConfig(
                load_in_8bit=True
            )

        self.tokenizer = AutoTokenizer.from_pretrained(
            self.model_name,
            cache_dir=self.cache_dir,
            trust_remote_code=True
        )

        if self.tokenizer.pad_token is None:
            self.tokenizer.pad_token = self.tokenizer.eos_token

        self.model = AutoModelForCausalLM.from_pretrained(
            self.model_name,
            cache_dir=self.cache_dir,
            device_map=self.device_map,
            quantization_config=quantization_config,
            torch_dtype=torch.float16,
            trust_remote_code=True
        )

        self.model.eval()
        logger.info("Model loaded successfully")

    def format_prompt(self, prompt: str) -> str:
        return f"<|user|>\n{prompt}<|end|>\n<|assistant|>\n"

    async def generate(
        self,
        prompt: str,
        max_tokens: int = 2048,
        temperature: float = 0.7,
        top_p: float = 0.9,
        stop_sequences: Optional[List[str]] = None
    ) -> dict:
        async with self._lock:
            formatted_prompt = self.format_prompt(prompt)

            inputs = self.tokenizer(
                formatted_prompt,
                return_tensors="pt",
                padding=True,
                truncation=True,
                max_length=4096
            ).to(self.model.device)

            with torch.no_grad():
                outputs = self.model.generate(
                    **inputs,
                    max_new_tokens=max_tokens,
                    temperature=temperature,
                    top_p=top_p,
                    do_sample=temperature > 0,
                    pad_token_id=self.tokenizer.pad_token_id,
                    eos_token_id=self.tokenizer.eos_token_id
                )

            generated_ids = outputs[0][inputs.input_ids.shape[1]:]
            text = self.tokenizer.decode(generated_ids, skip_special_tokens=True)

            if stop_sequences:
                for stop_seq in stop_sequences:
                    if stop_seq in text:
                        text = text[:text.index(stop_seq)]
                        break

            return {
                "text": text.strip(),
                "token_count": len(generated_ids),
                "model": self.model_name,
                "finish_reason": "stop"
            }

    async def generate_stream(
        self,
        prompt: str,
        max_tokens: int = 2048,
        temperature: float = 0.7,
        top_p: float = 0.9,
        stop_sequences: Optional[List[str]] = None
    ) -> AsyncGenerator[str, None]:
        formatted_prompt = self.format_prompt(prompt)

        inputs = self.tokenizer(
            formatted_prompt,
            return_tensors="pt",
            padding=True
        ).to(self.model.device)

        generated_text = ""

        async with self._lock:
            with torch.no_grad():
                for _ in range(max_tokens):
                    outputs = self.model(
                        **inputs,
                        use_cache=True
                    )

                    next_token_logits = outputs.logits[:, -1, :]

                    if temperature > 0:
                        next_token_logits = next_token_logits / temperature
                        probs = torch.softmax(next_token_logits, dim=-1)
                        next_token = torch.multinomial(probs, num_samples=1)
                    else:
                        next_token = torch.argmax(next_token_logits, dim=-1, keepdim=True)

                    if next_token.item() == self.tokenizer.eos_token_id:
                        break

                    token_text = self.tokenizer.decode(next_token[0], skip_special_tokens=True)
                    generated_text += token_text

                    if stop_sequences:
                        should_stop = False
                        for stop_seq in stop_sequences:
                            if stop_seq in generated_text:
                                should_stop = True
                                break
                        if should_stop:
                            break

                    yield token_text

                    inputs = {
                        "input_ids": torch.cat([inputs["input_ids"], next_token], dim=-1),
                        "attention_mask": torch.cat([
                            inputs["attention_mask"],
                            torch.ones((1, 1), device=inputs["attention_mask"].device)
                        ], dim=-1)
                    }

                    await asyncio.sleep(0)

    def count_tokens(self, text: str) -> int:
        return len(self.tokenizer.encode(text))

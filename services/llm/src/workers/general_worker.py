import json
import re
from typing import List, Dict, Any
from .base import BaseWorker, WorkerType, WorkerConfig


class GeneralWorker(BaseWorker):
    def __init__(self, config: WorkerConfig = None):
        super().__init__(config or WorkerConfig(
            temperature=0.7,
            max_tokens=2048
        ))

    @property
    def worker_type(self) -> WorkerType:
        return WorkerType.GENERAL

    def get_system_prompt(self) -> str:
        return """You are a helpful AI assistant specialized in analyzing text and answering questions.
Provide clear, accurate, and well-structured responses.
When extracting information, be thorough but precise.
If uncertain, express your confidence level."""

    def get_extraction_prompt(self, text: str) -> str:
        return f"""Analyze and respond to the following:

{text}"""

    def parse_response(self, response: str) -> List[Dict[str, Any]]:
        return [{
            'type': 'general_response',
            'content': response,
            'confidence': 0.8
        }]

    async def chat(self, message: str, context: str = None) -> str:
        prompt = message
        if context:
            prompt = f"Context:\n{context}\n\nQuestion: {message}"

        result = await self.extract(prompt)
        return result.raw_response

    async def summarize(self, text: str, max_length: int = 200) -> str:
        prompt = f"""Summarize the following text in {max_length} words or less:

{text}

Summary:"""
        result = await self.extract(prompt)
        return result.raw_response

    async def classify(self, text: str, categories: List[str]) -> Dict[str, float]:
        categories_str = ", ".join(categories)
        prompt = f"""Classify the following text into one or more of these categories: {categories_str}

Text: {text}

Return JSON with category names as keys and confidence scores (0-1) as values.

JSON:"""
        result = await self.extract(prompt)

        try:
            json_match = re.search(r'\{[\s\S]*\}', result.raw_response)
            if json_match:
                return json.loads(json_match.group())
        except json.JSONDecodeError:
            pass

        return {cat: 0.0 for cat in categories}

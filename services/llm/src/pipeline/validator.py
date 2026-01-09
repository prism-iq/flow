import httpx
from typing import Optional
import asyncio

from ..utils.logger import logger


class HaikuValidator:
    """Uses Haiku to validate/improve LLM outputs."""

    def __init__(self, api_key: Optional[str] = None, enabled: bool = False):
        self.api_key = api_key
        self.enabled = enabled and api_key is not None
        self.base_url = "https://api.anthropic.com/v1"
        self.model = "claude-3-haiku-20240307"

    async def validate(self, prompt: str, response: str) -> dict:
        """Validate response quality and coherence."""
        if not self.enabled:
            return {"valid": True, "score": 1.0, "feedback": None}

        validation_prompt = f"""Evaluate this AI response for quality:

Original prompt: {prompt}

AI Response: {response}

Rate the response on:
1. Relevance (0-10): Does it address the prompt?
2. Coherence (0-10): Is it logically consistent?
3. Completeness (0-10): Does it fully answer?

Return JSON: {{"relevance": X, "coherence": X, "completeness": X, "feedback": "brief feedback"}}"""

        try:
            async with httpx.AsyncClient() as client:
                result = await client.post(
                    f"{self.base_url}/messages",
                    headers={
                        "x-api-key": self.api_key,
                        "anthropic-version": "2023-06-01",
                        "Content-Type": "application/json"
                    },
                    json={
                        "model": self.model,
                        "max_tokens": 256,
                        "messages": [{"role": "user", "content": validation_prompt}]
                    },
                    timeout=10.0
                )

                if result.status_code == 200:
                    data = result.json()
                    content = data["content"][0]["text"]
                    import json
                    scores = json.loads(content)
                    avg_score = (scores["relevance"] + scores["coherence"] + scores["completeness"]) / 30
                    return {
                        "valid": avg_score >= 0.5,
                        "score": avg_score,
                        "feedback": scores.get("feedback"),
                        "details": scores
                    }
        except Exception as e:
            logger.error(f"Validation error: {e}")

        return {"valid": True, "score": 1.0, "feedback": None}

    async def improve(self, prompt: str, response: str, feedback: str) -> str:
        """Use Haiku to improve a response based on feedback."""
        if not self.enabled:
            return response

        improve_prompt = f"""Improve this AI response based on feedback:

Original prompt: {prompt}
Original response: {response}
Feedback: {feedback}

Provide an improved response that addresses the feedback while maintaining accuracy."""

        try:
            async with httpx.AsyncClient() as client:
                result = await client.post(
                    f"{self.base_url}/messages",
                    headers={
                        "x-api-key": self.api_key,
                        "anthropic-version": "2023-06-01",
                        "Content-Type": "application/json"
                    },
                    json={
                        "model": self.model,
                        "max_tokens": 2048,
                        "messages": [{"role": "user", "content": improve_prompt}]
                    },
                    timeout=30.0
                )

                if result.status_code == 200:
                    data = result.json()
                    return data["content"][0]["text"]
        except Exception as e:
            logger.error(f"Improvement error: {e}")

        return response

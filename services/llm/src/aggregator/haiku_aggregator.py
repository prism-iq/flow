import httpx
import json
import asyncio
from typing import List, Dict, Any, Optional
from dataclasses import dataclass
from enum import Enum

from ..utils.logger import logger


class ValidationStatus(Enum):
    VALID = "valid"
    INVALID = "invalid"
    NEEDS_REVIEW = "needs_review"
    MERGED = "merged"


@dataclass
class ValidationResult:
    status: ValidationStatus
    confidence: float
    entities: List[Dict[str, Any]]
    corrections: List[Dict[str, Any]]
    reasoning: str
    graph_operations: List[Dict[str, Any]]


class HaikuAggregator:
    def __init__(
        self,
        api_key: Optional[str] = None,
        base_url: str = "https://api.anthropic.com/v1",
        model: str = "claude-3-haiku-20240307",
        enabled: bool = True
    ):
        self.api_key = api_key
        self.base_url = base_url
        self.model = model
        self.enabled = enabled and api_key is not None
        self._request_count = 0

    async def validate_and_aggregate(
        self,
        entities: List[Dict[str, Any]],
        original_text: str,
        context: Dict[str, Any] = None
    ) -> ValidationResult:
        if not self.enabled:
            return self._passthrough_validation(entities)

        prompt = self._build_validation_prompt(entities, original_text, context)

        try:
            response = await self._call_haiku(prompt)
            return self._parse_validation_response(response, entities)
        except Exception as e:
            logger.error(f"Haiku validation failed: {e}")
            return self._passthrough_validation(entities)

    async def _call_haiku(self, prompt: str) -> str:
        async with httpx.AsyncClient() as client:
            response = await client.post(
                f"{self.base_url}/messages",
                headers={
                    "x-api-key": self.api_key,
                    "anthropic-version": "2023-06-01",
                    "Content-Type": "application/json"
                },
                json={
                    "model": self.model,
                    "max_tokens": 2048,
                    "messages": [{"role": "user", "content": prompt}]
                },
                timeout=30.0
            )

            self._request_count += 1

            if response.status_code != 200:
                raise Exception(f"Haiku API error: {response.status_code}")

            data = response.json()
            return data["content"][0]["text"]

    def _build_validation_prompt(
        self,
        entities: List[Dict[str, Any]],
        original_text: str,
        context: Dict[str, Any] = None
    ) -> str:
        entities_json = json.dumps(entities, indent=2, default=str)

        return f"""You are a validation AI. Review extracted entities for accuracy and consistency.

ORIGINAL TEXT:
{original_text[:2000]}

EXTRACTED ENTITIES:
{entities_json}

{f"CONTEXT: {json.dumps(context)}" if context else ""}

TASKS:
1. Verify each entity is correctly extracted from the text
2. Check for duplicates that should be merged
3. Identify any errors or inconsistencies
4. Suggest corrections where needed
5. Generate graph operations (CREATE node, CREATE relationship)

Return JSON:
{{
    "status": "valid|invalid|needs_review",
    "confidence": 0.0-1.0,
    "validated_entities": [...],
    "corrections": [
        {{"entity_index": N, "field": "...", "old_value": "...", "new_value": "...", "reason": "..."}}
    ],
    "duplicates_to_merge": [
        {{"indices": [N, M], "merged_entity": {{...}}}}
    ],
    "graph_operations": [
        {{"operation": "CREATE_NODE", "label": "Person", "properties": {{...}}}},
        {{"operation": "CREATE_EDGE", "from": "...", "to": "...", "label": "SENT", "properties": {{...}}}}
    ],
    "reasoning": "Brief explanation"
}}

JSON:"""

    def _parse_validation_response(
        self,
        response: str,
        original_entities: List[Dict[str, Any]]
    ) -> ValidationResult:
        try:
            import re
            json_match = re.search(r'\{[\s\S]*\}', response)
            if not json_match:
                return self._passthrough_validation(original_entities)

            data = json.loads(json_match.group())

            status_map = {
                "valid": ValidationStatus.VALID,
                "invalid": ValidationStatus.INVALID,
                "needs_review": ValidationStatus.NEEDS_REVIEW,
            }

            validated = data.get("validated_entities", original_entities)

            for merge in data.get("duplicates_to_merge", []):
                indices = merge.get("indices", [])
                merged = merge.get("merged_entity")
                if merged and indices:
                    for idx in sorted(indices, reverse=True):
                        if idx < len(validated):
                            validated.pop(idx)
                    validated.append(merged)

            return ValidationResult(
                status=status_map.get(data.get("status", "valid"), ValidationStatus.VALID),
                confidence=data.get("confidence", 0.8),
                entities=validated,
                corrections=data.get("corrections", []),
                reasoning=data.get("reasoning", ""),
                graph_operations=data.get("graph_operations", [])
            )

        except json.JSONDecodeError:
            logger.warning("Failed to parse Haiku response as JSON")
            return self._passthrough_validation(original_entities)

    def _passthrough_validation(self, entities: List[Dict[str, Any]]) -> ValidationResult:
        graph_ops = []
        for entity in entities:
            worker_type = entity.get('_source_worker', 'general')
            label_map = {
                'date': 'Date',
                'person': 'Person',
                'org': 'Organization',
                'amount': 'Amount',
            }
            label = label_map.get(worker_type, 'Entity')

            props = {k: v for k, v in entity.items() if not k.startswith('_')}
            graph_ops.append({
                "operation": "CREATE_NODE",
                "label": label,
                "properties": props
            })

        return ValidationResult(
            status=ValidationStatus.VALID,
            confidence=0.7,
            entities=entities,
            corrections=[],
            reasoning="Passthrough validation (Haiku disabled)",
            graph_operations=graph_ops
        )

    async def generate_summary(
        self,
        entities: List[Dict[str, Any]],
        original_text: str
    ) -> str:
        if not self.enabled:
            return self._generate_basic_summary(entities)

        prompt = f"""Summarize the key information extracted from this text:

TEXT:
{original_text[:1500]}

ENTITIES FOUND:
- People: {len([e for e in entities if e.get('_source_worker') == 'person'])}
- Organizations: {len([e for e in entities if e.get('_source_worker') == 'org'])}
- Dates: {len([e for e in entities if e.get('_source_worker') == 'date'])}
- Amounts: {len([e for e in entities if e.get('_source_worker') == 'amount'])}

Provide a 2-3 sentence summary of the main points and relationships discovered."""

        try:
            return await self._call_haiku(prompt)
        except Exception:
            return self._generate_basic_summary(entities)

    def _generate_basic_summary(self, entities: List[Dict[str, Any]]) -> str:
        counts = {}
        for e in entities:
            worker = e.get('_source_worker', 'general')
            counts[worker] = counts.get(worker, 0) + 1

        parts = []
        if counts.get('person'):
            parts.append(f"{counts['person']} person(s)")
        if counts.get('org'):
            parts.append(f"{counts['org']} organization(s)")
        if counts.get('date'):
            parts.append(f"{counts['date']} date(s)")
        if counts.get('amount'):
            parts.append(f"{counts['amount']} amount(s)")

        return f"Extracted: {', '.join(parts) if parts else 'no entities'}"

    async def validate(
        self,
        entities: List[Dict[str, Any]],
        original_text: str
    ) -> List[Dict[str, Any]]:
        """Simple validate method that returns validated entities."""
        result = await self.validate_and_aggregate(entities, original_text)
        return result.entities

    @property
    def stats(self) -> Dict[str, Any]:
        return {
            "enabled": self.enabled,
            "model": self.model,
            "request_count": self._request_count,
        }

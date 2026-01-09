import re
import json
from typing import List, Dict, Any
from datetime import datetime
from .base import BaseWorker, WorkerType, WorkerConfig


class DateWorker(BaseWorker):
    DATE_PATTERNS = [
        r'\b(\d{1,2})[/-](\d{1,2})[/-](\d{2,4})\b',
        r'\b(\d{4})[/-](\d{1,2})[/-](\d{1,2})\b',
        r'\b(January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d{1,2}),?\s+(\d{4})\b',
        r'\b(\d{1,2})\s+(January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d{4})\b',
        r'\b(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+(\d{1,2}),?\s+(\d{4})\b',
    ]

    def __init__(self, config: WorkerConfig = None):
        super().__init__(config or WorkerConfig())

    @property
    def worker_type(self) -> WorkerType:
        return WorkerType.DATE

    def get_system_prompt(self) -> str:
        return """You are a specialized date extraction AI. Your task is to identify and extract all dates, time references, and temporal expressions from text.

Extract:
- Specific dates (e.g., "January 15, 2024", "2024-01-15", "01/15/24")
- Relative dates (e.g., "next Tuesday", "last week", "in 3 days")
- Time periods (e.g., "Q1 2024", "fiscal year 2023", "summer 2024")
- Deadlines and milestones
- Meeting times and schedules

Output JSON array with: date_string, normalized_date (ISO format if possible), date_type, context"""

    def get_extraction_prompt(self, text: str) -> str:
        return f"""Extract all dates and temporal references from this text:

TEXT:
{text}

Return a JSON array of objects with fields:
- date_string: the exact date text found
- normalized_date: ISO 8601 format (YYYY-MM-DD) if possible, null otherwise
- date_type: one of [specific, relative, period, deadline, meeting]
- context: brief context of what the date refers to
- confidence: 0.0 to 1.0

JSON:"""

    def parse_response(self, response: str) -> List[Dict[str, Any]]:
        entities = []

        try:
            json_match = re.search(r'\[[\s\S]*\]', response)
            if json_match:
                parsed = json.loads(json_match.group())
                if isinstance(parsed, list):
                    entities.extend(parsed)
        except json.JSONDecodeError:
            pass

        for pattern in self.DATE_PATTERNS:
            matches = re.finditer(pattern, response, re.IGNORECASE)
            for match in matches:
                date_str = match.group()
                if not any(e.get('date_string') == date_str for e in entities):
                    normalized = self._normalize_date(date_str)
                    entities.append({
                        'date_string': date_str,
                        'normalized_date': normalized,
                        'date_type': 'specific',
                        'context': '',
                        'confidence': 0.8 if normalized else 0.5
                    })

        return entities

    def _normalize_date(self, date_str: str) -> str:
        formats = [
            '%m/%d/%Y', '%m-%d-%Y', '%Y-%m-%d', '%Y/%m/%d',
            '%B %d, %Y', '%d %B %Y', '%b %d, %Y',
            '%m/%d/%y', '%m-%d-%y',
        ]

        for fmt in formats:
            try:
                dt = datetime.strptime(date_str.strip(), fmt)
                return dt.strftime('%Y-%m-%d')
            except ValueError:
                continue

        return None

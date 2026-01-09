import re
import json
from typing import List, Dict, Any
from .base import BaseWorker, WorkerType, WorkerConfig


class PersonWorker(BaseWorker):
    NAME_PATTERNS = [
        r'\b([A-Z][a-z]+)\s+([A-Z][a-z]+)\b',
        r'\b([A-Z][a-z]+)\s+([A-Z]\.)\s+([A-Z][a-z]+)\b',
        r'\b(Mr\.|Mrs\.|Ms\.|Dr\.|Prof\.)\s+([A-Z][a-z]+)\s+([A-Z][a-z]+)\b',
    ]

    EMAIL_PATTERN = r'\b([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})\b'

    def __init__(self, config: WorkerConfig = None):
        super().__init__(config or WorkerConfig())

    @property
    def worker_type(self) -> WorkerType:
        return WorkerType.PERSON

    def get_system_prompt(self) -> str:
        return """You are a specialized person/entity extraction AI. Your task is to identify and extract all references to people from text.

Extract:
- Full names (e.g., "John Smith", "Dr. Jane Doe")
- Email addresses associated with people
- Titles and roles (e.g., "CEO", "Director of Sales")
- Relationships (e.g., "John's manager", "the defendant")
- Aliases and nicknames

Output JSON array with: name, email, title, role, organization, context"""

    def get_extraction_prompt(self, text: str) -> str:
        return f"""Extract all people and their information from this text:

TEXT:
{text}

Return a JSON array of objects with fields:
- name: full name of the person
- email: email address if found
- title: professional title if mentioned
- role: their role in the context
- organization: associated organization if mentioned
- relationships: list of relationships to other people
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

        emails = re.findall(self.EMAIL_PATTERN, response)
        for email in emails:
            if not any(e.get('email') == email for e in entities):
                name = self._extract_name_from_email(email)
                entities.append({
                    'name': name,
                    'email': email,
                    'title': None,
                    'role': None,
                    'organization': self._extract_org_from_email(email),
                    'relationships': [],
                    'confidence': 0.7
                })

        return entities

    def _extract_name_from_email(self, email: str) -> str:
        local_part = email.split('@')[0]
        parts = re.split(r'[._-]', local_part)
        if len(parts) >= 2:
            return ' '.join(p.capitalize() for p in parts[:2])
        return local_part.capitalize()

    def _extract_org_from_email(self, email: str) -> str:
        domain = email.split('@')[1]
        org = domain.split('.')[0]
        return org.capitalize()

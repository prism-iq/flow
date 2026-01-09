import re
import json
from typing import List, Dict, Any
from .base import BaseWorker, WorkerType, WorkerConfig


class OrganizationWorker(BaseWorker):
    ORG_SUFFIXES = [
        'Inc', 'LLC', 'Ltd', 'Corp', 'Corporation', 'Company', 'Co',
        'LLP', 'LP', 'PLC', 'GmbH', 'AG', 'SA', 'NV', 'BV',
        'Foundation', 'Institute', 'Association', 'Group', 'Partners',
    ]

    def __init__(self, config: WorkerConfig = None):
        super().__init__(config or WorkerConfig())

    @property
    def worker_type(self) -> WorkerType:
        return WorkerType.ORGANIZATION

    def get_system_prompt(self) -> str:
        return """You are a specialized organization extraction AI. Your task is to identify and extract all references to organizations, companies, and institutions from text.

Extract:
- Company names (e.g., "Apple Inc.", "Goldman Sachs")
- Government agencies (e.g., "FBI", "SEC", "Department of Justice")
- Non-profits and NGOs
- Educational institutions
- Subsidiaries and parent companies
- Industry/sector information

Output JSON array with: name, type, aliases, parent_org, industry, context"""

    def get_extraction_prompt(self, text: str) -> str:
        return f"""Extract all organizations and companies from this text:

TEXT:
{text}

Return a JSON array of objects with fields:
- name: official name of the organization
- type: one of [company, government, nonprofit, educational, other]
- aliases: list of alternative names or abbreviations
- parent_org: parent organization if mentioned
- industry: industry or sector
- location: headquarters or primary location
- context: how the organization is referenced
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

        suffix_pattern = '|'.join(re.escape(s) for s in self.ORG_SUFFIXES)
        pattern = rf'\b([A-Z][A-Za-z\s&]+)\s+({suffix_pattern})\.?\b'

        matches = re.finditer(pattern, response)
        for match in matches:
            org_name = f"{match.group(1).strip()} {match.group(2)}"
            if not any(e.get('name', '').lower() == org_name.lower() for e in entities):
                entities.append({
                    'name': org_name,
                    'type': 'company',
                    'aliases': [],
                    'parent_org': None,
                    'industry': None,
                    'location': None,
                    'context': '',
                    'confidence': 0.8
                })

        return entities

import re
import json
from typing import List, Dict, Any
from decimal import Decimal, InvalidOperation
from .base import BaseWorker, WorkerType, WorkerConfig


class AmountWorker(BaseWorker):
    CURRENCY_SYMBOLS = {
        '$': 'USD', '€': 'EUR', '£': 'GBP', '¥': 'JPY',
        '₹': 'INR', '₽': 'RUB', '₿': 'BTC', 'CHF': 'CHF',
    }

    AMOUNT_PATTERNS = [
        r'[\$€£¥₹₽]\s*(\d{1,3}(?:,\d{3})*(?:\.\d{2})?)\s*(million|billion|thousand|M|B|K)?',
        r'(\d{1,3}(?:,\d{3})*(?:\.\d{2})?)\s*(USD|EUR|GBP|JPY|CHF|INR)',
        r'(\d{1,3}(?:,\d{3})*(?:\.\d{2})?)\s*(million|billion|thousand|M|B|K)?\s*(?:dollars?|euros?|pounds?)',
    ]

    MULTIPLIERS = {
        'thousand': 1000, 'k': 1000, 'K': 1000,
        'million': 1000000, 'm': 1000000, 'M': 1000000,
        'billion': 1000000000, 'b': 1000000000, 'B': 1000000000,
    }

    def __init__(self, config: WorkerConfig = None):
        super().__init__(config or WorkerConfig())

    @property
    def worker_type(self) -> WorkerType:
        return WorkerType.AMOUNT

    def get_system_prompt(self) -> str:
        return """You are a specialized financial amount extraction AI. Your task is to identify and extract all monetary values, quantities, and financial figures from text.

Extract:
- Currency amounts (e.g., "$1.5 million", "€500,000")
- Percentages and rates (e.g., "15%", "3.5% interest")
- Quantities and counts
- Financial metrics (revenue, profit, market cap)
- Ranges (e.g., "$10-15 million")

Output JSON array with: amount, currency, type, context, normalized_value"""

    def get_extraction_prompt(self, text: str) -> str:
        return f"""Extract all monetary amounts and financial figures from this text:

TEXT:
{text}

Return a JSON array of objects with fields:
- amount_string: the exact amount text found
- amount: numeric value
- currency: currency code (USD, EUR, etc.)
- multiplier: thousand/million/billion if applicable
- normalized_value: full numeric value in base units
- amount_type: one of [currency, percentage, quantity, metric]
- context: what the amount refers to
- confidence: 0.0 to 1.0

JSON:"""

    def parse_response(self, response: str) -> List[Dict[str, Any]]:
        entities = []

        try:
            json_match = re.search(r'\[[\s\S]*\]', response)
            if json_match:
                parsed = json.loads(json_match.group())
                if isinstance(parsed, list):
                    for item in parsed:
                        if 'normalized_value' not in item and 'amount' in item:
                            item['normalized_value'] = self._normalize_amount(item)
                        entities.append(item)
        except json.JSONDecodeError:
            pass

        for symbol, currency in self.CURRENCY_SYMBOLS.items():
            escaped_symbol = re.escape(symbol)
            pattern = rf'{escaped_symbol}\s*(\d{{1,3}}(?:,\d{{3}})*(?:\.\d{{1,2}})?)\s*(million|billion|thousand|M|B|K)?'
            matches = re.finditer(pattern, response, re.IGNORECASE)

            for match in matches:
                amount_str = match.group(0)
                if not any(e.get('amount_string') == amount_str for e in entities):
                    amount = self._parse_number(match.group(1))
                    multiplier = match.group(2)

                    if multiplier:
                        mult_value = self.MULTIPLIERS.get(multiplier.lower(), 1)
                        normalized = amount * mult_value
                    else:
                        normalized = amount

                    entities.append({
                        'amount_string': amount_str,
                        'amount': float(amount),
                        'currency': currency,
                        'multiplier': multiplier,
                        'normalized_value': float(normalized),
                        'amount_type': 'currency',
                        'context': '',
                        'confidence': 0.9
                    })

        return entities

    def _parse_number(self, num_str: str) -> Decimal:
        try:
            cleaned = num_str.replace(',', '')
            return Decimal(cleaned)
        except InvalidOperation:
            return Decimal(0)

    def _normalize_amount(self, item: Dict) -> float:
        amount = item.get('amount', 0)
        multiplier = item.get('multiplier', '')

        if isinstance(amount, str):
            amount = float(self._parse_number(amount))

        if multiplier:
            mult_value = self.MULTIPLIERS.get(multiplier.lower(), 1)
            return amount * mult_value

        return amount

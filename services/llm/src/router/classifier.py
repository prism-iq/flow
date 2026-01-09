import re
from typing import List, Tuple, Dict
from enum import Enum
from dataclasses import dataclass


class QueryType(Enum):
    DATE = "date"
    PERSON = "person"
    ORGANIZATION = "org"
    AMOUNT = "amount"
    GENERAL = "general"
    MULTI = "multi"


@dataclass
class ClassificationResult:
    primary_type: QueryType
    all_types: List[Tuple[QueryType, float]]
    confidence: float
    signals: Dict[str, List[str]]


class QueryClassifier:
    DATE_KEYWORDS = [
        'when', 'date', 'time', 'day', 'month', 'year', 'deadline',
        'schedule', 'calendar', 'meeting', 'appointment', 'january',
        'february', 'march', 'april', 'may', 'june', 'july', 'august',
        'september', 'october', 'november', 'december', 'monday',
        'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday',
        'yesterday', 'today', 'tomorrow', 'week', 'quarter', 'fiscal',
    ]

    PERSON_KEYWORDS = [
        'who', 'person', 'people', 'name', 'employee', 'manager',
        'director', 'ceo', 'cfo', 'president', 'executive', 'staff',
        'team', 'contact', 'author', 'sender', 'recipient', 'mr',
        'mrs', 'ms', 'dr', 'prof', 'sir', 'sent by', 'from', 'to',
    ]

    ORG_KEYWORDS = [
        'company', 'organization', 'corporation', 'firm', 'business',
        'enterprise', 'inc', 'llc', 'ltd', 'corp', 'agency', 'department',
        'institution', 'foundation', 'association', 'group', 'bank',
        'fund', 'partner', 'client', 'vendor', 'supplier', 'subsidiary',
    ]

    AMOUNT_KEYWORDS = [
        'how much', 'amount', 'price', 'cost', 'value', 'total',
        'sum', 'payment', 'invoice', 'dollar', 'euro', 'pound',
        'million', 'billion', 'thousand', 'revenue', 'profit',
        'expense', 'budget', 'fee', 'rate', 'percentage', '%',
        '$', '€', '£', 'usd', 'eur', 'gbp',
    ]

    DATE_PATTERNS = [
        r'\d{1,2}[/-]\d{1,2}[/-]\d{2,4}',
        r'\d{4}[/-]\d{1,2}[/-]\d{1,2}',
        r'(jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)\w*\s+\d{1,2}',
        r'\d{1,2}\s+(jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)',
    ]

    AMOUNT_PATTERNS = [
        r'[\$€£¥]\s*\d+',
        r'\d+\s*(usd|eur|gbp|dollars?|euros?|pounds?)',
        r'\d+\s*(million|billion|thousand|[MBK])\b',
        r'\d+(\.\d{2})?\s*%',
    ]

    EMAIL_PATTERN = r'[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}'

    def __init__(self):
        self._compile_patterns()

    def _compile_patterns(self):
        self.date_re = [re.compile(p, re.IGNORECASE) for p in self.DATE_PATTERNS]
        self.amount_re = [re.compile(p, re.IGNORECASE) for p in self.AMOUNT_PATTERNS]
        self.email_re = re.compile(self.EMAIL_PATTERN)

    def classify(self, query: str) -> ClassificationResult:
        query_lower = query.lower()
        signals = {
            'date': [],
            'person': [],
            'org': [],
            'amount': [],
        }
        scores = {
            QueryType.DATE: 0.0,
            QueryType.PERSON: 0.0,
            QueryType.ORGANIZATION: 0.0,
            QueryType.AMOUNT: 0.0,
            QueryType.GENERAL: 0.1,
        }

        for kw in self.DATE_KEYWORDS:
            if kw in query_lower:
                scores[QueryType.DATE] += 0.15
                signals['date'].append(f"keyword:{kw}")

        for kw in self.PERSON_KEYWORDS:
            if kw in query_lower:
                scores[QueryType.PERSON] += 0.15
                signals['person'].append(f"keyword:{kw}")

        for kw in self.ORG_KEYWORDS:
            if kw in query_lower:
                scores[QueryType.ORGANIZATION] += 0.15
                signals['org'].append(f"keyword:{kw}")

        for kw in self.AMOUNT_KEYWORDS:
            if kw in query_lower:
                scores[QueryType.AMOUNT] += 0.15
                signals['amount'].append(f"keyword:{kw}")

        for pattern in self.date_re:
            matches = pattern.findall(query)
            if matches:
                scores[QueryType.DATE] += 0.3 * len(matches)
                signals['date'].append(f"pattern:{len(matches)} matches")

        for pattern in self.amount_re:
            matches = pattern.findall(query)
            if matches:
                scores[QueryType.AMOUNT] += 0.3 * len(matches)
                signals['amount'].append(f"pattern:{len(matches)} matches")

        emails = self.email_re.findall(query)
        if emails:
            scores[QueryType.PERSON] += 0.25 * len(emails)
            signals['person'].append(f"email:{len(emails)} found")

        for score_type in scores:
            scores[score_type] = min(scores[score_type], 1.0)

        sorted_types = sorted(scores.items(), key=lambda x: x[1], reverse=True)
        primary_type = sorted_types[0][0]
        primary_score = sorted_types[0][1]

        high_scores = [(t, s) for t, s in sorted_types if s >= 0.3]
        if len(high_scores) > 1:
            primary_type = QueryType.MULTI

        return ClassificationResult(
            primary_type=primary_type,
            all_types=sorted_types,
            confidence=primary_score,
            signals=signals
        )

    def get_relevant_types(self, query: str, threshold: float = 0.2) -> List[QueryType]:
        result = self.classify(query)
        return [t for t, s in result.all_types if s >= threshold and t != QueryType.GENERAL]

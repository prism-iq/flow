from typing import List, Dict, Any, Tuple
from dataclasses import dataclass
import re
from difflib import SequenceMatcher


@dataclass
class MergeCandidate:
    index_a: int
    index_b: int
    similarity: float
    merge_type: str


class EntityMerger:
    def __init__(self, similarity_threshold: float = 0.85):
        self.similarity_threshold = similarity_threshold

    def find_duplicates(self, entities: List[Dict[str, Any]]) -> List[MergeCandidate]:
        candidates = []

        for i in range(len(entities)):
            for j in range(i + 1, len(entities)):
                similarity, merge_type = self._calculate_similarity(
                    entities[i], entities[j]
                )

                if similarity >= self.similarity_threshold:
                    candidates.append(MergeCandidate(
                        index_a=i,
                        index_b=j,
                        similarity=similarity,
                        merge_type=merge_type
                    ))

        return candidates

    def merge_entities(self, entities: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        if not entities:
            return []

        candidates = self.find_duplicates(entities)

        merged_indices = set()
        merged_entities = []

        for candidate in sorted(candidates, key=lambda x: x.similarity, reverse=True):
            if candidate.index_a in merged_indices or candidate.index_b in merged_indices:
                continue

            entity_a = entities[candidate.index_a]
            entity_b = entities[candidate.index_b]

            merged = self._merge_two_entities(entity_a, entity_b)
            merged['_merged_from'] = [candidate.index_a, candidate.index_b]
            merged['_merge_confidence'] = candidate.similarity

            merged_entities.append(merged)
            merged_indices.add(candidate.index_a)
            merged_indices.add(candidate.index_b)

        for i, entity in enumerate(entities):
            if i not in merged_indices:
                merged_entities.append(entity)

        return merged_entities

    def _calculate_similarity(
        self,
        entity_a: Dict[str, Any],
        entity_b: Dict[str, Any]
    ) -> Tuple[float, str]:
        type_a = entity_a.get('_source_worker', '')
        type_b = entity_b.get('_source_worker', '')

        if type_a != type_b:
            return 0.0, 'different_type'

        if type_a == 'person':
            return self._person_similarity(entity_a, entity_b)
        elif type_a == 'org':
            return self._org_similarity(entity_a, entity_b)
        elif type_a == 'date':
            return self._date_similarity(entity_a, entity_b)
        elif type_a == 'amount':
            return self._amount_similarity(entity_a, entity_b)
        else:
            return self._generic_similarity(entity_a, entity_b)

    def _person_similarity(
        self,
        a: Dict[str, Any],
        b: Dict[str, Any]
    ) -> Tuple[float, str]:
        email_a = a.get('email', '').lower()
        email_b = b.get('email', '').lower()

        if email_a and email_b and email_a == email_b:
            return 1.0, 'email_match'

        name_a = a.get('name', '').lower()
        name_b = b.get('name', '').lower()

        if name_a and name_b:
            similarity = SequenceMatcher(None, name_a, name_b).ratio()
            if similarity > 0.9:
                return similarity, 'name_match'

            parts_a = set(name_a.split())
            parts_b = set(name_b.split())
            if parts_a & parts_b:
                overlap = len(parts_a & parts_b) / max(len(parts_a), len(parts_b))
                return overlap, 'name_overlap'

        return 0.0, 'no_match'

    def _org_similarity(
        self,
        a: Dict[str, Any],
        b: Dict[str, Any]
    ) -> Tuple[float, str]:
        name_a = self._normalize_org_name(a.get('name', ''))
        name_b = self._normalize_org_name(b.get('name', ''))

        if name_a == name_b:
            return 1.0, 'exact_match'

        similarity = SequenceMatcher(None, name_a, name_b).ratio()
        return similarity, 'fuzzy_match'

    def _normalize_org_name(self, name: str) -> str:
        name = name.lower()
        suffixes = ['inc', 'llc', 'ltd', 'corp', 'corporation', 'company', 'co']
        for suffix in suffixes:
            name = re.sub(rf'\b{suffix}\.?\b', '', name)
        return ' '.join(name.split())

    def _date_similarity(
        self,
        a: Dict[str, Any],
        b: Dict[str, Any]
    ) -> Tuple[float, str]:
        norm_a = a.get('normalized_date')
        norm_b = b.get('normalized_date')

        if norm_a and norm_b and norm_a == norm_b:
            return 1.0, 'date_match'

        return 0.0, 'no_match'

    def _amount_similarity(
        self,
        a: Dict[str, Any],
        b: Dict[str, Any]
    ) -> Tuple[float, str]:
        val_a = a.get('normalized_value', 0)
        val_b = b.get('normalized_value', 0)

        if val_a and val_b:
            if val_a == val_b:
                return 1.0, 'exact_amount'

            ratio = min(val_a, val_b) / max(val_a, val_b) if max(val_a, val_b) > 0 else 0
            if ratio > 0.99:
                return ratio, 'similar_amount'

        return 0.0, 'no_match'

    def _generic_similarity(
        self,
        a: Dict[str, Any],
        b: Dict[str, Any]
    ) -> Tuple[float, str]:
        str_a = str(a)
        str_b = str(b)
        similarity = SequenceMatcher(None, str_a, str_b).ratio()
        return similarity, 'generic'

    def _merge_two_entities(
        self,
        a: Dict[str, Any],
        b: Dict[str, Any]
    ) -> Dict[str, Any]:
        merged = {}

        all_keys = set(a.keys()) | set(b.keys())

        for key in all_keys:
            val_a = a.get(key)
            val_b = b.get(key)

            if val_a is None:
                merged[key] = val_b
            elif val_b is None:
                merged[key] = val_a
            elif isinstance(val_a, list) and isinstance(val_b, list):
                merged[key] = list(set(val_a + val_b))
            elif isinstance(val_a, (int, float)) and isinstance(val_b, (int, float)):
                conf_a = a.get('confidence', 0.5)
                conf_b = b.get('confidence', 0.5)
                if conf_a >= conf_b:
                    merged[key] = val_a
                else:
                    merged[key] = val_b
            elif len(str(val_a)) >= len(str(val_b)):
                merged[key] = val_a
            else:
                merged[key] = val_b

        conf_a = a.get('confidence', 0.5)
        conf_b = b.get('confidence', 0.5)
        merged['confidence'] = max(conf_a, conf_b)

        return merged

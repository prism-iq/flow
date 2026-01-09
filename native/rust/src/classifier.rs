use serde::{Deserialize, Serialize};
use regex::Regex;
use std::collections::HashMap;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum QueryType {
    Date,
    Person,
    Organization,
    Amount,
    General,
    Multi,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClassificationResult {
    pub primary_type: QueryType,
    pub scores: HashMap<String, f64>,
    pub confidence: f64,
    pub signals: Vec<String>,
}

pub struct FastClassifier {
    date_keywords: Vec<&'static str>,
    person_keywords: Vec<&'static str>,
    org_keywords: Vec<&'static str>,
    amount_keywords: Vec<&'static str>,
    date_patterns: Vec<Regex>,
    amount_patterns: Vec<Regex>,
    email_pattern: Regex,
}

impl FastClassifier {
    pub fn new() -> Self {
        Self {
            date_keywords: vec![
                "when", "date", "time", "day", "month", "year", "deadline",
                "schedule", "calendar", "meeting", "january", "february",
                "march", "april", "may", "june", "july", "august",
                "september", "october", "november", "december", "monday",
                "tuesday", "wednesday", "thursday", "friday", "saturday",
                "sunday", "yesterday", "today", "tomorrow", "week", "quarter",
            ],
            person_keywords: vec![
                "who", "person", "people", "name", "employee", "manager",
                "director", "ceo", "cfo", "president", "executive", "staff",
                "team", "contact", "author", "sender", "recipient", "mr",
                "mrs", "ms", "dr", "sent by", "from", "to",
            ],
            org_keywords: vec![
                "company", "organization", "corporation", "firm", "business",
                "enterprise", "inc", "llc", "ltd", "corp", "agency",
                "department", "institution", "foundation", "bank", "fund",
            ],
            amount_keywords: vec![
                "how much", "amount", "price", "cost", "value", "total",
                "sum", "payment", "invoice", "dollar", "euro", "million",
                "billion", "thousand", "revenue", "profit", "expense",
            ],
            date_patterns: vec![
                Regex::new(r"\d{1,2}[/-]\d{1,2}[/-]\d{2,4}").unwrap(),
                Regex::new(r"\d{4}[/-]\d{1,2}[/-]\d{1,2}").unwrap(),
                Regex::new(r"(?i)(jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)\w*\s+\d{1,2}").unwrap(),
            ],
            amount_patterns: vec![
                Regex::new(r"[\$€£¥]\s*[\d,]+").unwrap(),
                Regex::new(r"(?i)\d+\s*(usd|eur|gbp|dollars?|euros?)").unwrap(),
                Regex::new(r"(?i)\d+\s*(million|billion|thousand|[MBK])\b").unwrap(),
                Regex::new(r"\d+(\.\d{2})?\s*%").unwrap(),
            ],
            email_pattern: Regex::new(r"[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}").unwrap(),
        }
    }

    pub fn classify(&self, query: &str) -> ClassificationResult {
        let query_lower = query.to_lowercase();
        let mut scores: HashMap<String, f64> = HashMap::new();
        let mut signals = Vec::new();

        scores.insert("date".to_string(), 0.0);
        scores.insert("person".to_string(), 0.0);
        scores.insert("org".to_string(), 0.0);
        scores.insert("amount".to_string(), 0.0);
        scores.insert("general".to_string(), 0.1);

        // Keyword scoring
        for kw in &self.date_keywords {
            if query_lower.contains(kw) {
                *scores.get_mut("date").unwrap() += 0.15;
                signals.push(format!("date_kw:{}", kw));
            }
        }

        for kw in &self.person_keywords {
            if query_lower.contains(kw) {
                *scores.get_mut("person").unwrap() += 0.15;
                signals.push(format!("person_kw:{}", kw));
            }
        }

        for kw in &self.org_keywords {
            if query_lower.contains(kw) {
                *scores.get_mut("org").unwrap() += 0.15;
                signals.push(format!("org_kw:{}", kw));
            }
        }

        for kw in &self.amount_keywords {
            if query_lower.contains(kw) {
                *scores.get_mut("amount").unwrap() += 0.15;
                signals.push(format!("amount_kw:{}", kw));
            }
        }

        // Pattern scoring
        for pattern in &self.date_patterns {
            let count = pattern.find_iter(query).count();
            if count > 0 {
                *scores.get_mut("date").unwrap() += 0.3 * count as f64;
                signals.push(format!("date_pattern:{}", count));
            }
        }

        for pattern in &self.amount_patterns {
            let count = pattern.find_iter(query).count();
            if count > 0 {
                *scores.get_mut("amount").unwrap() += 0.3 * count as f64;
                signals.push(format!("amount_pattern:{}", count));
            }
        }

        let email_count = self.email_pattern.find_iter(query).count();
        if email_count > 0 {
            *scores.get_mut("person").unwrap() += 0.25 * email_count as f64;
            signals.push(format!("email:{}", email_count));
        }

        // Cap scores at 1.0
        for score in scores.values_mut() {
            *score = score.min(1.0);
        }

        // Determine primary type
        let mut max_type = "general";
        let mut max_score = 0.0;
        let mut high_count = 0;

        for (type_name, &score) in &scores {
            if score > max_score {
                max_score = score;
                max_type = type_name;
            }
            if score >= 0.3 {
                high_count += 1;
            }
        }

        let primary_type = if high_count > 1 {
            QueryType::Multi
        } else {
            match max_type {
                "date" => QueryType::Date,
                "person" => QueryType::Person,
                "org" => QueryType::Organization,
                "amount" => QueryType::Amount,
                _ => QueryType::General,
            }
        };

        ClassificationResult {
            primary_type,
            scores,
            confidence: max_score,
            signals,
        }
    }

    pub fn get_relevant_types(&self, query: &str, threshold: f64) -> Vec<QueryType> {
        let result = self.classify(query);
        let mut types = Vec::new();

        for (type_name, &score) in &result.scores {
            if score >= threshold && type_name != "general" {
                types.push(match type_name.as_str() {
                    "date" => QueryType::Date,
                    "person" => QueryType::Person,
                    "org" => QueryType::Organization,
                    "amount" => QueryType::Amount,
                    _ => continue,
                });
            }
        }

        types
    }
}

impl Default for FastClassifier {
    fn default() -> Self {
        Self::new()
    }
}

// FFI exports
#[no_mangle]
pub extern "C" fn flow_classifier_create() -> *mut FastClassifier {
    Box::into_raw(Box::new(FastClassifier::new()))
}

#[no_mangle]
pub extern "C" fn flow_classifier_destroy(ptr: *mut FastClassifier) {
    if !ptr.is_null() {
        unsafe { drop(Box::from_raw(ptr)) }
    }
}

#[no_mangle]
pub extern "C" fn flow_classifier_classify(
    ptr: *mut FastClassifier,
    query: *const std::os::raw::c_char,
    result_json: *mut *mut std::os::raw::c_char,
) -> i32 {
    if ptr.is_null() || query.is_null() || result_json.is_null() {
        return -1;
    }

    unsafe {
        let classifier = &*ptr;
        let query_str = match std::ffi::CStr::from_ptr(query).to_str() {
            Ok(s) => s,
            Err(_) => return -2,
        };

        let result = classifier.classify(query_str);

        match serde_json::to_string(&result) {
            Ok(json) => {
                match std::ffi::CString::new(json) {
                    Ok(cstr) => {
                        *result_json = cstr.into_raw();
                        0
                    }
                    Err(_) => -3,
                }
            }
            Err(_) => -4,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_date_classification() {
        let classifier = FastClassifier::new();
        let result = classifier.classify("When was the meeting on 2024-01-15?");
        assert_eq!(result.primary_type, QueryType::Date);
    }

    #[test]
    fn test_amount_classification() {
        let classifier = FastClassifier::new();
        let result = classifier.classify("The total cost was $5 million");
        assert_eq!(result.primary_type, QueryType::Amount);
    }

    #[test]
    fn test_person_classification() {
        let classifier = FastClassifier::new();
        let result = classifier.classify("Who sent the email to john@example.com?");
        assert_eq!(result.primary_type, QueryType::Person);
    }

    #[test]
    fn test_multi_classification() {
        let classifier = FastClassifier::new();
        let result = classifier.classify("Who paid $1 million on January 15?");
        assert_eq!(result.primary_type, QueryType::Multi);
    }
}

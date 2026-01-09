use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum TokenizerError {
    #[error("Invalid input: {0}")]
    InvalidInput(String),
    #[error("Token not found: {0}")]
    TokenNotFound(String),
    #[error("Encoding error: {0}")]
    EncodingError(String),
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Token {
    pub id: u32,
    pub text: String,
    pub start: usize,
    pub end: usize,
}

#[derive(Debug, Clone)]
pub struct TokenizerConfig {
    pub vocab_size: usize,
    pub max_length: usize,
    pub pad_token: String,
    pub unk_token: String,
    pub bos_token: String,
    pub eos_token: String,
}

impl Default for TokenizerConfig {
    fn default() -> Self {
        Self {
            vocab_size: 32000,
            max_length: 4096,
            pad_token: "<pad>".to_string(),
            unk_token: "<unk>".to_string(),
            bos_token: "<s>".to_string(),
            eos_token: "</s>".to_string(),
        }
    }
}

pub struct Tokenizer {
    config: TokenizerConfig,
    vocab: HashMap<String, u32>,
    reverse_vocab: HashMap<u32, String>,
    special_tokens: Vec<String>,
}

impl Tokenizer {
    pub fn new(config: TokenizerConfig) -> Self {
        let mut vocab = HashMap::new();
        let mut reverse_vocab = HashMap::new();

        let special_tokens = vec![
            config.pad_token.clone(),
            config.unk_token.clone(),
            config.bos_token.clone(),
            config.eos_token.clone(),
        ];

        for (i, token) in special_tokens.iter().enumerate() {
            vocab.insert(token.clone(), i as u32);
            reverse_vocab.insert(i as u32, token.clone());
        }

        Self {
            config,
            vocab,
            reverse_vocab,
            special_tokens,
        }
    }

    pub fn encode(&self, text: &str) -> Result<Vec<Token>, TokenizerError> {
        if text.is_empty() {
            return Ok(vec![]);
        }

        let mut tokens = Vec::new();
        let mut current_pos = 0;

        for word in text.split_whitespace() {
            let start = text[current_pos..].find(word).unwrap_or(0) + current_pos;
            let end = start + word.len();

            let id = self.vocab.get(word).copied().unwrap_or(1); // UNK token

            tokens.push(Token {
                id,
                text: word.to_string(),
                start,
                end,
            });

            current_pos = end;
        }

        if tokens.len() > self.config.max_length {
            tokens.truncate(self.config.max_length);
        }

        Ok(tokens)
    }

    pub fn decode(&self, token_ids: &[u32]) -> Result<String, TokenizerError> {
        let mut result = String::new();

        for &id in token_ids {
            if let Some(text) = self.reverse_vocab.get(&id) {
                if !result.is_empty() {
                    result.push(' ');
                }
                result.push_str(text);
            }
        }

        Ok(result)
    }

    pub fn encode_batch(&self, texts: &[&str]) -> Result<Vec<Vec<Token>>, TokenizerError> {
        use rayon::prelude::*;

        texts
            .par_iter()
            .map(|text| self.encode(text))
            .collect()
    }

    pub fn add_token(&mut self, token: &str) -> u32 {
        if let Some(&id) = self.vocab.get(token) {
            return id;
        }

        let id = self.vocab.len() as u32;
        self.vocab.insert(token.to_string(), id);
        self.reverse_vocab.insert(id, token.to_string());
        id
    }

    pub fn vocab_size(&self) -> usize {
        self.vocab.len()
    }

    pub fn count_tokens(&self, text: &str) -> usize {
        text.split_whitespace().count()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_tokenize_basic() {
        let tokenizer = Tokenizer::new(TokenizerConfig::default());
        let tokens = tokenizer.encode("hello world").unwrap();
        assert_eq!(tokens.len(), 2);
    }

    #[test]
    fn test_empty_input() {
        let tokenizer = Tokenizer::new(TokenizerConfig::default());
        let tokens = tokenizer.encode("").unwrap();
        assert!(tokens.is_empty());
    }
}

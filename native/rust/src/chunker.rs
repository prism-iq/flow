use serde::{Deserialize, Serialize};
use rayon::prelude::*;
use unicode_segmentation::UnicodeSegmentation;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Chunk {
    pub id: usize,
    pub content: String,
    pub start_offset: usize,
    pub end_offset: usize,
    pub token_count: usize,
    pub metadata: ChunkMetadata,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct ChunkMetadata {
    pub source: Option<String>,
    pub section: Option<String>,
    pub page: Option<usize>,
}

#[derive(Debug, Clone)]
pub struct ChunkerConfig {
    pub chunk_size: usize,
    pub chunk_overlap: usize,
    pub min_chunk_size: usize,
    pub separators: Vec<String>,
    pub preserve_sentences: bool,
}

impl Default for ChunkerConfig {
    fn default() -> Self {
        Self {
            chunk_size: 512,
            chunk_overlap: 64,
            min_chunk_size: 100,
            separators: vec![
                "\n\n".to_string(),
                "\n".to_string(),
                ". ".to_string(),
                "! ".to_string(),
                "? ".to_string(),
                " ".to_string(),
            ],
            preserve_sentences: true,
        }
    }
}

pub struct TextChunker {
    config: ChunkerConfig,
}

impl TextChunker {
    pub fn new(config: ChunkerConfig) -> Self {
        Self { config }
    }

    pub fn chunk(&self, text: &str) -> Vec<Chunk> {
        if text.is_empty() {
            return vec![];
        }

        let mut chunks = Vec::new();
        let mut current_start = 0;
        let mut chunk_id = 0;

        while current_start < text.len() {
            let end = self.find_chunk_end(text, current_start);
            let chunk_text = &text[current_start..end];

            if !chunk_text.trim().is_empty() {
                chunks.push(Chunk {
                    id: chunk_id,
                    content: chunk_text.to_string(),
                    start_offset: current_start,
                    end_offset: end,
                    token_count: self.estimate_tokens(chunk_text),
                    metadata: ChunkMetadata::default(),
                });
                chunk_id += 1;
            }

            if end >= text.len() {
                break;
            }

            current_start = if self.config.chunk_overlap > 0 && end > self.config.chunk_overlap {
                end - self.config.chunk_overlap
            } else {
                end
            };
        }

        chunks
    }

    fn find_chunk_end(&self, text: &str, start: usize) -> usize {
        let remaining = &text[start..];
        let max_end = (start + self.config.chunk_size).min(text.len());

        if max_end >= text.len() {
            return text.len();
        }

        let search_region = &text[start..max_end];

        for sep in &self.config.separators {
            if let Some(pos) = search_region.rfind(sep) {
                let end = start + pos + sep.len();
                if end - start >= self.config.min_chunk_size {
                    return end;
                }
            }
        }

        max_end
    }

    pub fn chunk_batch(&self, texts: &[&str]) -> Vec<Vec<Chunk>> {
        texts
            .par_iter()
            .map(|text| self.chunk(text))
            .collect()
    }

    pub fn chunk_with_metadata(&self, text: &str, metadata: ChunkMetadata) -> Vec<Chunk> {
        self.chunk(text)
            .into_iter()
            .map(|mut chunk| {
                chunk.metadata = metadata.clone();
                chunk
            })
            .collect()
    }

    fn estimate_tokens(&self, text: &str) -> usize {
        text.unicode_words().count()
    }

    pub fn merge_small_chunks(&self, chunks: Vec<Chunk>) -> Vec<Chunk> {
        let mut merged = Vec::new();
        let mut current: Option<Chunk> = None;

        for chunk in chunks {
            match &mut current {
                Some(c) if c.token_count + chunk.token_count < self.config.chunk_size => {
                    c.content.push_str(&chunk.content);
                    c.end_offset = chunk.end_offset;
                    c.token_count += chunk.token_count;
                }
                _ => {
                    if let Some(c) = current.take() {
                        merged.push(c);
                    }
                    current = Some(chunk);
                }
            }
        }

        if let Some(c) = current {
            merged.push(c);
        }

        merged
    }

    pub fn split_by_sentences(&self, text: &str) -> Vec<String> {
        let mut sentences = Vec::new();
        let mut current = String::new();

        for c in text.chars() {
            current.push(c);
            if c == '.' || c == '!' || c == '?' {
                if !current.trim().is_empty() {
                    sentences.push(current.trim().to_string());
                }
                current = String::new();
            }
        }

        if !current.trim().is_empty() {
            sentences.push(current.trim().to_string());
        }

        sentences
    }
}

impl Default for TextChunker {
    fn default() -> Self {
        Self::new(ChunkerConfig::default())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_basic_chunking() {
        let chunker = TextChunker::default();
        let text = "Hello world. This is a test. Another sentence here.";
        let chunks = chunker.chunk(text);
        assert!(!chunks.is_empty());
    }

    #[test]
    fn test_empty_input() {
        let chunker = TextChunker::default();
        let chunks = chunker.chunk("");
        assert!(chunks.is_empty());
    }

    #[test]
    fn test_sentence_split() {
        let chunker = TextChunker::default();
        let sentences = chunker.split_by_sentences("Hello. World! How are you?");
        assert_eq!(sentences.len(), 3);
    }
}

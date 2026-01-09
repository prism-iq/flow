use serde::{Deserialize, Serialize};
use thiserror::Error;
use regex::Regex;

#[derive(Error, Debug)]
pub enum ParseError {
    #[error("Invalid syntax at position {0}")]
    InvalidSyntax(usize),
    #[error("Unexpected token: {0}")]
    UnexpectedToken(String),
    #[error("Missing delimiter: {0}")]
    MissingDelimiter(char),
    #[error("Regex error: {0}")]
    RegexError(#[from] regex::Error),
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum NodeType {
    Text,
    CodeBlock,
    InlineCode,
    Link,
    Image,
    Header,
    List,
    Quote,
    Bold,
    Italic,
    Paragraph,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AstNode {
    pub node_type: NodeType,
    pub content: String,
    pub children: Vec<AstNode>,
    pub metadata: Option<NodeMetadata>,
    pub span: Span,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NodeMetadata {
    pub language: Option<String>,
    pub url: Option<String>,
    pub level: Option<u8>,
    pub alt_text: Option<String>,
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
pub struct Span {
    pub start: usize,
    pub end: usize,
}

pub struct Parser {
    code_block_re: Regex,
    inline_code_re: Regex,
    link_re: Regex,
    image_re: Regex,
    header_re: Regex,
    bold_re: Regex,
    italic_re: Regex,
}

impl Parser {
    pub fn new() -> Result<Self, ParseError> {
        Ok(Self {
            code_block_re: Regex::new(r"```(\w*)\n([\s\S]*?)```")?,
            inline_code_re: Regex::new(r"`([^`]+)`")?,
            link_re: Regex::new(r"\[([^\]]+)\]\(([^\)]+)\)")?,
            image_re: Regex::new(r"!\[([^\]]*)\]\(([^\)]+)\)")?,
            header_re: Regex::new(r"^(#{1,6})\s+(.+)$")?,
            bold_re: Regex::new(r"\*\*([^*]+)\*\*")?,
            italic_re: Regex::new(r"\*([^*]+)\*")?,
        })
    }

    pub fn parse(&self, input: &str) -> Result<Vec<AstNode>, ParseError> {
        let mut nodes = Vec::new();
        let mut remaining = input;
        let mut offset = 0;

        while !remaining.is_empty() {
            if let Some(node) = self.try_parse_code_block(remaining, offset)? {
                let len = node.span.end - node.span.start;
                remaining = &remaining[len..];
                offset += len;
                nodes.push(node);
                continue;
            }

            if let Some(node) = self.try_parse_header(remaining, offset)? {
                let line_end = remaining.find('\n').unwrap_or(remaining.len());
                remaining = &remaining[line_end..].trim_start_matches('\n');
                offset += line_end + 1;
                nodes.push(node);
                continue;
            }

            let line_end = remaining.find('\n').unwrap_or(remaining.len());
            let line = &remaining[..line_end];

            if !line.trim().is_empty() {
                let inline_nodes = self.parse_inline(line, offset)?;
                nodes.push(AstNode {
                    node_type: NodeType::Paragraph,
                    content: String::new(),
                    children: inline_nodes,
                    metadata: None,
                    span: Span {
                        start: offset,
                        end: offset + line_end,
                    },
                });
            }

            remaining = &remaining[line_end..].trim_start_matches('\n');
            offset += line_end + 1;
        }

        Ok(nodes)
    }

    fn try_parse_code_block(&self, input: &str, offset: usize) -> Result<Option<AstNode>, ParseError> {
        if let Some(caps) = self.code_block_re.captures(input) {
            if caps.get(0).map(|m| m.start()) == Some(0) {
                let full_match = caps.get(0).unwrap();
                let language = caps.get(1).map(|m| m.as_str().to_string());
                let code = caps.get(2).map(|m| m.as_str().to_string()).unwrap_or_default();

                return Ok(Some(AstNode {
                    node_type: NodeType::CodeBlock,
                    content: code,
                    children: vec![],
                    metadata: Some(NodeMetadata {
                        language,
                        url: None,
                        level: None,
                        alt_text: None,
                    }),
                    span: Span {
                        start: offset,
                        end: offset + full_match.end(),
                    },
                }));
            }
        }
        Ok(None)
    }

    fn try_parse_header(&self, input: &str, offset: usize) -> Result<Option<AstNode>, ParseError> {
        let line = input.lines().next().unwrap_or("");

        if let Some(caps) = self.header_re.captures(line) {
            let hashes = caps.get(1).unwrap().as_str();
            let content = caps.get(2).unwrap().as_str();

            return Ok(Some(AstNode {
                node_type: NodeType::Header,
                content: content.to_string(),
                children: vec![],
                metadata: Some(NodeMetadata {
                    language: None,
                    url: None,
                    level: Some(hashes.len() as u8),
                    alt_text: None,
                }),
                span: Span {
                    start: offset,
                    end: offset + line.len(),
                },
            }));
        }
        Ok(None)
    }

    fn parse_inline(&self, input: &str, offset: usize) -> Result<Vec<AstNode>, ParseError> {
        let mut nodes = Vec::new();
        let mut last_end = 0;

        for caps in self.link_re.captures_iter(input) {
            let m = caps.get(0).unwrap();
            let start = m.start();

            if start > last_end {
                nodes.push(AstNode {
                    node_type: NodeType::Text,
                    content: input[last_end..start].to_string(),
                    children: vec![],
                    metadata: None,
                    span: Span {
                        start: offset + last_end,
                        end: offset + start,
                    },
                });
            }

            let text = caps.get(1).unwrap().as_str();
            let url = caps.get(2).unwrap().as_str();

            nodes.push(AstNode {
                node_type: NodeType::Link,
                content: text.to_string(),
                children: vec![],
                metadata: Some(NodeMetadata {
                    language: None,
                    url: Some(url.to_string()),
                    level: None,
                    alt_text: None,
                }),
                span: Span {
                    start: offset + start,
                    end: offset + m.end(),
                },
            });

            last_end = m.end();
        }

        if last_end < input.len() {
            nodes.push(AstNode {
                node_type: NodeType::Text,
                content: input[last_end..].to_string(),
                children: vec![],
                metadata: None,
                span: Span {
                    start: offset + last_end,
                    end: offset + input.len(),
                },
            });
        }

        if nodes.is_empty() {
            nodes.push(AstNode {
                node_type: NodeType::Text,
                content: input.to_string(),
                children: vec![],
                metadata: None,
                span: Span {
                    start: offset,
                    end: offset + input.len(),
                },
            });
        }

        Ok(nodes)
    }

    pub fn to_plain_text(&self, nodes: &[AstNode]) -> String {
        nodes
            .iter()
            .map(|node| self.node_to_text(node))
            .collect::<Vec<_>>()
            .join("\n")
    }

    fn node_to_text(&self, node: &AstNode) -> String {
        if !node.children.is_empty() {
            node.children
                .iter()
                .map(|n| self.node_to_text(n))
                .collect::<Vec<_>>()
                .join("")
        } else {
            node.content.clone()
        }
    }
}

impl Default for Parser {
    fn default() -> Self {
        Self::new().expect("Failed to create parser")
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_header() {
        let parser = Parser::new().unwrap();
        let nodes = parser.parse("# Hello World").unwrap();
        assert_eq!(nodes.len(), 1);
        assert_eq!(nodes[0].node_type, NodeType::Header);
    }

    #[test]
    fn test_parse_code_block() {
        let parser = Parser::new().unwrap();
        let input = "```rust\nfn main() {}\n```";
        let nodes = parser.parse(input).unwrap();
        assert_eq!(nodes.len(), 1);
        assert_eq!(nodes[0].node_type, NodeType::CodeBlock);
    }
}

//go:build !cgo

package native

import (
	"regexp"
	"strings"
	"unicode"
)

type EntityType int

const (
	EntityTypeDate         EntityType = 0
	EntityTypePerson       EntityType = 1
	EntityTypeOrganization EntityType = 2
	EntityTypeAmount       EntityType = 3
	EntityTypeEmail        EntityType = 4
	EntityTypeUnknown      EntityType = 99
)

type Entity struct {
	Value      string
	Type       EntityType
	Start      int
	End        int
	Confidence float32
}

type Match struct {
	Start      int
	End        int
	PatternID  int
	Confidence float32
}

type Token struct {
	Text  string
	Type  int
	Start int
	End   int
}

type PatternMatcher struct {
	patterns []struct {
		text       string
		id         int
		confidence float32
	}
}

func NewPatternMatcher() *PatternMatcher {
	return &PatternMatcher{}
}

func (pm *PatternMatcher) AddPattern(pattern string, id int, confidence float32) {
	pm.patterns = append(pm.patterns, struct {
		text       string
		id         int
		confidence float32
	}{pattern, id, confidence})
}

func (pm *PatternMatcher) FindAll(text string) []Match {
	var matches []Match
	textLower := strings.ToLower(text)

	for _, p := range pm.patterns {
		patternLower := strings.ToLower(p.text)
		idx := 0
		for {
			pos := strings.Index(textLower[idx:], patternLower)
			if pos == -1 {
				break
			}
			actualPos := idx + pos
			matches = append(matches, Match{
				Start:      actualPos,
				End:        actualPos + len(p.text),
				PatternID:  p.id,
				Confidence: p.confidence,
			})
			idx = actualPos + 1
		}
	}

	return matches
}

func (pm *PatternMatcher) Close() {}

type AhoCorasick struct {
	patterns []struct {
		text string
		id   int
	}
}

func NewAhoCorasick() *AhoCorasick {
	return &AhoCorasick{}
}

func (ac *AhoCorasick) AddPattern(pattern string, id int) {
	ac.patterns = append(ac.patterns, struct {
		text string
		id   int
	}{pattern, id})
}

func (ac *AhoCorasick) Build() {}

func (ac *AhoCorasick) Search(text string) []Match {
	var matches []Match
	textLower := strings.ToLower(text)

	for _, p := range ac.patterns {
		patternLower := strings.ToLower(p.text)
		idx := 0
		for {
			pos := strings.Index(textLower[idx:], patternLower)
			if pos == -1 {
				break
			}
			actualPos := idx + pos
			matches = append(matches, Match{
				Start:      actualPos,
				End:        actualPos + len(p.text),
				PatternID:  p.id,
				Confidence: 0.9,
			})
			idx = actualPos + 1
		}
	}

	return matches
}

func (ac *AhoCorasick) Close() {}

type Tokenizer struct{}

func NewTokenizer() *Tokenizer {
	return &Tokenizer{}
}

func (t *Tokenizer) Tokenize(text string) []Token {
	var tokens []Token
	var current strings.Builder
	var currentStart int

	for i, r := range text {
		if unicode.IsSpace(r) {
			if current.Len() > 0 {
				tokens = append(tokens, Token{
					Text:  current.String(),
					Type:  0,
					Start: currentStart,
					End:   i,
				})
				current.Reset()
			}
			tokens = append(tokens, Token{
				Text:  string(r),
				Type:  5,
				Start: i,
				End:   i + 1,
			})
			currentStart = i + 1
		} else if unicode.IsPunct(r) && r != '@' && r != '.' && r != '_' {
			if current.Len() > 0 {
				tokens = append(tokens, Token{
					Text:  current.String(),
					Type:  0,
					Start: currentStart,
					End:   i,
				})
				current.Reset()
			}
			tokens = append(tokens, Token{
				Text:  string(r),
				Type:  5,
				Start: i,
				End:   i + 1,
			})
			currentStart = i + 1
		} else {
			if current.Len() == 0 {
				currentStart = i
			}
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, Token{
			Text:  current.String(),
			Type:  0,
			Start: currentStart,
			End:   len(text),
		})
	}

	return tokens
}

func (t *Tokenizer) Close() {}

type EntityMatcher struct {
	datePatterns   []*regexp.Regexp
	amountPatterns []*regexp.Regexp
	emailPattern   *regexp.Regexp
	keywords       map[EntityType][]string
}

func NewEntityMatcher() *EntityMatcher {
	return &EntityMatcher{
		keywords:     make(map[EntityType][]string),
		emailPattern: regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
	}
}

func (em *EntityMatcher) AddDatePatterns() {
	em.datePatterns = []*regexp.Regexp{
		regexp.MustCompile(`\d{1,2}[/-]\d{1,2}[/-]\d{2,4}`),
		regexp.MustCompile(`\d{4}[/-]\d{1,2}[/-]\d{1,2}`),
		regexp.MustCompile(`(?i)(January|February|March|April|May|June|July|August|September|October|November|December)\s+\d{1,2},?\s+\d{4}`),
		regexp.MustCompile(`(?i)(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{1,2},?\s+\d{4}`),
	}
}

func (em *EntityMatcher) AddAmountPatterns() {
	em.amountPatterns = []*regexp.Regexp{
		regexp.MustCompile(`\$[\d,]+(\.\d{2})?`),
		regexp.MustCompile(`(?i)[\d,]+\s*(USD|EUR|GBP|dollars?|euros?)`),
		regexp.MustCompile(`(?i)\d+\s*(million|billion|thousand|[MBK])\b`),
	}
}

func (em *EntityMatcher) AddKeywords(entityType EntityType, keywords []string) {
	em.keywords[entityType] = append(em.keywords[entityType], keywords...)
}

func (em *EntityMatcher) Extract(text string) []Entity {
	var entities []Entity

	for _, re := range em.datePatterns {
		matches := re.FindAllStringIndex(text, -1)
		for _, m := range matches {
			entities = append(entities, Entity{
				Value:      text[m[0]:m[1]],
				Type:       EntityTypeDate,
				Start:      m[0],
				End:        m[1],
				Confidence: 0.85,
			})
		}
	}

	for _, re := range em.amountPatterns {
		matches := re.FindAllStringIndex(text, -1)
		for _, m := range matches {
			entities = append(entities, Entity{
				Value:      text[m[0]:m[1]],
				Type:       EntityTypeAmount,
				Start:      m[0],
				End:        m[1],
				Confidence: 0.9,
			})
		}
	}

	if em.emailPattern != nil {
		matches := em.emailPattern.FindAllStringIndex(text, -1)
		for _, m := range matches {
			entities = append(entities, Entity{
				Value:      text[m[0]:m[1]],
				Type:       EntityTypeEmail,
				Start:      m[0],
				End:        m[1],
				Confidence: 0.95,
			})
		}
	}

	textLower := strings.ToLower(text)
	for entityType, keywords := range em.keywords {
		for _, kw := range keywords {
			kwLower := strings.ToLower(kw)
			idx := 0
			for {
				pos := strings.Index(textLower[idx:], kwLower)
				if pos == -1 {
					break
				}
				actualPos := idx + pos
				entities = append(entities, Entity{
					Value:      text[actualPos : actualPos+len(kw)],
					Type:       entityType,
					Start:      actualPos,
					End:        actualPos + len(kw),
					Confidence: 0.7,
				})
				idx = actualPos + 1
			}
		}
	}

	return entities
}

func (em *EntityMatcher) ExtractType(text string, entityType EntityType) []Entity {
	all := em.Extract(text)
	var filtered []Entity
	for _, e := range all {
		if e.Type == entityType {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func (em *EntityMatcher) Close() {}

func ExtractAllParallel(text string) []Entity {
	em := NewEntityMatcher()
	em.AddDatePatterns()
	em.AddAmountPatterns()
	return em.Extract(text)
}

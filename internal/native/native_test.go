package native

import (
	"testing"
)

func TestPatternMatcher(t *testing.T) {
	pm := NewPatternMatcher()
	defer pm.Close()

	pm.AddPattern("hello", 1, 0.9)
	pm.AddPattern("world", 2, 0.8)

	matches := pm.FindAll("hello world, hello again")

	if len(matches) < 3 {
		t.Errorf("Expected at least 3 matches, got %d", len(matches))
	}

	for _, m := range matches {
		t.Logf("Match: start=%d, end=%d, patternID=%d, confidence=%.2f",
			m.Start, m.End, m.PatternID, m.Confidence)
	}
}

func TestAhoCorasick(t *testing.T) {
	ac := NewAhoCorasick()
	defer ac.Close()

	ac.AddPattern("he", 1)
	ac.AddPattern("she", 2)
	ac.AddPattern("his", 3)
	ac.AddPattern("hers", 4)

	ac.Build()

	matches := ac.Search("ushers")

	if len(matches) < 2 {
		t.Errorf("Expected at least 2 matches, got %d", len(matches))
	}

	for _, m := range matches {
		t.Logf("AC Match: start=%d, end=%d, patternID=%d",
			m.Start, m.End, m.PatternID)
	}
}

func TestTokenizer(t *testing.T) {
	tok := NewTokenizer()
	defer tok.Close()

	tokens := tok.Tokenize("Hello, World! user@email.com $100")

	if len(tokens) == 0 {
		t.Error("Expected tokens, got none")
	}

	for _, tok := range tokens {
		t.Logf("Token: text=%q, type=%d, start=%d, end=%d",
			tok.Text, tok.Type, tok.Start, tok.End)
	}
}

func TestEntityMatcher(t *testing.T) {
	em := NewEntityMatcher()
	defer em.Close()

	em.AddDatePatterns()
	em.AddAmountPatterns()

	text := "Meeting on 2024-01-15 about the $5 million deal with john@acme.com"
	entities := em.Extract(text)

	if len(entities) < 3 {
		t.Errorf("Expected at least 3 entities, got %d", len(entities))
	}

	for _, e := range entities {
		t.Logf("Entity: value=%q, type=%d, start=%d, end=%d, confidence=%.2f",
			e.Value, e.Type, e.Start, e.End, e.Confidence)
	}
}

func TestExtractAllParallel(t *testing.T) {
	text := "Payment of $1,000,000 on 12/25/2024 from ceo@company.com"
	entities := ExtractAllParallel(text)

	if len(entities) < 2 {
		t.Errorf("Expected at least 2 entities, got %d", len(entities))
	}

	for _, e := range entities {
		t.Logf("Parallel Entity: value=%q, type=%d, confidence=%.2f",
			e.Value, e.Type, e.Confidence)
	}
}

func BenchmarkPatternMatcher(b *testing.B) {
	pm := NewPatternMatcher()
	defer pm.Close()

	keywords := []string{"meeting", "deadline", "project", "budget", "team", "schedule"}
	for i, kw := range keywords {
		pm.AddPattern(kw, i, 0.8)
	}

	text := "The meeting deadline for the project budget review is next week. The team schedule needs to be updated for the meeting."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.FindAll(text)
	}
}

func BenchmarkEntityMatcher(b *testing.B) {
	em := NewEntityMatcher()
	defer em.Close()

	em.AddDatePatterns()
	em.AddAmountPatterns()

	text := "Payment of $1,000,000 due on 2024-12-25 from john.doe@company.com for Q4 report"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		em.Extract(text)
	}
}

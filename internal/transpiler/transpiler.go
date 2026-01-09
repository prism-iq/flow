package transpiler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"flow/internal/config"
)

type Transpiler struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Transpiler {
	return &Transpiler{cfg: cfg}
}

func (t *Transpiler) Transpile(source, filename string) (string, error) {
	// Check cache first
	if t.cfg.UseCache {
		if cached, ok := t.checkCache(source); ok {
			if t.cfg.Debug {
				fmt.Println("[DEBUG] Cache hit")
			}
			return cached, nil
		}
	}

	// Call Claude API
	cppCode, err := t.callClaude(source, filename)
	if err != nil {
		return "", err
	}

	// Cache the result
	if t.cfg.UseCache {
		t.saveCache(source, cppCode)
	}

	return cppCode, nil
}

func (t *Transpiler) TranspileWithFeedback(source, cppCode, errorMsg string) (string, error) {
	return t.callClaudeWithError(source, cppCode, errorMsg)
}

func (t *Transpiler) callClaude(source, filename string) (string, error) {
	if t.cfg.APIKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	// Load syntax documentation for context
	syntaxDoc := t.loadSyntaxDoc()

	systemPrompt := t.buildSystemPrompt(syntaxDoc)

	userPrompt := fmt.Sprintf("Transpile this Flow code to C++17:\n\nFile: %s\n\n%s\n\nOutput ONLY the C++ code, no explanations or markdown.", filename, source)

	return t.apiCall(systemPrompt, userPrompt)
}

func (t *Transpiler) callClaudeWithError(source, cppCode, errorMsg string) (string, error) {
	if t.cfg.APIKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	systemPrompt := `You are a C++ error fixer. Fix compilation errors in C++ code.
Output ONLY the corrected C++ code, no explanations.`

	userPrompt := fmt.Sprintf(`The following C++ code (transpiled from Flow) failed to compile:

Compiler Error:
%s

C++ Code:
%s

Original Flow Source:
%s

Fix the C++ code to resolve the compilation error. Output ONLY the corrected C++ code.`, errorMsg, cppCode, source)

	return t.apiCall(systemPrompt, userPrompt)
}

func (t *Transpiler) buildSystemPrompt(syntaxDoc string) string {
	return fmt.Sprintf(`You are the Flow-to-C++ transpiler. Flow v0.3 reads like English.

CRITICAL RULES:
1. Output ONLY valid C++ code - no markdown, no explanations, no code fences
2. Use modern C++17 features
3. Always include necessary headers (#include <iostream>, #include <string>, etc.)
4. Always return 0 from main()
5. Preserve the logic and intent exactly

FLOW v0.3 VOCABULARY → C++ TRANSLATION:

ENTRY POINT:
- to start: → int main() { ... return 0; }

VARIABLES:
- name is "x" → const auto name = std::string("x");
- age is 25 → const auto age = 25;
- x is 0, can change → auto x = 0;
- x becomes 5 → x = 5;

FUNCTIONS:
- to greet someone: → void greet(const std::string& someone) {
- to add a and b: → auto add(auto a, auto b) {
- return x → return x;

STRUCTS:
- a Person has: → struct Person {
    name as text →     std::string name;
    age as number →    int age;
- a Person can greet: → void greet() const {
- my name → name (or this->name)
- bob's name → bob.name

CONDITIONS:
- if x: → if (x) {
- otherwise: → } else {
- otherwise if x: → } else if (x) {

LOOPS:
- for each item in list: → for (const auto& item : list) {
- for each i in 1 to 10: → for (int i = 1; i <= 10; i++) {
- repeat 5 times: → for (int i = 0; i < 5; i++) {
- while x < 10: → while (x < 10) {
- skip → continue;
- stop → break;

LOGIC:
- and → &&
- or → ||
- not → !
- yes → true
- no → false
- x is 0 → x == 0

OUTPUT:
- say "text" → std::cout << "text" << std::endl;
- say x → std::cout << x << std::endl;
- "Hello, {name}" → "Hello, " << name (string interpolation)

COLLECTIONS:
- numbers are [1, 2, 3] → std::vector<int> numbers = {1, 2, 3};
- numbers at 0 → numbers[0]

SYNTAX REFERENCE:
%s

Output the C++ code directly. No markdown. No explanations. Just valid C++ code.`, syntaxDoc)
}

func (t *Transpiler) loadSyntaxDoc() string {
	syntaxPath := filepath.Join(t.cfg.DocsDir, "SYNTAX.md")
	content, err := os.ReadFile(syntaxPath)
	if err != nil {
		// Fallback to minimal syntax reference
		return `Flow v0.3 uses natural English:
- to start: for entry point
- to greet someone: for functions
- return x for return values
- a Person has: for structs
- a Person can greet: for methods
- say "hello" for print
- for each i in 1 to 10: for loops
- repeat 5 times: for counted loops
- while x < 10: for while loops
- if/otherwise for conditionals
- name is "x" for assignment
- x becomes 5 for reassignment
- x, can change for mutable
- yes/no for booleans
- and/or/not for logic
- skip/stop for continue/break`
	}
	// Return just the transpilation rules section if file is too long
	if len(content) > 4000 {
		// Find the transpilation rules section
		text := string(content)
		if idx := strings.Index(text, "## Transpilation Rules"); idx != -1 {
			return text[idx:]
		}
	}
	return string(content)
}

func (t *Transpiler) apiCall(systemPrompt, userPrompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model":      "claude-sonnet-4-20250514",
		"max_tokens": 4096,
		"system":     systemPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", t.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	cppCode := result.Content[0].Text

	// Clean up any markdown code fences that might have slipped through
	cppCode = strings.TrimPrefix(cppCode, "```cpp\n")
	cppCode = strings.TrimPrefix(cppCode, "```c++\n")
	cppCode = strings.TrimPrefix(cppCode, "```\n")
	cppCode = strings.TrimSuffix(cppCode, "\n```")
	cppCode = strings.TrimSuffix(cppCode, "```")

	return strings.TrimSpace(cppCode), nil
}

// Cache functions

type cacheEntry struct {
	Cpp       string    `json:"cpp"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

func (t *Transpiler) cacheKey(source string) string {
	hash := sha256.Sum256([]byte(source))
	return hex.EncodeToString(hash[:])
}

func (t *Transpiler) checkCache(source string) (string, bool) {
	cacheFile := filepath.Join(t.cfg.CacheDir, "patterns.json")
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return "", false
	}

	var cache map[string]cacheEntry
	if err := json.Unmarshal(data, &cache); err != nil {
		return "", false
	}

	key := t.cacheKey(source)
	if entry, ok := cache[key]; ok {
		// Check if cache is not too old (30 days)
		if time.Since(entry.Timestamp) < 30*24*time.Hour {
			return entry.Cpp, true
		}
	}

	return "", false
}

func (t *Transpiler) saveCache(source, cpp string) {
	cacheFile := filepath.Join(t.cfg.CacheDir, "patterns.json")

	// Load existing cache
	var cache map[string]cacheEntry
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		cache = make(map[string]cacheEntry)
	} else {
		if err := json.Unmarshal(data, &cache); err != nil {
			cache = make(map[string]cacheEntry)
		}
	}

	// Add new entry
	key := t.cacheKey(source)
	cache[key] = cacheEntry{
		Cpp:       cpp,
		Timestamp: time.Now(),
		Version:   "0.3.0",
	}

	// Save cache
	data, err = json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return
	}

	os.MkdirAll(t.cfg.CacheDir, 0755)
	os.WriteFile(cacheFile, data, 0644)
}

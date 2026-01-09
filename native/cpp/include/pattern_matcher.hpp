#pragma once

#include <cstdint>
#include <cstddef>
#include <vector>
#include <string>
#include <string_view>
#include <array>
#include <memory>
#include <unordered_map>

#ifdef __AVX2__
#include <immintrin.h>
#endif

namespace flow {

struct Match {
    size_t start;
    size_t end;
    size_t pattern_id;
    float confidence;
};

struct PatternConfig {
    bool case_sensitive = false;
    bool whole_word = false;
    float base_confidence = 0.8f;
};

class SIMDMatcher {
public:
    SIMDMatcher() = default;

    void add_pattern(std::string_view pattern, size_t id, float confidence = 0.8f);
    std::vector<Match> find_all(std::string_view text) const;
    size_t count_matches(std::string_view text) const;

private:
    struct Pattern {
        std::string text;
        std::string text_lower;
        size_t id;
        float confidence;
    };

    std::vector<Pattern> patterns_;
    bool case_sensitive_ = false;

#ifdef __AVX2__
    __m256i find_char_avx2(const char* data, size_t len, char c) const;
    std::vector<size_t> find_all_char_avx2(const char* data, size_t len, char c) const;
#endif

    std::vector<size_t> find_all_char_scalar(const char* data, size_t len, char c) const;
};

class AhoCorasick {
public:
    AhoCorasick();
    ~AhoCorasick();

    void add_pattern(const std::string& pattern, size_t id);
    void build();
    std::vector<Match> search(std::string_view text) const;

private:
    static constexpr size_t ALPHABET_SIZE = 256;

    struct Node {
        std::array<int, ALPHABET_SIZE> children;
        int fail = 0;
        std::vector<std::pair<size_t, size_t>> outputs;

        Node() { children.fill(-1); }
    };

    std::vector<Node> nodes_;
    std::vector<std::string> patterns_;
    bool built_ = false;

    void build_failure_links();
};

class FastTokenizer {
public:
    enum class TokenType {
        Word,
        Number,
        Date,
        Email,
        Currency,
        Punctuation,
        Whitespace,
        Unknown
    };

    struct Token {
        std::string_view text;
        TokenType type;
        size_t start;
        size_t end;
    };

    FastTokenizer();

    std::vector<Token> tokenize(std::string_view text) const;
    std::vector<std::string_view> split_words(std::string_view text) const;

private:
    bool is_word_char(char c) const;
    bool is_digit(char c) const;
    bool is_whitespace(char c) const;
    TokenType classify_token(std::string_view token) const;

    std::array<bool, 256> word_chars_;
    std::array<bool, 256> digit_chars_;
    std::array<bool, 256> whitespace_chars_;
};

class EntityMatcher {
public:
    enum class EntityType {
        Date,
        Person,
        Organization,
        Amount,
        Email,
        Unknown
    };

    struct Entity {
        std::string value;
        EntityType type;
        size_t start;
        size_t end;
        float confidence;
        std::unordered_map<std::string, std::string> metadata;
    };

    EntityMatcher();

    void add_date_patterns();
    void add_amount_patterns();
    void add_email_pattern();
    void add_keywords(EntityType type, const std::vector<std::string>& keywords);

    std::vector<Entity> extract(std::string_view text) const;
    std::vector<Entity> extract_type(std::string_view text, EntityType type) const;

private:
    std::unique_ptr<AhoCorasick> keyword_matcher_;
    std::vector<std::pair<EntityType, std::string>> date_patterns_;
    std::vector<std::pair<EntityType, std::string>> amount_patterns_;
    std::unordered_map<size_t, EntityType> keyword_types_;
    size_t next_keyword_id_ = 0;

    std::vector<Entity> extract_dates(std::string_view text) const;
    std::vector<Entity> extract_amounts(std::string_view text) const;
    std::vector<Entity> extract_emails(std::string_view text) const;
};

}

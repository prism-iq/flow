#include "pattern_matcher.hpp"
#include <algorithm>
#include <cctype>
#include <queue>
#include <cstring>
#include <regex>

namespace flow {

void SIMDMatcher::add_pattern(std::string_view pattern, size_t id, float confidence) {
    Pattern p;
    p.text = std::string(pattern);
    p.text_lower.reserve(pattern.size());
    for (char c : pattern) {
        p.text_lower.push_back(static_cast<char>(std::tolower(static_cast<unsigned char>(c))));
    }
    p.id = id;
    p.confidence = confidence;
    patterns_.push_back(std::move(p));
}

#ifdef __AVX2__
std::vector<size_t> SIMDMatcher::find_all_char_avx2(const char* data, size_t len, char c) const {
    std::vector<size_t> positions;
    __m256i needle = _mm256_set1_epi8(c);

    size_t i = 0;
    for (; i + 32 <= len; i += 32) {
        __m256i chunk = _mm256_loadu_si256(reinterpret_cast<const __m256i*>(data + i));
        __m256i cmp = _mm256_cmpeq_epi8(chunk, needle);
        uint32_t mask = static_cast<uint32_t>(_mm256_movemask_epi8(cmp));

        while (mask) {
            int bit = __builtin_ctz(mask);
            positions.push_back(i + bit);
            mask &= mask - 1;
        }
    }

    for (; i < len; i++) {
        if (data[i] == c) {
            positions.push_back(i);
        }
    }

    return positions;
}
#endif

std::vector<size_t> SIMDMatcher::find_all_char_scalar(const char* data, size_t len, char c) const {
    std::vector<size_t> positions;
    for (size_t i = 0; i < len; i++) {
        if (data[i] == c) {
            positions.push_back(i);
        }
    }
    return positions;
}

std::vector<Match> SIMDMatcher::find_all(std::string_view text) const {
    std::vector<Match> matches;

    if (patterns_.empty() || text.empty()) {
        return matches;
    }

    std::string text_lower;
    if (!case_sensitive_) {
        text_lower.reserve(text.size());
        for (char c : text) {
            text_lower.push_back(static_cast<char>(std::tolower(static_cast<unsigned char>(c))));
        }
    }

    const std::string_view search_text = case_sensitive_ ? text : text_lower;

    for (const auto& pattern : patterns_) {
        const std::string& needle = case_sensitive_ ? pattern.text : pattern.text_lower;

        if (needle.empty() || needle.size() > search_text.size()) {
            continue;
        }

        char first_char = needle[0];
        std::vector<size_t> candidates;

#ifdef __AVX2__
        candidates = find_all_char_avx2(search_text.data(), search_text.size(), first_char);
#else
        candidates = find_all_char_scalar(search_text.data(), search_text.size(), first_char);
#endif

        for (size_t pos : candidates) {
            if (pos + needle.size() <= search_text.size()) {
                if (std::memcmp(search_text.data() + pos, needle.data(), needle.size()) == 0) {
                    matches.push_back({
                        pos,
                        pos + needle.size(),
                        pattern.id,
                        pattern.confidence
                    });
                }
            }
        }
    }

    std::sort(matches.begin(), matches.end(),
              [](const Match& a, const Match& b) { return a.start < b.start; });

    return matches;
}

size_t SIMDMatcher::count_matches(std::string_view text) const {
    return find_all(text).size();
}

AhoCorasick::AhoCorasick() {
    nodes_.emplace_back();
}

AhoCorasick::~AhoCorasick() = default;

void AhoCorasick::add_pattern(const std::string& pattern, size_t id) {
    if (pattern.empty()) return;

    patterns_.push_back(pattern);

    int node = 0;
    for (unsigned char c : pattern) {
        if (nodes_[node].children[c] == -1) {
            nodes_[node].children[c] = static_cast<int>(nodes_.size());
            nodes_.emplace_back();
        }
        node = nodes_[node].children[c];
    }

    nodes_[node].outputs.push_back({id, pattern.size()});
    built_ = false;
}

void AhoCorasick::build() {
    if (built_) return;

    std::queue<int> queue;

    for (size_t c = 0; c < ALPHABET_SIZE; c++) {
        if (nodes_[0].children[c] != -1) {
            nodes_[nodes_[0].children[c]].fail = 0;
            queue.push(nodes_[0].children[c]);
        }
    }

    while (!queue.empty()) {
        int curr = queue.front();
        queue.pop();

        for (size_t c = 0; c < ALPHABET_SIZE; c++) {
            int child = nodes_[curr].children[c];
            if (child == -1) continue;

            int fail = nodes_[curr].fail;
            while (fail != 0 && nodes_[fail].children[c] == -1) {
                fail = nodes_[fail].fail;
            }

            nodes_[child].fail = (nodes_[fail].children[c] != -1 && nodes_[fail].children[c] != child)
                                     ? nodes_[fail].children[c]
                                     : 0;

            for (const auto& output : nodes_[nodes_[child].fail].outputs) {
                nodes_[child].outputs.push_back(output);
            }

            queue.push(child);
        }
    }

    built_ = true;
}

std::vector<Match> AhoCorasick::search(std::string_view text) const {
    std::vector<Match> matches;

    if (!built_ || text.empty()) {
        return matches;
    }

    int state = 0;
    for (size_t i = 0; i < text.size(); i++) {
        unsigned char c = static_cast<unsigned char>(text[i]);

        while (state != 0 && nodes_[state].children[c] == -1) {
            state = nodes_[state].fail;
        }

        if (nodes_[state].children[c] != -1) {
            state = nodes_[state].children[c];
        }

        for (const auto& [pattern_id, pattern_len] : nodes_[state].outputs) {
            matches.push_back({
                i - pattern_len + 1,
                i + 1,
                pattern_id,
                0.9f
            });
        }
    }

    return matches;
}

FastTokenizer::FastTokenizer() {
    word_chars_.fill(false);
    digit_chars_.fill(false);
    whitespace_chars_.fill(false);

    for (int i = 'a'; i <= 'z'; i++) word_chars_[i] = true;
    for (int i = 'A'; i <= 'Z'; i++) word_chars_[i] = true;
    for (int i = '0'; i <= '9'; i++) {
        word_chars_[i] = true;
        digit_chars_[i] = true;
    }
    word_chars_['_'] = true;
    word_chars_['\''] = true;

    whitespace_chars_[' '] = true;
    whitespace_chars_['\t'] = true;
    whitespace_chars_['\n'] = true;
    whitespace_chars_['\r'] = true;
}

bool FastTokenizer::is_word_char(char c) const {
    return word_chars_[static_cast<unsigned char>(c)];
}

bool FastTokenizer::is_digit(char c) const {
    return digit_chars_[static_cast<unsigned char>(c)];
}

bool FastTokenizer::is_whitespace(char c) const {
    return whitespace_chars_[static_cast<unsigned char>(c)];
}

FastTokenizer::TokenType FastTokenizer::classify_token(std::string_view token) const {
    if (token.empty()) return TokenType::Unknown;

    if (token.find('@') != std::string_view::npos && token.find('.') != std::string_view::npos) {
        return TokenType::Email;
    }

    if (token[0] == '$' || token[0] == '\xe2') {
        return TokenType::Currency;
    }

    bool has_digit = false;
    bool has_alpha = false;
    bool has_date_sep = false;

    for (char c : token) {
        if (is_digit(c)) has_digit = true;
        else if (std::isalpha(static_cast<unsigned char>(c))) has_alpha = true;
        if (c == '/' || c == '-') has_date_sep = true;
    }

    if (has_digit && has_date_sep && !has_alpha) {
        return TokenType::Date;
    }

    if (has_digit && !has_alpha) {
        return TokenType::Number;
    }

    if (has_alpha) {
        return TokenType::Word;
    }

    if (is_whitespace(token[0])) {
        return TokenType::Whitespace;
    }

    return TokenType::Punctuation;
}

std::vector<FastTokenizer::Token> FastTokenizer::tokenize(std::string_view text) const {
    std::vector<Token> tokens;

    if (text.empty()) return tokens;

    size_t start = 0;
    size_t i = 0;

    while (i < text.size()) {
        if (is_whitespace(text[i])) {
            if (i > start) {
                std::string_view tok = text.substr(start, i - start);
                tokens.push_back({tok, classify_token(tok), start, i});
            }

            size_t ws_start = i;
            while (i < text.size() && is_whitespace(text[i])) i++;
            tokens.push_back({text.substr(ws_start, i - ws_start), TokenType::Whitespace, ws_start, i});
            start = i;
        } else if (is_word_char(text[i]) || text[i] == '@' || text[i] == '.' || text[i] == '$') {
            size_t tok_start = i;
            while (i < text.size() && (is_word_char(text[i]) || text[i] == '@' || text[i] == '.' ||
                                       text[i] == '/' || text[i] == '-' || text[i] == '$' || text[i] == ',')) {
                i++;
            }
            std::string_view tok = text.substr(tok_start, i - tok_start);
            tokens.push_back({tok, classify_token(tok), tok_start, i});
            start = i;
        } else {
            tokens.push_back({text.substr(i, 1), TokenType::Punctuation, i, i + 1});
            i++;
            start = i;
        }
    }

    return tokens;
}

std::vector<std::string_view> FastTokenizer::split_words(std::string_view text) const {
    std::vector<std::string_view> words;

    auto tokens = tokenize(text);
    for (const auto& tok : tokens) {
        if (tok.type == TokenType::Word) {
            words.push_back(tok.text);
        }
    }

    return words;
}

EntityMatcher::EntityMatcher() : keyword_matcher_(std::make_unique<AhoCorasick>()) {}

void EntityMatcher::add_date_patterns() {
    date_patterns_ = {
        {EntityType::Date, R"(\d{1,2}[/-]\d{1,2}[/-]\d{2,4})"},
        {EntityType::Date, R"(\d{4}[/-]\d{1,2}[/-]\d{1,2})"},
        {EntityType::Date, R"((January|February|March|April|May|June|July|August|September|October|November|December)\s+\d{1,2},?\s+\d{4})"},
        {EntityType::Date, R"((Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{1,2},?\s+\d{4})"},
    };
}

void EntityMatcher::add_amount_patterns() {
    amount_patterns_ = {
        {EntityType::Amount, R"(\$[\d,]+(\.\d{2})?)"},
        {EntityType::Amount, R"([\d,]+\s*(USD|EUR|GBP|dollars?|euros?))"},
        {EntityType::Amount, R"(\d+\s*(million|billion|thousand|[MBK])\b)"},
    };
}

void EntityMatcher::add_email_pattern() {
}

void EntityMatcher::add_keywords(EntityType type, const std::vector<std::string>& keywords) {
    for (const auto& kw : keywords) {
        keyword_matcher_->add_pattern(kw, next_keyword_id_);
        keyword_types_[next_keyword_id_] = type;
        next_keyword_id_++;
    }
}

std::vector<EntityMatcher::Entity> EntityMatcher::extract_dates(std::string_view text) const {
    std::vector<Entity> entities;
    std::string text_str(text);

    for (const auto& [type, pattern] : date_patterns_) {
        try {
            std::regex re(pattern, std::regex::icase);
            auto begin = std::sregex_iterator(text_str.begin(), text_str.end(), re);
            auto end = std::sregex_iterator();

            for (auto it = begin; it != end; ++it) {
                Entity e;
                e.value = it->str();
                e.type = EntityType::Date;
                e.start = static_cast<size_t>(it->position());
                e.end = e.start + e.value.size();
                e.confidence = 0.85f;
                entities.push_back(std::move(e));
            }
        } catch (...) {
        }
    }

    return entities;
}

std::vector<EntityMatcher::Entity> EntityMatcher::extract_amounts(std::string_view text) const {
    std::vector<Entity> entities;
    std::string text_str(text);

    for (const auto& [type, pattern] : amount_patterns_) {
        try {
            std::regex re(pattern, std::regex::icase);
            auto begin = std::sregex_iterator(text_str.begin(), text_str.end(), re);
            auto end = std::sregex_iterator();

            for (auto it = begin; it != end; ++it) {
                Entity e;
                e.value = it->str();
                e.type = EntityType::Amount;
                e.start = static_cast<size_t>(it->position());
                e.end = e.start + e.value.size();
                e.confidence = 0.9f;
                entities.push_back(std::move(e));
            }
        } catch (...) {
        }
    }

    return entities;
}

std::vector<EntityMatcher::Entity> EntityMatcher::extract_emails(std::string_view text) const {
    std::vector<Entity> entities;
    std::string text_str(text);

    try {
        std::regex email_re(R"([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})");
        auto begin = std::sregex_iterator(text_str.begin(), text_str.end(), email_re);
        auto end = std::sregex_iterator();

        for (auto it = begin; it != end; ++it) {
            Entity e;
            e.value = it->str();
            e.type = EntityType::Email;
            e.start = static_cast<size_t>(it->position());
            e.end = e.start + e.value.size();
            e.confidence = 0.95f;
            entities.push_back(std::move(e));
        }
    } catch (...) {
    }

    return entities;
}

std::vector<EntityMatcher::Entity> EntityMatcher::extract(std::string_view text) const {
    std::vector<Entity> all_entities;

    auto dates = extract_dates(text);
    auto amounts = extract_amounts(text);
    auto emails = extract_emails(text);

    all_entities.insert(all_entities.end(), dates.begin(), dates.end());
    all_entities.insert(all_entities.end(), amounts.begin(), amounts.end());
    all_entities.insert(all_entities.end(), emails.begin(), emails.end());

    if (!keyword_types_.empty()) {
        keyword_matcher_->build();
        auto keyword_matches = keyword_matcher_->search(text);

        for (const auto& match : keyword_matches) {
            auto it = keyword_types_.find(match.pattern_id);
            if (it != keyword_types_.end()) {
                Entity e;
                e.value = std::string(text.substr(match.start, match.end - match.start));
                e.type = it->second;
                e.start = match.start;
                e.end = match.end;
                e.confidence = match.confidence;
                all_entities.push_back(std::move(e));
            }
        }
    }

    std::sort(all_entities.begin(), all_entities.end(),
              [](const Entity& a, const Entity& b) { return a.start < b.start; });

    return all_entities;
}

std::vector<EntityMatcher::Entity> EntityMatcher::extract_type(std::string_view text, EntityType type) const {
    std::vector<Entity> entities;

    switch (type) {
        case EntityType::Date:
            return extract_dates(text);
        case EntityType::Amount:
            return extract_amounts(text);
        case EntityType::Email:
            return extract_emails(text);
        default:
            break;
    }

    auto all = extract(text);
    for (auto& e : all) {
        if (e.type == type) {
            entities.push_back(std::move(e));
        }
    }

    return entities;
}

}

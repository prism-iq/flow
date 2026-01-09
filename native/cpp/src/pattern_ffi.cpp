#include "pattern_ffi.hpp"
#include "pattern_matcher.hpp"
#include <cstring>
#include <thread>
#include <future>
#include <vector>

extern "C" {

FlowPatternMatcherHandle flow_pattern_matcher_create() {
    return new flow::SIMDMatcher();
}

void flow_pattern_matcher_destroy(FlowPatternMatcherHandle handle) {
    delete static_cast<flow::SIMDMatcher*>(handle);
}

void flow_pattern_matcher_add_pattern(FlowPatternMatcherHandle handle,
                                       const char* pattern,
                                       size_t id,
                                       float confidence) {
    if (!handle || !pattern) return;
    auto* matcher = static_cast<flow::SIMDMatcher*>(handle);
    matcher->add_pattern(pattern, id, confidence);
}

int flow_pattern_matcher_find_all(FlowPatternMatcherHandle handle,
                                   const char* text,
                                   size_t text_len,
                                   FlowMatch** matches,
                                   size_t* num_matches) {
    if (!handle || !text || !matches || !num_matches) return -1;

    auto* matcher = static_cast<flow::SIMDMatcher*>(handle);
    auto results = matcher->find_all(std::string_view(text, text_len));

    *num_matches = results.size();
    if (results.empty()) {
        *matches = nullptr;
        return 0;
    }

    *matches = static_cast<FlowMatch*>(malloc(sizeof(FlowMatch) * results.size()));
    if (!*matches) return -2;

    for (size_t i = 0; i < results.size(); i++) {
        (*matches)[i].start = results[i].start;
        (*matches)[i].end = results[i].end;
        (*matches)[i].pattern_id = results[i].pattern_id;
        (*matches)[i].confidence = results[i].confidence;
    }

    return 0;
}

void flow_pattern_matcher_free_matches(FlowMatch* matches) {
    free(matches);
}

FlowAhoCorasickHandle flow_aho_corasick_create() {
    return new flow::AhoCorasick();
}

void flow_aho_corasick_destroy(FlowAhoCorasickHandle handle) {
    delete static_cast<flow::AhoCorasick*>(handle);
}

void flow_aho_corasick_add_pattern(FlowAhoCorasickHandle handle,
                                    const char* pattern,
                                    size_t id) {
    if (!handle || !pattern) return;
    auto* ac = static_cast<flow::AhoCorasick*>(handle);
    ac->add_pattern(pattern, id);
}

void flow_aho_corasick_build(FlowAhoCorasickHandle handle) {
    if (!handle) return;
    auto* ac = static_cast<flow::AhoCorasick*>(handle);
    ac->build();
}

int flow_aho_corasick_search(FlowAhoCorasickHandle handle,
                              const char* text,
                              size_t text_len,
                              FlowMatch** matches,
                              size_t* num_matches) {
    if (!handle || !text || !matches || !num_matches) return -1;

    auto* ac = static_cast<flow::AhoCorasick*>(handle);
    auto results = ac->search(std::string_view(text, text_len));

    *num_matches = results.size();
    if (results.empty()) {
        *matches = nullptr;
        return 0;
    }

    *matches = static_cast<FlowMatch*>(malloc(sizeof(FlowMatch) * results.size()));
    if (!*matches) return -2;

    for (size_t i = 0; i < results.size(); i++) {
        (*matches)[i].start = results[i].start;
        (*matches)[i].end = results[i].end;
        (*matches)[i].pattern_id = results[i].pattern_id;
        (*matches)[i].confidence = results[i].confidence;
    }

    return 0;
}

FlowTokenizerHandle flow_tokenizer_create() {
    return new flow::FastTokenizer();
}

void flow_tokenizer_destroy(FlowTokenizerHandle handle) {
    delete static_cast<flow::FastTokenizer*>(handle);
}

int flow_tokenizer_tokenize(FlowTokenizerHandle handle,
                             const char* text,
                             size_t text_len,
                             FlowToken** tokens,
                             size_t* num_tokens) {
    if (!handle || !text || !tokens || !num_tokens) return -1;

    auto* tokenizer = static_cast<flow::FastTokenizer*>(handle);
    auto results = tokenizer->tokenize(std::string_view(text, text_len));

    *num_tokens = results.size();
    if (results.empty()) {
        *tokens = nullptr;
        return 0;
    }

    *tokens = static_cast<FlowToken*>(malloc(sizeof(FlowToken) * results.size()));
    if (!*tokens) return -2;

    for (size_t i = 0; i < results.size(); i++) {
        (*tokens)[i].text = static_cast<char*>(malloc(results[i].text.size() + 1));
        if ((*tokens)[i].text) {
            memcpy((*tokens)[i].text, results[i].text.data(), results[i].text.size());
            (*tokens)[i].text[results[i].text.size()] = '\0';
        }
        (*tokens)[i].type = static_cast<int>(results[i].type);
        (*tokens)[i].start = results[i].start;
        (*tokens)[i].end = results[i].end;
    }

    return 0;
}

void flow_tokenizer_free_tokens(FlowToken* tokens, size_t num_tokens) {
    if (!tokens) return;
    for (size_t i = 0; i < num_tokens; i++) {
        free(tokens[i].text);
    }
    free(tokens);
}

FlowEntityMatcherHandle flow_entity_matcher_create() {
    return new flow::EntityMatcher();
}

void flow_entity_matcher_destroy(FlowEntityMatcherHandle handle) {
    delete static_cast<flow::EntityMatcher*>(handle);
}

void flow_entity_matcher_add_date_patterns(FlowEntityMatcherHandle handle) {
    if (!handle) return;
    auto* matcher = static_cast<flow::EntityMatcher*>(handle);
    matcher->add_date_patterns();
}

void flow_entity_matcher_add_amount_patterns(FlowEntityMatcherHandle handle) {
    if (!handle) return;
    auto* matcher = static_cast<flow::EntityMatcher*>(handle);
    matcher->add_amount_patterns();
}

void flow_entity_matcher_add_keywords(FlowEntityMatcherHandle handle,
                                       FlowEntityType type,
                                       const char** keywords,
                                       size_t num_keywords) {
    if (!handle || !keywords) return;
    auto* matcher = static_cast<flow::EntityMatcher*>(handle);

    std::vector<std::string> kw_vec;
    for (size_t i = 0; i < num_keywords; i++) {
        if (keywords[i]) {
            kw_vec.push_back(keywords[i]);
        }
    }

    flow::EntityMatcher::EntityType cpp_type;
    switch (type) {
        case FLOW_ENTITY_DATE: cpp_type = flow::EntityMatcher::EntityType::Date; break;
        case FLOW_ENTITY_PERSON: cpp_type = flow::EntityMatcher::EntityType::Person; break;
        case FLOW_ENTITY_ORGANIZATION: cpp_type = flow::EntityMatcher::EntityType::Organization; break;
        case FLOW_ENTITY_AMOUNT: cpp_type = flow::EntityMatcher::EntityType::Amount; break;
        case FLOW_ENTITY_EMAIL: cpp_type = flow::EntityMatcher::EntityType::Email; break;
        default: cpp_type = flow::EntityMatcher::EntityType::Unknown; break;
    }

    matcher->add_keywords(cpp_type, kw_vec);
}

static FlowEntityType convert_entity_type(flow::EntityMatcher::EntityType type) {
    switch (type) {
        case flow::EntityMatcher::EntityType::Date: return FLOW_ENTITY_DATE;
        case flow::EntityMatcher::EntityType::Person: return FLOW_ENTITY_PERSON;
        case flow::EntityMatcher::EntityType::Organization: return FLOW_ENTITY_ORGANIZATION;
        case flow::EntityMatcher::EntityType::Amount: return FLOW_ENTITY_AMOUNT;
        case flow::EntityMatcher::EntityType::Email: return FLOW_ENTITY_EMAIL;
        default: return FLOW_ENTITY_UNKNOWN;
    }
}

int flow_entity_matcher_extract(FlowEntityMatcherHandle handle,
                                 const char* text,
                                 size_t text_len,
                                 FlowEntity** entities,
                                 size_t* num_entities) {
    if (!handle || !text || !entities || !num_entities) return -1;

    auto* matcher = static_cast<flow::EntityMatcher*>(handle);
    auto results = matcher->extract(std::string_view(text, text_len));

    *num_entities = results.size();
    if (results.empty()) {
        *entities = nullptr;
        return 0;
    }

    *entities = static_cast<FlowEntity*>(malloc(sizeof(FlowEntity) * results.size()));
    if (!*entities) return -2;

    for (size_t i = 0; i < results.size(); i++) {
        (*entities)[i].value = static_cast<char*>(malloc(results[i].value.size() + 1));
        if ((*entities)[i].value) {
            memcpy((*entities)[i].value, results[i].value.c_str(), results[i].value.size());
            (*entities)[i].value[results[i].value.size()] = '\0';
        }
        (*entities)[i].type = convert_entity_type(results[i].type);
        (*entities)[i].start = results[i].start;
        (*entities)[i].end = results[i].end;
        (*entities)[i].confidence = results[i].confidence;
    }

    return 0;
}

int flow_entity_matcher_extract_type(FlowEntityMatcherHandle handle,
                                      const char* text,
                                      size_t text_len,
                                      FlowEntityType type,
                                      FlowEntity** entities,
                                      size_t* num_entities) {
    if (!handle || !text || !entities || !num_entities) return -1;

    auto* matcher = static_cast<flow::EntityMatcher*>(handle);

    flow::EntityMatcher::EntityType cpp_type;
    switch (type) {
        case FLOW_ENTITY_DATE: cpp_type = flow::EntityMatcher::EntityType::Date; break;
        case FLOW_ENTITY_PERSON: cpp_type = flow::EntityMatcher::EntityType::Person; break;
        case FLOW_ENTITY_ORGANIZATION: cpp_type = flow::EntityMatcher::EntityType::Organization; break;
        case FLOW_ENTITY_AMOUNT: cpp_type = flow::EntityMatcher::EntityType::Amount; break;
        case FLOW_ENTITY_EMAIL: cpp_type = flow::EntityMatcher::EntityType::Email; break;
        default: cpp_type = flow::EntityMatcher::EntityType::Unknown; break;
    }

    auto results = matcher->extract_type(std::string_view(text, text_len), cpp_type);

    *num_entities = results.size();
    if (results.empty()) {
        *entities = nullptr;
        return 0;
    }

    *entities = static_cast<FlowEntity*>(malloc(sizeof(FlowEntity) * results.size()));
    if (!*entities) return -2;

    for (size_t i = 0; i < results.size(); i++) {
        (*entities)[i].value = static_cast<char*>(malloc(results[i].value.size() + 1));
        if ((*entities)[i].value) {
            memcpy((*entities)[i].value, results[i].value.c_str(), results[i].value.size());
            (*entities)[i].value[results[i].value.size()] = '\0';
        }
        (*entities)[i].type = convert_entity_type(results[i].type);
        (*entities)[i].start = results[i].start;
        (*entities)[i].end = results[i].end;
        (*entities)[i].confidence = results[i].confidence;
    }

    return 0;
}

void flow_entity_matcher_free_entities(FlowEntity* entities, size_t num_entities) {
    if (!entities) return;
    for (size_t i = 0; i < num_entities; i++) {
        free(entities[i].value);
    }
    free(entities);
}

int flow_extract_all_parallel(const char* text,
                               size_t text_len,
                               FlowEntity** entities,
                               size_t* num_entities) {
    if (!text || !entities || !num_entities) return -1;

    std::string_view text_view(text, text_len);

    auto date_future = std::async(std::launch::async, [text_view]() {
        flow::EntityMatcher matcher;
        matcher.add_date_patterns();
        return matcher.extract_type(text_view, flow::EntityMatcher::EntityType::Date);
    });

    auto amount_future = std::async(std::launch::async, [text_view]() {
        flow::EntityMatcher matcher;
        matcher.add_amount_patterns();
        return matcher.extract_type(text_view, flow::EntityMatcher::EntityType::Amount);
    });

    auto email_future = std::async(std::launch::async, [text_view]() {
        flow::EntityMatcher matcher;
        return matcher.extract_type(text_view, flow::EntityMatcher::EntityType::Email);
    });

    auto dates = date_future.get();
    auto amounts = amount_future.get();
    auto emails = email_future.get();

    std::vector<flow::EntityMatcher::Entity> all_results;
    all_results.reserve(dates.size() + amounts.size() + emails.size());
    all_results.insert(all_results.end(), dates.begin(), dates.end());
    all_results.insert(all_results.end(), amounts.begin(), amounts.end());
    all_results.insert(all_results.end(), emails.begin(), emails.end());

    *num_entities = all_results.size();
    if (all_results.empty()) {
        *entities = nullptr;
        return 0;
    }

    *entities = static_cast<FlowEntity*>(malloc(sizeof(FlowEntity) * all_results.size()));
    if (!*entities) return -2;

    for (size_t i = 0; i < all_results.size(); i++) {
        (*entities)[i].value = static_cast<char*>(malloc(all_results[i].value.size() + 1));
        if ((*entities)[i].value) {
            memcpy((*entities)[i].value, all_results[i].value.c_str(), all_results[i].value.size());
            (*entities)[i].value[all_results[i].value.size()] = '\0';
        }
        (*entities)[i].type = convert_entity_type(all_results[i].type);
        (*entities)[i].start = all_results[i].start;
        (*entities)[i].end = all_results[i].end;
        (*entities)[i].confidence = all_results[i].confidence;
    }

    return 0;
}

}

#pragma once

#include <cstdint>
#include <cstddef>

#ifdef __cplusplus
extern "C" {
#endif

typedef void* FlowPatternMatcherHandle;
typedef void* FlowAhoCorasickHandle;
typedef void* FlowTokenizerHandle;
typedef void* FlowEntityMatcherHandle;

typedef enum {
    FLOW_ENTITY_DATE = 0,
    FLOW_ENTITY_PERSON = 1,
    FLOW_ENTITY_ORGANIZATION = 2,
    FLOW_ENTITY_AMOUNT = 3,
    FLOW_ENTITY_EMAIL = 4,
    FLOW_ENTITY_UNKNOWN = 99
} FlowEntityType;

typedef struct {
    size_t start;
    size_t end;
    size_t pattern_id;
    float confidence;
} FlowMatch;

typedef struct {
    char* value;
    FlowEntityType type;
    size_t start;
    size_t end;
    float confidence;
} FlowEntity;

typedef struct {
    char* text;
    int type;
    size_t start;
    size_t end;
} FlowToken;

FlowPatternMatcherHandle flow_pattern_matcher_create();
void flow_pattern_matcher_destroy(FlowPatternMatcherHandle handle);
void flow_pattern_matcher_add_pattern(FlowPatternMatcherHandle handle,
                                       const char* pattern,
                                       size_t id,
                                       float confidence);
int flow_pattern_matcher_find_all(FlowPatternMatcherHandle handle,
                                   const char* text,
                                   size_t text_len,
                                   FlowMatch** matches,
                                   size_t* num_matches);
void flow_pattern_matcher_free_matches(FlowMatch* matches);

FlowAhoCorasickHandle flow_aho_corasick_create();
void flow_aho_corasick_destroy(FlowAhoCorasickHandle handle);
void flow_aho_corasick_add_pattern(FlowAhoCorasickHandle handle,
                                    const char* pattern,
                                    size_t id);
void flow_aho_corasick_build(FlowAhoCorasickHandle handle);
int flow_aho_corasick_search(FlowAhoCorasickHandle handle,
                              const char* text,
                              size_t text_len,
                              FlowMatch** matches,
                              size_t* num_matches);

FlowTokenizerHandle flow_tokenizer_create();
void flow_tokenizer_destroy(FlowTokenizerHandle handle);
int flow_tokenizer_tokenize(FlowTokenizerHandle handle,
                             const char* text,
                             size_t text_len,
                             FlowToken** tokens,
                             size_t* num_tokens);
void flow_tokenizer_free_tokens(FlowToken* tokens, size_t num_tokens);

FlowEntityMatcherHandle flow_entity_matcher_create();
void flow_entity_matcher_destroy(FlowEntityMatcherHandle handle);
void flow_entity_matcher_add_date_patterns(FlowEntityMatcherHandle handle);
void flow_entity_matcher_add_amount_patterns(FlowEntityMatcherHandle handle);
void flow_entity_matcher_add_keywords(FlowEntityMatcherHandle handle,
                                       FlowEntityType type,
                                       const char** keywords,
                                       size_t num_keywords);
int flow_entity_matcher_extract(FlowEntityMatcherHandle handle,
                                 const char* text,
                                 size_t text_len,
                                 FlowEntity** entities,
                                 size_t* num_entities);
int flow_entity_matcher_extract_type(FlowEntityMatcherHandle handle,
                                      const char* text,
                                      size_t text_len,
                                      FlowEntityType type,
                                      FlowEntity** entities,
                                      size_t* num_entities);
void flow_entity_matcher_free_entities(FlowEntity* entities, size_t num_entities);

int flow_extract_all_parallel(const char* text,
                               size_t text_len,
                               FlowEntity** entities,
                               size_t* num_entities);

#ifdef __cplusplus
}
#endif

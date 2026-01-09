//go:build cgo

package native

/*
#cgo CFLAGS: -I${SRCDIR}/../../native/cpp/include
#cgo LDFLAGS: -L${SRCDIR}/../../native/cpp/build -lflow_synapses -lstdc++ -lm -lpthread

#include "pattern_ffi.hpp"
#include <stdlib.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

type EntityType int

const (
	EntityTypeDate         EntityType = C.FLOW_ENTITY_DATE
	EntityTypePerson       EntityType = C.FLOW_ENTITY_PERSON
	EntityTypeOrganization EntityType = C.FLOW_ENTITY_ORGANIZATION
	EntityTypeAmount       EntityType = C.FLOW_ENTITY_AMOUNT
	EntityTypeEmail        EntityType = C.FLOW_ENTITY_EMAIL
	EntityTypeUnknown      EntityType = C.FLOW_ENTITY_UNKNOWN
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
	handle C.FlowPatternMatcherHandle
}

func NewPatternMatcher() *PatternMatcher {
	pm := &PatternMatcher{
		handle: C.flow_pattern_matcher_create(),
	}
	runtime.SetFinalizer(pm, (*PatternMatcher).Close)
	return pm
}

func (pm *PatternMatcher) AddPattern(pattern string, id int, confidence float32) {
	cPattern := C.CString(pattern)
	defer C.free(unsafe.Pointer(cPattern))

	C.flow_pattern_matcher_add_pattern(pm.handle, cPattern, C.size_t(id), C.float(confidence))
}

func (pm *PatternMatcher) FindAll(text string) []Match {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	var matches *C.FlowMatch
	var numMatches C.size_t

	result := C.flow_pattern_matcher_find_all(pm.handle, cText, C.size_t(len(text)), &matches, &numMatches)
	if result != 0 || numMatches == 0 {
		return nil
	}
	defer C.flow_pattern_matcher_free_matches(matches)

	goMatches := make([]Match, int(numMatches))
	matchSlice := unsafe.Slice(matches, int(numMatches))

	for i, m := range matchSlice {
		goMatches[i] = Match{
			Start:      int(m.start),
			End:        int(m.end),
			PatternID:  int(m.pattern_id),
			Confidence: float32(m.confidence),
		}
	}

	return goMatches
}

func (pm *PatternMatcher) Close() {
	if pm.handle != nil {
		C.flow_pattern_matcher_destroy(pm.handle)
		pm.handle = nil
	}
}

type AhoCorasick struct {
	handle C.FlowAhoCorasickHandle
}

func NewAhoCorasick() *AhoCorasick {
	ac := &AhoCorasick{
		handle: C.flow_aho_corasick_create(),
	}
	runtime.SetFinalizer(ac, (*AhoCorasick).Close)
	return ac
}

func (ac *AhoCorasick) AddPattern(pattern string, id int) {
	cPattern := C.CString(pattern)
	defer C.free(unsafe.Pointer(cPattern))

	C.flow_aho_corasick_add_pattern(ac.handle, cPattern, C.size_t(id))
}

func (ac *AhoCorasick) Build() {
	C.flow_aho_corasick_build(ac.handle)
}

func (ac *AhoCorasick) Search(text string) []Match {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	var matches *C.FlowMatch
	var numMatches C.size_t

	result := C.flow_aho_corasick_search(ac.handle, cText, C.size_t(len(text)), &matches, &numMatches)
	if result != 0 || numMatches == 0 {
		return nil
	}
	defer C.flow_pattern_matcher_free_matches(matches)

	goMatches := make([]Match, int(numMatches))
	matchSlice := unsafe.Slice(matches, int(numMatches))

	for i, m := range matchSlice {
		goMatches[i] = Match{
			Start:      int(m.start),
			End:        int(m.end),
			PatternID:  int(m.pattern_id),
			Confidence: float32(m.confidence),
		}
	}

	return goMatches
}

func (ac *AhoCorasick) Close() {
	if ac.handle != nil {
		C.flow_aho_corasick_destroy(ac.handle)
		ac.handle = nil
	}
}

type Tokenizer struct {
	handle C.FlowTokenizerHandle
}

func NewTokenizer() *Tokenizer {
	t := &Tokenizer{
		handle: C.flow_tokenizer_create(),
	}
	runtime.SetFinalizer(t, (*Tokenizer).Close)
	return t
}

func (t *Tokenizer) Tokenize(text string) []Token {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	var tokens *C.FlowToken
	var numTokens C.size_t

	result := C.flow_tokenizer_tokenize(t.handle, cText, C.size_t(len(text)), &tokens, &numTokens)
	if result != 0 || numTokens == 0 {
		return nil
	}
	defer C.flow_tokenizer_free_tokens(tokens, numTokens)

	goTokens := make([]Token, int(numTokens))
	tokenSlice := unsafe.Slice(tokens, int(numTokens))

	for i, tok := range tokenSlice {
		goTokens[i] = Token{
			Text:  C.GoString(tok.text),
			Type:  int(tok._type),
			Start: int(tok.start),
			End:   int(tok.end),
		}
	}

	return goTokens
}

func (t *Tokenizer) Close() {
	if t.handle != nil {
		C.flow_tokenizer_destroy(t.handle)
		t.handle = nil
	}
}

type EntityMatcher struct {
	handle C.FlowEntityMatcherHandle
}

func NewEntityMatcher() *EntityMatcher {
	em := &EntityMatcher{
		handle: C.flow_entity_matcher_create(),
	}
	runtime.SetFinalizer(em, (*EntityMatcher).Close)
	return em
}

func (em *EntityMatcher) AddDatePatterns() {
	C.flow_entity_matcher_add_date_patterns(em.handle)
}

func (em *EntityMatcher) AddAmountPatterns() {
	C.flow_entity_matcher_add_amount_patterns(em.handle)
}

func (em *EntityMatcher) AddKeywords(entityType EntityType, keywords []string) {
	if len(keywords) == 0 {
		return
	}

	cKeywords := make([]*C.char, len(keywords))
	for i, kw := range keywords {
		cKeywords[i] = C.CString(kw)
	}
	defer func() {
		for _, ckw := range cKeywords {
			C.free(unsafe.Pointer(ckw))
		}
	}()

	C.flow_entity_matcher_add_keywords(
		em.handle,
		C.FlowEntityType(entityType),
		(**C.char)(unsafe.Pointer(&cKeywords[0])),
		C.size_t(len(keywords)),
	)
}

func (em *EntityMatcher) Extract(text string) []Entity {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	var entities *C.FlowEntity
	var numEntities C.size_t

	result := C.flow_entity_matcher_extract(em.handle, cText, C.size_t(len(text)), &entities, &numEntities)
	if result != 0 || numEntities == 0 {
		return nil
	}
	defer C.flow_entity_matcher_free_entities(entities, numEntities)

	return convertEntities(entities, numEntities)
}

func (em *EntityMatcher) ExtractType(text string, entityType EntityType) []Entity {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	var entities *C.FlowEntity
	var numEntities C.size_t

	result := C.flow_entity_matcher_extract_type(
		em.handle,
		cText,
		C.size_t(len(text)),
		C.FlowEntityType(entityType),
		&entities,
		&numEntities,
	)
	if result != 0 || numEntities == 0 {
		return nil
	}
	defer C.flow_entity_matcher_free_entities(entities, numEntities)

	return convertEntities(entities, numEntities)
}

func (em *EntityMatcher) Close() {
	if em.handle != nil {
		C.flow_entity_matcher_destroy(em.handle)
		em.handle = nil
	}
}

func ExtractAllParallel(text string) []Entity {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	var entities *C.FlowEntity
	var numEntities C.size_t

	result := C.flow_extract_all_parallel(cText, C.size_t(len(text)), &entities, &numEntities)
	if result != 0 || numEntities == 0 {
		return nil
	}
	defer C.flow_entity_matcher_free_entities(entities, numEntities)

	return convertEntities(entities, numEntities)
}

func convertEntities(entities *C.FlowEntity, numEntities C.size_t) []Entity {
	goEntities := make([]Entity, int(numEntities))
	entitySlice := unsafe.Slice(entities, int(numEntities))

	for i, e := range entitySlice {
		goEntities[i] = Entity{
			Value:      C.GoString(e.value),
			Type:       EntityType(e._type),
			Start:      int(e.start),
			End:        int(e.end),
			Confidence: float32(e.confidence),
		}
	}

	return goEntities
}

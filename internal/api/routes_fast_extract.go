package api

import (
	"net/http"

	"flow/internal/native"
	"flow/pkg/logger"

	"github.com/gin-gonic/gin"
)

type FastExtractRequest struct {
	Text      string   `json:"text" binding:"required"`
	Types     []string `json:"types,omitempty"`
	Threshold float32  `json:"threshold,omitempty"`
}

type FastEntity struct {
	Value      string            `json:"value"`
	Type       string            `json:"type"`
	Start      int               `json:"start"`
	End        int               `json:"end"`
	Confidence float32           `json:"confidence"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type FastExtractResponse struct {
	Entities    []FastEntity `json:"entities"`
	ProcessedMs float64      `json:"processed_ms"`
	Engine      string       `json:"engine"`
}

func RegisterFastExtractRoutes(rg *gin.RouterGroup, log *logger.Logger) {
	handler := &fastExtractHandler{
		log: log.WithComponent("fast-extract"),
	}

	f := rg.Group("/fast")
	{
		f.POST("/extract", handler.extract)
		f.POST("/extract/dates", handler.extractDates)
		f.POST("/extract/amounts", handler.extractAmounts)
		f.POST("/extract/emails", handler.extractEmails)
		f.POST("/tokenize", handler.tokenize)
		f.POST("/classify", handler.classify)
	}
}

type fastExtractHandler struct {
	log *logger.Logger
}

func (h *fastExtractHandler) extract(c *gin.Context) {
	var req FastExtractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	em := native.NewEntityMatcher()
	defer em.Close()

	em.AddDatePatterns()
	em.AddAmountPatterns()

	entities := em.Extract(req.Text)

	threshold := req.Threshold
	if threshold == 0 {
		threshold = 0.5
	}

	var results []FastEntity
	for _, e := range entities {
		if e.Confidence >= threshold {
			results = append(results, FastEntity{
				Value:      e.Value,
				Type:       entityTypeToString(e.Type),
				Start:      e.Start,
				End:        e.End,
				Confidence: e.Confidence,
			})
		}
	}

	c.JSON(http.StatusOK, FastExtractResponse{
		Entities: results,
		Engine:   "native-go",
	})
}

func (h *fastExtractHandler) extractDates(c *gin.Context) {
	var req FastExtractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	em := native.NewEntityMatcher()
	defer em.Close()

	em.AddDatePatterns()

	entities := em.ExtractType(req.Text, native.EntityTypeDate)

	var results []FastEntity
	for _, e := range entities {
		results = append(results, FastEntity{
			Value:      e.Value,
			Type:       "date",
			Start:      e.Start,
			End:        e.End,
			Confidence: e.Confidence,
		})
	}

	c.JSON(http.StatusOK, FastExtractResponse{
		Entities: results,
		Engine:   "native-go",
	})
}

func (h *fastExtractHandler) extractAmounts(c *gin.Context) {
	var req FastExtractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	em := native.NewEntityMatcher()
	defer em.Close()

	em.AddAmountPatterns()

	entities := em.ExtractType(req.Text, native.EntityTypeAmount)

	var results []FastEntity
	for _, e := range entities {
		results = append(results, FastEntity{
			Value:      e.Value,
			Type:       "amount",
			Start:      e.Start,
			End:        e.End,
			Confidence: e.Confidence,
		})
	}

	c.JSON(http.StatusOK, FastExtractResponse{
		Entities: results,
		Engine:   "native-go",
	})
}

func (h *fastExtractHandler) extractEmails(c *gin.Context) {
	var req FastExtractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	em := native.NewEntityMatcher()
	defer em.Close()

	entities := em.ExtractType(req.Text, native.EntityTypeEmail)

	var results []FastEntity
	for _, e := range entities {
		results = append(results, FastEntity{
			Value:      e.Value,
			Type:       "email",
			Start:      e.Start,
			End:        e.End,
			Confidence: e.Confidence,
		})
	}

	c.JSON(http.StatusOK, FastExtractResponse{
		Entities: results,
		Engine:   "native-go",
	})
}

func (h *fastExtractHandler) tokenize(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tok := native.NewTokenizer()
	defer tok.Close()

	tokens := tok.Tokenize(req.Text)

	type TokenResponse struct {
		Text  string `json:"text"`
		Type  int    `json:"type"`
		Start int    `json:"start"`
		End   int    `json:"end"`
	}

	var results []TokenResponse
	for _, t := range tokens {
		results = append(results, TokenResponse{
			Text:  t.Text,
			Type:  t.Type,
			Start: t.Start,
			End:   t.End,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tokens": results,
		"count":  len(results),
		"engine": "native-go",
	})
}

func (h *fastExtractHandler) classify(c *gin.Context) {
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pm := native.NewPatternMatcher()
	defer pm.Close()

	dateKeywords := []string{"when", "date", "time", "day", "month", "year", "deadline", "schedule"}
	personKeywords := []string{"who", "person", "name", "employee", "manager", "ceo", "sent", "from", "to"}
	orgKeywords := []string{"company", "organization", "corp", "inc", "firm", "bank", "agency"}
	amountKeywords := []string{"amount", "price", "cost", "dollar", "million", "total", "payment"}

	for i, kw := range dateKeywords {
		pm.AddPattern(kw, 100+i, 0.2)
	}
	for i, kw := range personKeywords {
		pm.AddPattern(kw, 200+i, 0.2)
	}
	for i, kw := range orgKeywords {
		pm.AddPattern(kw, 300+i, 0.2)
	}
	for i, kw := range amountKeywords {
		pm.AddPattern(kw, 400+i, 0.2)
	}

	matches := pm.FindAll(req.Text)

	scores := map[string]float32{
		"date":   0,
		"person": 0,
		"org":    0,
		"amount": 0,
	}

	for _, m := range matches {
		switch {
		case m.PatternID >= 100 && m.PatternID < 200:
			scores["date"] += m.Confidence
		case m.PatternID >= 200 && m.PatternID < 300:
			scores["person"] += m.Confidence
		case m.PatternID >= 300 && m.PatternID < 400:
			scores["org"] += m.Confidence
		case m.PatternID >= 400 && m.PatternID < 500:
			scores["amount"] += m.Confidence
		}
	}

	for k := range scores {
		if scores[k] > 1.0 {
			scores[k] = 1.0
		}
	}

	var primaryType string
	var maxScore float32
	for t, s := range scores {
		if s > maxScore {
			maxScore = s
			primaryType = t
		}
	}

	if maxScore < 0.2 {
		primaryType = "general"
	}

	var workerTypes []string
	for t, s := range scores {
		if s >= 0.2 {
			workerTypes = append(workerTypes, t)
		}
	}

	if len(workerTypes) == 0 {
		workerTypes = []string{"date", "person", "org", "amount"}
	}

	c.JSON(http.StatusOK, gin.H{
		"primary_type": primaryType,
		"scores":       scores,
		"worker_types": workerTypes,
		"engine":       "native-go",
	})
}

func entityTypeToString(t native.EntityType) string {
	switch t {
	case native.EntityTypeDate:
		return "date"
	case native.EntityTypePerson:
		return "person"
	case native.EntityTypeOrganization:
		return "organization"
	case native.EntityTypeAmount:
		return "amount"
	case native.EntityTypeEmail:
		return "email"
	default:
		return "unknown"
	}
}

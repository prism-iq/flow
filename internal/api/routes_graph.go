package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"flow/pkg/logger"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type GraphHandler struct {
	db  *sql.DB
	log *logger.Logger
}

type CreateNodeRequest struct {
	Label      string                 `json:"label" binding:"required"`
	Properties map[string]interface{} `json:"properties" binding:"required"`
}

type CreateEdgeRequest struct {
	FromNodeID string                 `json:"from_node_id" binding:"required"`
	ToNodeID   string                 `json:"to_node_id" binding:"required"`
	Label      string                 `json:"label" binding:"required"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

type StoreEntitiesRequest struct {
	Entities   []map[string]interface{} `json:"entities" binding:"required"`
	SourceText string                   `json:"source_text,omitempty"`
}

func RegisterGraphRoutes(rg *gin.RouterGroup, log *logger.Logger) {
	connStr := "host=localhost port=5432 user=postgres dbname=flow sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to database")
		return
	}

	if err := db.Ping(); err != nil {
		log.Error().Err(err).Msg("Failed to ping database")
		return
	}

	handler := &GraphHandler{
		db:  db,
		log: log.WithComponent("graph-handler"),
	}

	g := rg.Group("/graph")
	{
		g.POST("/node", handler.createNode)
		g.POST("/edge", handler.createEdge)
		g.POST("/store-entities", handler.storeEntities)
		g.GET("/nodes", handler.listNodes)
		g.GET("/nodes/:id/connections", handler.getConnections)
		g.GET("/search", handler.searchNodes)
	}
}

func (h *GraphHandler) createNode(c *gin.Context) {
	var req CreateNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	propsJSON, _ := json.Marshal(req.Properties)

	var nodeID string
	err := h.db.QueryRow(
		"SELECT create_graph_node($1, $2)",
		req.Label, string(propsJSON),
	).Scan(&nodeID)

	if err != nil {
		h.log.Error().Err(err).Msg("Failed to create node")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"node_id": nodeID,
		"label":   req.Label,
	})
}

func (h *GraphHandler) createEdge(c *gin.Context) {
	var req CreateEdgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	props := req.Properties
	if props == nil {
		props = make(map[string]interface{})
	}
	propsJSON, _ := json.Marshal(props)

	var edgeID string
	err := h.db.QueryRow(
		"SELECT create_graph_edge($1::uuid, $2::uuid, $3, $4)",
		req.FromNodeID, req.ToNodeID, req.Label, string(propsJSON),
	).Scan(&edgeID)

	if err != nil {
		h.log.Error().Err(err).Msg("Failed to create edge")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"edge_id": edgeID,
		"label":   req.Label,
	})
}

func (h *GraphHandler) storeEntities(c *gin.Context) {
	var req StoreEntitiesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var nodeIDs []string

	for _, entity := range req.Entities {
		entityType, _ := entity["type"].(string)
		if entityType == "" {
			entityType, _ = entity["source"].(string)
		}

		label := mapEntityTypeToLabel(entityType)

		propsJSON, _ := json.Marshal(entity)

		var nodeID string
		err := h.db.QueryRow(
			"SELECT create_graph_node($1, $2)",
			label, string(propsJSON),
		).Scan(&nodeID)

		if err != nil {
			h.log.Warn().Err(err).Str("type", entityType).Msg("Failed to create node for entity")
			continue
		}

		nodeIDs = append(nodeIDs, nodeID)

		_, err = h.db.Exec(`
			INSERT INTO entities (entity_type, value, confidence, source_text, graph_node_id, metadata)
			VALUES ($1, $2, $3, $4, $5::uuid, $6)
		`,
			entityType,
			getEntityValue(entity),
			getEntityConfidence(entity),
			req.SourceText,
			nodeID,
			string(propsJSON),
		)

		if err != nil {
			h.log.Warn().Err(err).Msg("Failed to insert entity record")
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"stored":   len(nodeIDs),
		"node_ids": nodeIDs,
	})
}

func (h *GraphHandler) listNodes(c *gin.Context) {
	label := c.Query("label")
	limit := c.DefaultQuery("limit", "50")

	var rows *sql.Rows
	var err error

	if label != "" {
		rows, err = h.db.Query(
			"SELECT node_id, label, properties, created_at FROM find_nodes($1, NULL, $2::int)",
			label, limit,
		)
	} else {
		rows, err = h.db.Query(
			"SELECT node_id, label, properties, created_at FROM find_nodes(NULL, NULL, $1::int)",
			limit,
		)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var nodes []map[string]interface{}
	for rows.Next() {
		var nodeID, nodeLabel, propsStr string
		var createdAt string
		if err := rows.Scan(&nodeID, &nodeLabel, &propsStr, &createdAt); err != nil {
			continue
		}

		var props map[string]interface{}
		json.Unmarshal([]byte(propsStr), &props)

		nodes = append(nodes, map[string]interface{}{
			"id":         nodeID,
			"label":      nodeLabel,
			"properties": props,
			"created_at": createdAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"nodes": nodes, "count": len(nodes)})
}

func (h *GraphHandler) getConnections(c *gin.Context) {
	nodeID := c.Param("id")
	edgeLabel := c.Query("edge_label")
	direction := c.DefaultQuery("direction", "both")

	var rows *sql.Rows
	var err error

	if edgeLabel != "" {
		rows, err = h.db.Query(
			"SELECT node_id, node_label, node_properties, edge_label, edge_properties, direction FROM find_connected_nodes($1::uuid, $2, $3)",
			nodeID, edgeLabel, direction,
		)
	} else {
		rows, err = h.db.Query(
			"SELECT node_id, node_label, node_properties, edge_label, edge_properties, direction FROM find_connected_nodes($1::uuid, NULL, $2)",
			nodeID, direction,
		)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var connections []map[string]interface{}
	for rows.Next() {
		var connNodeID, connNodeLabel, nodePropsStr, connEdgeLabel, edgePropsStr, dir string
		if err := rows.Scan(&connNodeID, &connNodeLabel, &nodePropsStr, &connEdgeLabel, &edgePropsStr, &dir); err != nil {
			continue
		}

		var nodeProps, edgeProps map[string]interface{}
		json.Unmarshal([]byte(nodePropsStr), &nodeProps)
		json.Unmarshal([]byte(edgePropsStr), &edgeProps)

		connections = append(connections, map[string]interface{}{
			"node_id":         connNodeID,
			"node_label":      connNodeLabel,
			"node_properties": nodeProps,
			"edge_label":      connEdgeLabel,
			"edge_properties": edgeProps,
			"direction":       dir,
		})
	}

	c.JSON(http.StatusOK, gin.H{"connections": connections, "count": len(connections)})
}

func (h *GraphHandler) searchNodes(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	label := c.Query("label")
	limit := c.DefaultQuery("limit", "20")

	var rows *sql.Rows
	var err error

	if label != "" {
		rows, err = h.db.Query(
			"SELECT node_id, label, properties, relevance FROM search_nodes($1, $2, $3::int)",
			query, label, limit,
		)
	} else {
		rows, err = h.db.Query(
			"SELECT node_id, label, properties, relevance FROM search_nodes($1, NULL, $2::int)",
			query, limit,
		)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var nodeID, nodeLabel, propsStr string
		var relevance float64
		if err := rows.Scan(&nodeID, &nodeLabel, &propsStr, &relevance); err != nil {
			continue
		}

		var props map[string]interface{}
		json.Unmarshal([]byte(propsStr), &props)

		results = append(results, map[string]interface{}{
			"id":         nodeID,
			"label":      nodeLabel,
			"properties": props,
			"relevance":  relevance,
		})
	}

	c.JSON(http.StatusOK, gin.H{"results": results, "count": len(results)})
}

func mapEntityTypeToLabel(entityType string) string {
	switch entityType {
	case "date":
		return "Date"
	case "person":
		return "Person"
	case "org", "organization":
		return "Organization"
	case "amount":
		return "Amount"
	case "email":
		return "Email"
	default:
		return "Entity"
	}
}

func getEntityValue(entity map[string]interface{}) string {
	if v, ok := entity["value"]; ok {
		switch val := v.(type) {
		case string:
			return val
		default:
			b, _ := json.Marshal(val)
			return string(b)
		}
	}
	return ""
}

func getEntityConfidence(entity map[string]interface{}) float64 {
	if v, ok := entity["confidence"]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		}
	}
	return 0.8
}

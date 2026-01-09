package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"flow/pkg/logger"
)

type GraphDB struct {
	db        *sql.DB
	graphName string
	log       *logger.Logger
}

type Node struct {
	ID         int64                  `json:"id"`
	Label      string                 `json:"label"`
	Properties map[string]interface{} `json:"properties"`
}

type Edge struct {
	ID         int64                  `json:"id"`
	Label      string                 `json:"label"`
	StartID    int64                  `json:"start_id"`
	EndID      int64                  `json:"end_id"`
	Properties map[string]interface{} `json:"properties"`
}

type GraphPath struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

func NewGraphDB(db *sql.DB, graphName string, log *logger.Logger) *GraphDB {
	return &GraphDB{
		db:        db,
		graphName: graphName,
		log:       log.WithComponent("graph-db"),
	}
}

func (g *GraphDB) CreateNode(ctx context.Context, label string, properties map[string]interface{}) (*Node, error) {
	propsJSON, err := json.Marshal(properties)
	if err != nil {
		return nil, fmt.Errorf("marshal properties: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT * FROM cypher('%s', $$
			CREATE (n:%s %s)
			RETURN id(n), '%s', properties(n)
		$$) AS (id agtype, label agtype, props agtype)
	`, g.graphName, label, string(propsJSON), label)

	var nodeID int64
	var nodeLabel, propsStr string

	err = g.db.QueryRowContext(ctx, query).Scan(&nodeID, &nodeLabel, &propsStr)
	if err != nil {
		return nil, fmt.Errorf("create node: %w", err)
	}

	return &Node{
		ID:         nodeID,
		Label:      label,
		Properties: properties,
	}, nil
}

func (g *GraphDB) CreateEdge(ctx context.Context, startID, endID int64, label string, properties map[string]interface{}) (*Edge, error) {
	propsJSON := "{}"
	if properties != nil {
		b, _ := json.Marshal(properties)
		propsJSON = string(b)
	}

	query := fmt.Sprintf(`
		SELECT * FROM cypher('%s', $$
			MATCH (a), (b)
			WHERE id(a) = %d AND id(b) = %d
			CREATE (a)-[r:%s %s]->(b)
			RETURN id(r)
		$$) AS (id agtype)
	`, g.graphName, startID, endID, label, propsJSON)

	var edgeID int64
	err := g.db.QueryRowContext(ctx, query).Scan(&edgeID)
	if err != nil {
		return nil, fmt.Errorf("create edge: %w", err)
	}

	return &Edge{
		ID:         edgeID,
		Label:      label,
		StartID:    startID,
		EndID:      endID,
		Properties: properties,
	}, nil
}

func (g *GraphDB) FindNodeByProperty(ctx context.Context, label, propertyName string, propertyValue interface{}) (*Node, error) {
	query := fmt.Sprintf(`
		SELECT * FROM cypher('%s', $$
			MATCH (n:%s {%s: '%v'})
			RETURN id(n), properties(n)
			LIMIT 1
		$$) AS (id agtype, props agtype)
	`, g.graphName, label, propertyName, propertyValue)

	var nodeID int64
	var propsStr string

	err := g.db.QueryRowContext(ctx, query).Scan(&nodeID, &propsStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find node: %w", err)
	}

	var props map[string]interface{}
	json.Unmarshal([]byte(propsStr), &props)

	return &Node{
		ID:         nodeID,
		Label:      label,
		Properties: props,
	}, nil
}

func (g *GraphDB) GetNodeNeighbors(ctx context.Context, nodeID int64, edgeLabel string, limit int) ([]Node, error) {
	edgeFilter := ""
	if edgeLabel != "" {
		edgeFilter = fmt.Sprintf(":%s", edgeLabel)
	}

	query := fmt.Sprintf(`
		SELECT * FROM cypher('%s', $$
			MATCH (n)-[%s]-(neighbor)
			WHERE id(n) = %d
			RETURN id(neighbor), labels(neighbor)[0], properties(neighbor)
			LIMIT %d
		$$) AS (id agtype, label agtype, props agtype)
	`, g.graphName, edgeFilter, nodeID, limit)

	rows, err := g.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get neighbors: %w", err)
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var id int64
		var label, propsStr string

		if err := rows.Scan(&id, &label, &propsStr); err != nil {
			continue
		}

		var props map[string]interface{}
		json.Unmarshal([]byte(propsStr), &props)

		nodes = append(nodes, Node{
			ID:         id,
			Label:      label,
			Properties: props,
		})
	}

	return nodes, nil
}

func (g *GraphDB) ShortestPath(ctx context.Context, startID, endID int64, maxDepth int) (*GraphPath, error) {
	query := fmt.Sprintf(`
		SELECT * FROM cypher('%s', $$
			MATCH path = shortestPath((a)-[*1..%d]-(b))
			WHERE id(a) = %d AND id(b) = %d
			RETURN nodes(path), relationships(path)
		$$) AS (nodes agtype, edges agtype)
	`, g.graphName, maxDepth, startID, endID)

	var nodesStr, edgesStr string
	err := g.db.QueryRowContext(ctx, query).Scan(&nodesStr, &edgesStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("shortest path: %w", err)
	}

	return &GraphPath{}, nil
}

func (g *GraphDB) ExecuteCypher(ctx context.Context, cypherQuery string) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`
		SELECT * FROM cypher('%s', $$%s$$) AS result(data agtype)
	`, g.graphName, cypherQuery)

	rows, err := g.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("execute cypher: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var dataStr string
		if err := rows.Scan(&dataStr); err != nil {
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			data = map[string]interface{}{"raw": dataStr}
		}
		results = append(results, data)
	}

	return results, nil
}

func (g *GraphDB) CountNodes(ctx context.Context, label string) (int64, error) {
	labelFilter := ""
	if label != "" {
		labelFilter = fmt.Sprintf(":%s", label)
	}

	query := fmt.Sprintf(`
		SELECT * FROM cypher('%s', $$
			MATCH (n%s)
			RETURN count(n)
		$$) AS (count agtype)
	`, g.graphName, labelFilter)

	var count int64
	err := g.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count nodes: %w", err)
	}

	return count, nil
}

func (g *GraphDB) CountEdges(ctx context.Context, label string) (int64, error) {
	labelFilter := ""
	if label != "" {
		labelFilter = fmt.Sprintf(":%s", label)
	}

	query := fmt.Sprintf(`
		SELECT * FROM cypher('%s', $$
			MATCH ()-[r%s]->()
			RETURN count(r)
		$$) AS (count agtype)
	`, g.graphName, labelFilter)

	var count int64
	err := g.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count edges: %w", err)
	}

	return count, nil
}

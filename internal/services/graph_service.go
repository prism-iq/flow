package services

import (
	"context"
	"database/sql"
	"fmt"

	"flow/pkg/logger"
)

type GraphService struct {
	db  *sql.DB
	log *logger.Logger
}

type Node struct {
	ID         string            `json:"id"`
	Label      string            `json:"label"`
	Properties map[string]string `json:"properties"`
}

type Edge struct {
	ID         string            `json:"id"`
	Label      string            `json:"label"`
	Source     string            `json:"source"`
	Target     string            `json:"target"`
	Properties map[string]string `json:"properties"`
}

type GraphResult struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

func NewGraphService(db *sql.DB, log *logger.Logger) *GraphService {
	return &GraphService{
		db:  db,
		log: log.WithComponent("graph-service"),
	}
}

func (s *GraphService) InitGraph(ctx context.Context, graphName string) error {
	query := fmt.Sprintf(`
		SELECT * FROM ag_catalog.create_graph('%s');
	`, graphName)

	_, err := s.db.ExecContext(ctx, query)
	if err != nil {
		s.log.Warn().Err(err).Str("graph", graphName).Msg("Graph may already exist")
	}
	return nil
}

func (s *GraphService) CreateNode(ctx context.Context, graphName, label string, properties map[string]interface{}) (*Node, error) {
	propsJSON, err := mapToAGE(properties)
	if err != nil {
		return nil, fmt.Errorf("serialize properties: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT * FROM cypher('%s', $$
			CREATE (n:%s %s)
			RETURN id(n), properties(n)
		$$) AS (id agtype, props agtype);
	`, graphName, label, propsJSON)

	var nodeID, props string
	err = s.db.QueryRowContext(ctx, query).Scan(&nodeID, &props)
	if err != nil {
		return nil, fmt.Errorf("create node: %w", err)
	}

	return &Node{
		ID:    nodeID,
		Label: label,
	}, nil
}

func (s *GraphService) CreateEdge(ctx context.Context, graphName, sourceID, targetID, label string, properties map[string]interface{}) (*Edge, error) {
	propsJSON, err := mapToAGE(properties)
	if err != nil {
		return nil, fmt.Errorf("serialize properties: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT * FROM cypher('%s', $$
			MATCH (a), (b)
			WHERE id(a) = %s AND id(b) = %s
			CREATE (a)-[r:%s %s]->(b)
			RETURN id(r)
		$$) AS (id agtype);
	`, graphName, sourceID, targetID, label, propsJSON)

	var edgeID string
	err = s.db.QueryRowContext(ctx, query).Scan(&edgeID)
	if err != nil {
		return nil, fmt.Errorf("create edge: %w", err)
	}

	return &Edge{
		ID:     edgeID,
		Label:  label,
		Source: sourceID,
		Target: targetID,
	}, nil
}

func (s *GraphService) QueryCypher(ctx context.Context, graphName, cypherQuery string) (*GraphResult, error) {
	query := fmt.Sprintf(`
		SELECT * FROM cypher('%s', $$%s$$) AS result(data agtype);
	`, graphName, cypherQuery)

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("execute cypher: %w", err)
	}
	defer rows.Close()

	result := &GraphResult{
		Nodes: make([]Node, 0),
		Edges: make([]Edge, 0),
	}

	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			continue
		}
		s.log.Debug().Str("data", data).Msg("Graph query result")
	}

	return result, nil
}

func (s *GraphService) FindPath(ctx context.Context, graphName, startID, endID string, maxDepth int) (*GraphResult, error) {
	cypherQuery := fmt.Sprintf(`
		MATCH path = shortestPath((a)-[*1..%d]-(b))
		WHERE id(a) = %s AND id(b) = %s
		RETURN path
	`, maxDepth, startID, endID)

	return s.QueryCypher(ctx, graphName, cypherQuery)
}

func mapToAGE(m map[string]interface{}) (string, error) {
	if len(m) == 0 {
		return "{}", nil
	}

	result := "{"
	first := true
	for k, v := range m {
		if !first {
			result += ", "
		}
		result += fmt.Sprintf("%s: '%v'", k, v)
		first = false
	}
	result += "}"
	return result, nil
}

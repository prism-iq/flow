-- Flow Database Schema (Simplified - without AGE)
-- PostgreSQL with JSONB for graph-like storage

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- ============================================
-- RELATIONAL TABLES
-- ============================================

-- Conversations table
CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL DEFAULT 'New Conversation',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Messages table
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID REFERENCES conversations(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    token_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_created ON messages(created_at);

-- ============================================
-- GRAPH-LIKE STORAGE (JSONB based)
-- ============================================

-- Nodes table (vertices)
CREATE TABLE IF NOT EXISTS graph_nodes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    label VARCHAR(50) NOT NULL,
    properties JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_graph_nodes_label ON graph_nodes(label);
CREATE INDEX IF NOT EXISTS idx_graph_nodes_properties ON graph_nodes USING gin(properties);

-- Edges table (relationships)
CREATE TABLE IF NOT EXISTS graph_edges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    from_node_id UUID REFERENCES graph_nodes(id) ON DELETE CASCADE,
    to_node_id UUID REFERENCES graph_nodes(id) ON DELETE CASCADE,
    label VARCHAR(50) NOT NULL,
    properties JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_graph_edges_from ON graph_edges(from_node_id);
CREATE INDEX IF NOT EXISTS idx_graph_edges_to ON graph_edges(to_node_id);
CREATE INDEX IF NOT EXISTS idx_graph_edges_label ON graph_edges(label);

-- Entities extracted table
CREATE TABLE IF NOT EXISTS entities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(50) NOT NULL,
    value TEXT NOT NULL,
    normalized_value TEXT,
    confidence FLOAT DEFAULT 1.0,
    source_text TEXT,
    graph_node_id UUID REFERENCES graph_nodes(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_entities_type ON entities(entity_type);
CREATE INDEX IF NOT EXISTS idx_entities_value ON entities USING gin(value gin_trgm_ops);

-- Graph operations log (for replay/audit)
CREATE TABLE IF NOT EXISTS graph_operations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    operation VARCHAR(50) NOT NULL,
    node_label VARCHAR(50),
    edge_label VARCHAR(50),
    from_node_id UUID,
    to_node_id UUID,
    properties JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    source_request_id UUID
);

CREATE INDEX IF NOT EXISTS idx_graph_ops_created ON graph_operations(created_at);

-- ============================================
-- FUNCTIONS
-- ============================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for conversations
DROP TRIGGER IF EXISTS conversations_updated_at ON conversations;
CREATE TRIGGER conversations_updated_at
    BEFORE UPDATE ON conversations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Trigger for graph_nodes
DROP TRIGGER IF EXISTS graph_nodes_updated_at ON graph_nodes;
CREATE TRIGGER graph_nodes_updated_at
    BEFORE UPDATE ON graph_nodes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Function to create a node
CREATE OR REPLACE FUNCTION create_graph_node(
    p_label VARCHAR,
    p_properties JSONB
)
RETURNS UUID AS $$
DECLARE
    v_node_id UUID;
BEGIN
    INSERT INTO graph_nodes (label, properties)
    VALUES (p_label, p_properties)
    RETURNING id INTO v_node_id;

    INSERT INTO graph_operations (operation, node_label, properties)
    VALUES ('CREATE_NODE', p_label, p_properties);

    RETURN v_node_id;
END;
$$ LANGUAGE plpgsql;

-- Function to create an edge
CREATE OR REPLACE FUNCTION create_graph_edge(
    p_from_node_id UUID,
    p_to_node_id UUID,
    p_label VARCHAR,
    p_properties JSONB DEFAULT '{}'::jsonb
)
RETURNS UUID AS $$
DECLARE
    v_edge_id UUID;
BEGIN
    INSERT INTO graph_edges (from_node_id, to_node_id, label, properties)
    VALUES (p_from_node_id, p_to_node_id, p_label, p_properties)
    RETURNING id INTO v_edge_id;

    INSERT INTO graph_operations (operation, edge_label, from_node_id, to_node_id, properties)
    VALUES ('CREATE_EDGE', p_label, p_from_node_id, p_to_node_id, p_properties);

    RETURN v_edge_id;
END;
$$ LANGUAGE plpgsql;

-- Function to find nodes by label and properties
CREATE OR REPLACE FUNCTION find_nodes(
    p_label VARCHAR DEFAULT NULL,
    p_properties JSONB DEFAULT NULL,
    p_limit INTEGER DEFAULT 100
)
RETURNS TABLE (
    node_id UUID,
    label VARCHAR,
    properties JSONB,
    created_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT n.id, n.label, n.properties, n.created_at
    FROM graph_nodes n
    WHERE (p_label IS NULL OR n.label = p_label)
      AND (p_properties IS NULL OR n.properties @> p_properties)
    ORDER BY n.created_at DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- Function to find connected nodes
CREATE OR REPLACE FUNCTION find_connected_nodes(
    p_node_id UUID,
    p_edge_label VARCHAR DEFAULT NULL,
    p_direction VARCHAR DEFAULT 'both'
)
RETURNS TABLE (
    node_id UUID,
    node_label VARCHAR,
    node_properties JSONB,
    edge_label VARCHAR,
    edge_properties JSONB,
    direction VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        n.id,
        n.label,
        n.properties,
        e.label,
        e.properties,
        CASE WHEN e.from_node_id = p_node_id THEN 'outgoing' ELSE 'incoming' END::VARCHAR
    FROM graph_edges e
    JOIN graph_nodes n ON (
        (e.from_node_id = p_node_id AND n.id = e.to_node_id) OR
        (e.to_node_id = p_node_id AND n.id = e.from_node_id AND p_direction IN ('both', 'incoming'))
    )
    WHERE (p_edge_label IS NULL OR e.label = p_edge_label)
      AND (p_direction = 'both' OR
           (p_direction = 'outgoing' AND e.from_node_id = p_node_id) OR
           (p_direction = 'incoming' AND e.to_node_id = p_node_id));
END;
$$ LANGUAGE plpgsql;

-- Function to search nodes by text
CREATE OR REPLACE FUNCTION search_nodes(
    p_search_term VARCHAR,
    p_label VARCHAR DEFAULT NULL,
    p_limit INTEGER DEFAULT 20
)
RETURNS TABLE (
    node_id UUID,
    label VARCHAR,
    properties JSONB,
    relevance FLOAT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        n.id,
        n.label,
        n.properties,
        similarity(n.properties::text, p_search_term)::FLOAT as relevance
    FROM graph_nodes n
    WHERE (p_label IS NULL OR n.label = p_label)
      AND n.properties::text ILIKE '%' || p_search_term || '%'
    ORDER BY relevance DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

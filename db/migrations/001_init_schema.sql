-- Flow Database Schema
-- PostgreSQL + Apache AGE

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS age;

-- Load AGE
LOAD 'age';
SET search_path = ag_catalog, "$user", public;

-- Create graph
SELECT create_graph('flow_graph');

-- ============================================
-- RELATIONAL TABLES (metadata, sessions, etc)
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

CREATE INDEX idx_messages_conversation ON messages(conversation_id);
CREATE INDEX idx_messages_created ON messages(created_at);

-- Documents table (for RAG)
CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    chunk_index INTEGER DEFAULT 0,
    embedding vector(384),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_documents_source ON documents(source);

-- Entities extracted table
CREATE TABLE IF NOT EXISTS entities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(50) NOT NULL,
    value TEXT NOT NULL,
    normalized_value TEXT,
    confidence FLOAT DEFAULT 1.0,
    source_document_id UUID REFERENCES documents(id),
    graph_node_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_entities_type ON entities(entity_type);
CREATE INDEX idx_entities_value ON entities USING gin(value gin_trgm_ops);

-- ============================================
-- GRAPH SCHEMA (Apache AGE)
-- ============================================

-- Create vertex labels
SELECT create_vlabel('flow_graph', 'Person');
SELECT create_vlabel('flow_graph', 'Organization');
SELECT create_vlabel('flow_graph', 'Email');
SELECT create_vlabel('flow_graph', 'Document');
SELECT create_vlabel('flow_graph', 'Event');
SELECT create_vlabel('flow_graph', 'Location');
SELECT create_vlabel('flow_graph', 'Amount');
SELECT create_vlabel('flow_graph', 'Date');

-- Create edge labels
SELECT create_elabel('flow_graph', 'SENT');
SELECT create_elabel('flow_graph', 'RECEIVED');
SELECT create_elabel('flow_graph', 'MENTIONS');
SELECT create_elabel('flow_graph', 'WORKS_FOR');
SELECT create_elabel('flow_graph', 'RELATED_TO');
SELECT create_elabel('flow_graph', 'OCCURRED_AT');
SELECT create_elabel('flow_graph', 'LOCATED_IN');
SELECT create_elabel('flow_graph', 'HAS_AMOUNT');
SELECT create_elabel('flow_graph', 'DATED');

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
CREATE TRIGGER conversations_updated_at
    BEFORE UPDATE ON conversations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Function to search graph by entity
CREATE OR REPLACE FUNCTION search_graph_entity(
    p_entity_type VARCHAR,
    p_search_term VARCHAR,
    p_limit INTEGER DEFAULT 10
)
RETURNS TABLE (
    node_id BIGINT,
    properties JSONB,
    connections BIGINT
) AS $$
BEGIN
    RETURN QUERY EXECUTE format(
        'SELECT * FROM cypher(''flow_graph'', $$
            MATCH (n:%I)
            WHERE n.name =~ ''(?i).*%s.*''
            OPTIONAL MATCH (n)-[r]-()
            RETURN id(n), properties(n), count(r)
            LIMIT %s
        $$) AS (id agtype, props agtype, conn_count agtype)',
        p_entity_type, p_search_term, p_limit
    );
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- INDEXES FOR GRAPH QUERIES
-- ============================================

-- These will be created by AGE automatically, but we ensure proper indexing

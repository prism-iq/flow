-- Email-specific schema for investigation data
-- Handles 14K+ nodes and 13K emails

-- ============================================
-- EMAIL TABLES
-- ============================================

CREATE TABLE IF NOT EXISTS emails (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id VARCHAR(255) UNIQUE,
    subject TEXT,
    sender VARCHAR(255),
    recipients TEXT[],
    cc TEXT[],
    bcc TEXT[],
    body TEXT,
    html_body TEXT,
    sent_date TIMESTAMP WITH TIME ZONE,
    received_date TIMESTAMP WITH TIME ZONE,
    thread_id VARCHAR(255),
    in_reply_to VARCHAR(255),
    attachments JSONB DEFAULT '[]'::jsonb,
    headers JSONB DEFAULT '{}'::jsonb,
    graph_node_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_emails_sender ON emails(sender);
CREATE INDEX idx_emails_subject ON emails USING gin(subject gin_trgm_ops);
CREATE INDEX idx_emails_body ON emails USING gin(body gin_trgm_ops);
CREATE INDEX idx_emails_sent_date ON emails(sent_date);
CREATE INDEX idx_emails_thread ON emails(thread_id);

-- Email participants (denormalized for fast lookup)
CREATE TABLE IF NOT EXISTS email_participants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email_id UUID REFERENCES emails(id) ON DELETE CASCADE,
    participant_email VARCHAR(255) NOT NULL,
    participant_name VARCHAR(255),
    role VARCHAR(20) NOT NULL CHECK (role IN ('from', 'to', 'cc', 'bcc')),
    graph_node_id BIGINT
);

CREATE INDEX idx_participants_email ON email_participants(email_id);
CREATE INDEX idx_participants_address ON email_participants(participant_email);

-- Extracted amounts from emails
CREATE TABLE IF NOT EXISTS extracted_amounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email_id UUID REFERENCES emails(id) ON DELETE CASCADE,
    amount DECIMAL(20, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    context TEXT,
    confidence FLOAT DEFAULT 1.0,
    graph_node_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_amounts_email ON extracted_amounts(email_id);
CREATE INDEX idx_amounts_value ON extracted_amounts(amount);

-- Extracted dates from emails
CREATE TABLE IF NOT EXISTS extracted_dates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email_id UUID REFERENCES emails(id) ON DELETE CASCADE,
    extracted_date TIMESTAMP WITH TIME ZONE NOT NULL,
    date_type VARCHAR(50),
    context TEXT,
    confidence FLOAT DEFAULT 1.0,
    graph_node_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_dates_email ON extracted_dates(email_id);
CREATE INDEX idx_dates_value ON extracted_dates(extracted_date);

-- ============================================
-- GRAPH VERTICES FOR EMAIL DATA
-- ============================================

-- These Cypher queries create the graph structure
-- Run via AGE after data import

-- Example: Create Person nodes from email participants
-- SELECT * FROM cypher('flow_graph', $$
--     CREATE (p:Person {
--         email: 'john@example.com',
--         name: 'John Doe',
--         first_seen: '2024-01-01',
--         email_count: 150
--     })
--     RETURN p
-- $$) AS (p agtype);

-- Example: Create Email node and relationships
-- SELECT * FROM cypher('flow_graph', $$
--     MATCH (sender:Person {email: 'john@example.com'})
--     MATCH (recipient:Person {email: 'jane@example.com'})
--     CREATE (e:Email {
--         message_id: 'msg-123',
--         subject: 'Meeting tomorrow',
--         sent_date: '2024-01-15T10:30:00Z'
--     })
--     CREATE (sender)-[:SENT]->(e)
--     CREATE (e)-[:RECEIVED]->(recipient)
--     RETURN e
-- $$) AS (e agtype);

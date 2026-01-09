-- Hypothesis Engine Schema
-- Stores investigative hypotheses with confidence tracking

-- Investigations (container for hypotheses)
CREATE TABLE IF NOT EXISTS investigations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'active',  -- active, paused, completed, archived
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Hypotheses
CREATE TABLE IF NOT EXISTS hypotheses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investigation_id UUID REFERENCES investigations(id) ON DELETE CASCADE,
    statement TEXT NOT NULL,
    hypothesis_type VARCHAR(50),  -- connection, anomaly, pattern, causation
    status VARCHAR(50) DEFAULT 'pending',  -- pending, testing, supported, refuted, inconclusive
    confidence FLOAT DEFAULT 0.5,  -- 0.0 - 1.0
    relevance FLOAT DEFAULT 0.5,   -- 0.0 - 1.0
    priority FLOAT GENERATED ALWAYS AS (confidence * relevance) STORED,
    test_description TEXT,
    parent_hypothesis_id UUID REFERENCES hypotheses(id),  -- for hypothesis chains
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Evidence linking hypotheses to entities/emails
CREATE TABLE IF NOT EXISTS hypothesis_evidence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hypothesis_id UUID REFERENCES hypotheses(id) ON DELETE CASCADE,
    evidence_type VARCHAR(50) NOT NULL,  -- email, entity, relationship, external
    evidence_id VARCHAR(255),  -- reference to email_id, entity_id, etc.
    evidence_text TEXT,
    supports BOOLEAN DEFAULT true,  -- true = supports, false = refutes
    weight FLOAT DEFAULT 1.0,  -- how strong is this evidence
    extracted_at TIMESTAMP DEFAULT NOW()
);

-- Hypothesis evaluations (Haiku scoring history)
CREATE TABLE IF NOT EXISTS hypothesis_evaluations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hypothesis_id UUID REFERENCES hypotheses(id) ON DELETE CASCADE,
    confidence_score FLOAT,
    relevance_score FLOAT,
    reasoning TEXT,
    model_used VARCHAR(100) DEFAULT 'claude-3-haiku',
    evaluated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_hypotheses_investigation ON hypotheses(investigation_id);
CREATE INDEX IF NOT EXISTS idx_hypotheses_status ON hypotheses(status);
CREATE INDEX IF NOT EXISTS idx_hypotheses_priority ON hypotheses(priority DESC);
CREATE INDEX IF NOT EXISTS idx_evidence_hypothesis ON hypothesis_evidence(hypothesis_id);

-- Function to update hypothesis confidence based on evidence
CREATE OR REPLACE FUNCTION update_hypothesis_confidence(h_id UUID)
RETURNS FLOAT AS $$
DECLARE
    new_confidence FLOAT;
BEGIN
    SELECT 
        CASE 
            WHEN COUNT(*) = 0 THEN 0.5
            ELSE LEAST(1.0, GREATEST(0.0,
                0.5 + (SUM(CASE WHEN supports THEN weight ELSE -weight END) / (COUNT(*) * 2))
            ))
        END INTO new_confidence
    FROM hypothesis_evidence
    WHERE hypothesis_id = h_id;
    
    UPDATE hypotheses SET confidence = new_confidence, updated_at = NOW()
    WHERE id = h_id;
    
    RETURN new_confidence;
END;
$$ LANGUAGE plpgsql;

-- Function to get top hypotheses for investigation
CREATE OR REPLACE FUNCTION get_top_hypotheses(inv_id UUID, limit_count INT DEFAULT 10)
RETURNS TABLE (
    id UUID,
    statement TEXT,
    hypothesis_type VARCHAR(50),
    status VARCHAR(50),
    confidence FLOAT,
    relevance FLOAT,
    priority FLOAT,
    evidence_count BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        h.id,
        h.statement,
        h.hypothesis_type,
        h.status,
        h.confidence,
        h.relevance,
        h.priority,
        COUNT(e.id) as evidence_count
    FROM hypotheses h
    LEFT JOIN hypothesis_evidence e ON h.id = e.hypothesis_id
    WHERE h.investigation_id = inv_id
    GROUP BY h.id
    ORDER BY h.priority DESC, h.created_at DESC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

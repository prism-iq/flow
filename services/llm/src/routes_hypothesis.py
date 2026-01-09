"""
Hypothesis API Routes - REST endpoints for hypothesis-driven investigation.
"""
from typing import Dict, List, Optional, Any
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field

from .hypothesis import HypothesisEngine

router = APIRouter(prefix="/hypothesis", tags=["hypothesis"])

# Global engine instance
engine = HypothesisEngine()


# Request/Response Models
class CreateInvestigationRequest(BaseModel):
    name: str
    description: str = ""
    entities: List[Dict[str, Any]] = Field(default_factory=list)


class AddEntitiesRequest(BaseModel):
    entities: List[Dict[str, Any]]
    auto_generate: bool = True


class AddEvidenceRequest(BaseModel):
    evidence_type: str  # email, entity, relationship, external
    evidence_id: Optional[str] = None
    text: str
    supports: bool = True
    weight: float = 1.0


class TestHypothesisRequest(BaseModel):
    test_result: str
    outcome: str  # supported, refuted, inconclusive


class GenerateRequest(BaseModel):
    context: Optional[str] = None
    max_hypotheses: int = 5


# Routes
@router.post("/investigations")
async def create_investigation(req: CreateInvestigationRequest):
    """Create a new investigation with optional initial entities."""
    try:
        inv = await engine.create_investigation(
            name=req.name,
            description=req.description,
            initial_entities=req.entities if req.entities else None
        )
        return {
            "investigation_id": inv.id,
            "name": inv.name,
            "hypotheses_generated": len(inv.hypotheses),
            "entities_count": len(inv.entities)
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/investigations/{investigation_id}")
async def get_investigation(investigation_id: str):
    """Get investigation details and analysis."""
    try:
        analysis = await engine.analyze_investigation(investigation_id)
        return analysis
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))


@router.get("/investigations/{investigation_id}/export")
async def export_investigation(investigation_id: str):
    """Export full investigation data."""
    data = engine.to_dict(investigation_id)
    if not data:
        raise HTTPException(status_code=404, detail="Investigation not found")
    return data


@router.post("/investigations/{investigation_id}/entities")
async def add_entities(investigation_id: str, req: AddEntitiesRequest):
    """Add entities to an investigation."""
    try:
        inv = await engine.add_entities(
            investigation_id=investigation_id,
            entities=req.entities,
            auto_generate=req.auto_generate
        )
        return {
            "entities_count": len(inv.entities),
            "hypotheses_count": len(inv.hypotheses)
        }
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))


@router.post("/investigations/{investigation_id}/generate")
async def generate_hypotheses(investigation_id: str, req: GenerateRequest):
    """Generate new hypotheses for an investigation."""
    try:
        hypotheses = await engine.generate_hypotheses(
            investigation_id=investigation_id,
            context=req.context,
            max_hypotheses=req.max_hypotheses
        )
        return {
            "generated": len(hypotheses),
            "hypotheses": [
                {
                    "id": h.id,
                    "statement": h.statement,
                    "type": h.hypothesis_type,
                    "confidence": h.confidence,
                    "relevance": h.relevance,
                    "test": h.test_description
                }
                for h in hypotheses
            ]
        }
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))


@router.get("/investigations/{investigation_id}/hypotheses")
async def list_hypotheses(
    investigation_id: str,
    status: Optional[str] = None,
    limit: int = 20
):
    """Get ranked hypotheses for an investigation."""
    try:
        hypotheses = await engine.get_ranked_hypotheses(
            investigation_id=investigation_id,
            status_filter=status,
            limit=limit
        )
        return {
            "count": len(hypotheses),
            "hypotheses": [
                {
                    "id": h.id,
                    "statement": h.statement,
                    "type": h.hypothesis_type,
                    "status": h.status,
                    "confidence": h.confidence,
                    "relevance": h.relevance,
                    "priority": h.priority,
                    "evidence_count": len(h.evidence),
                    "test": h.test_description
                }
                for h in hypotheses
            ]
        }
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))


@router.post("/investigations/{investigation_id}/hypotheses/{hypothesis_id}/score")
async def score_hypothesis(
    investigation_id: str,
    hypothesis_id: str,
    context: Optional[str] = None
):
    """Score/rescore a hypothesis using Haiku."""
    try:
        h = await engine.score_hypothesis(
            investigation_id=investigation_id,
            hypothesis_id=hypothesis_id,
            additional_context=context
        )
        return {
            "id": h.id,
            "confidence": h.confidence,
            "relevance": h.relevance,
            "priority": h.priority,
            "latest_evaluation": h.evaluations[-1] if h.evaluations else None
        }
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))


@router.post("/investigations/{investigation_id}/hypotheses/{hypothesis_id}/evidence")
async def add_evidence(
    investigation_id: str,
    hypothesis_id: str,
    req: AddEvidenceRequest
):
    """Add evidence to a hypothesis."""
    try:
        h = await engine.add_evidence(
            investigation_id=investigation_id,
            hypothesis_id=hypothesis_id,
            evidence={
                "type": req.evidence_type,
                "id": req.evidence_id,
                "text": req.text,
                "supports": req.supports,
                "weight": req.weight
            }
        )
        return {
            "id": h.id,
            "confidence": h.confidence,
            "evidence_count": len(h.evidence),
            "latest_evaluation": h.evaluations[-1] if h.evaluations else None
        }
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))


@router.post("/investigations/{investigation_id}/hypotheses/{hypothesis_id}/test")
async def test_hypothesis(
    investigation_id: str,
    hypothesis_id: str,
    req: TestHypothesisRequest
):
    """Record a hypothesis test result and generate follow-ups."""
    if req.outcome not in ["supported", "refuted", "inconclusive"]:
        raise HTTPException(status_code=400, detail="Invalid outcome")
    
    try:
        result = await engine.test_hypothesis(
            investigation_id=investigation_id,
            hypothesis_id=hypothesis_id,
            test_result=req.test_result,
            outcome=req.outcome
        )
        
        h = result["updated_hypothesis"]
        followups = result["followup_hypotheses"]
        
        return {
            "hypothesis": {
                "id": h.id,
                "status": h.status,
                "confidence": h.confidence
            },
            "followup_hypotheses": [
                {
                    "id": f.id,
                    "statement": f.statement,
                    "type": f.hypothesis_type,
                    "confidence": f.confidence
                }
                for f in followups
            ]
        }
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))


class QuickAnalyzeRequest(BaseModel):
    entities: List[Dict[str, Any]]
    context: str = ""


# Quick analysis endpoint
@router.post("/analyze")
async def quick_analyze(req: QuickAnalyzeRequest):
    """Quick analysis: create investigation, generate hypotheses, return analysis."""
    entities = req.entities
    context = req.context
    try:
        inv = await engine.create_investigation(
            name="Quick Analysis",
            description=context,
            initial_entities=entities
        )
        
        # Score all hypotheses
        for h in inv.hypotheses:
            await engine.score_hypothesis(inv.id, h.id)
        
        analysis = await engine.analyze_investigation(inv.id)
        return analysis
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

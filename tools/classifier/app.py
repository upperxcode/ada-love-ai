from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List
import os

# This is a minimal local classifier service that exposes POST /classify
# Expected input JSON:
# {"text": "...", "candidate_labels": ["label1","label2"]}
# Response JSON:
# {"labels": [...], "scores": [...], "top_label": "..."}

# Default model: uses HF zero-shot via transformers pipeline (local inference).
# Change BACKEND_MODEL env var to another supported zero-shot model if you prefer.

BACKEND_MODEL = os.environ.get("BACKEND_MODEL", "facebook/bart-large-mnli")

app = FastAPI(title="Local Intent Classifier")

class ClassifyReq(BaseModel):
    text: str
    candidate_labels: List[str]

class ClassifyResp(BaseModel):
    labels: List[str]
    scores: List[float]
    top_label: str

# lazy import of transformers to avoid heavy startup if not used
_classifier = None

def get_classifier():
    global _classifier
    if _classifier is None:
        try:
            from transformers import pipeline
            _classifier = pipeline("zero-shot-classification", model=BACKEND_MODEL)
        except Exception as e:
            raise RuntimeError(f"failed to initialize classifier pipeline: {e}")
    return _classifier

@app.post("/classify", response_model=ClassifyResp)
async def classify(req: ClassifyReq):
    if not req.candidate_labels or not req.text:
        raise HTTPException(status_code=400, detail="text and candidate_labels required")
    cls = get_classifier()
    # pipeline returns e.g. {labels: [...], scores:[...], sequence: ...}
    try:
        r = cls(req.text, req.candidate_labels)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

    labels = r["labels"] if isinstance(r, dict) and "labels" in r else r.labels
    scores = r["scores"] if isinstance(r, dict) and "scores" in r else r.scores
    top_label = labels[0] if len(labels) > 0 else ""
    return ClassifyResp(labels=labels, scores=[float(s) for s in scores], top_label=top_label)

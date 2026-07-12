Local Zero-shot Intent Classifier

This is a minimal FastAPI app that exposes POST /classify and runs a local
zero-shot classification pipeline using HuggingFace `transformers`.

Quick start (recommended in a venv):

python -m venv venv
source venv/bin/activate
pip install -r requirements.txt

# optionally set model (defaults to facebook/bart-large-mnli):
export BACKEND_MODEL="facebook/bart-large-mnli"

# run the service
uvicorn app:app --host 127.0.0.1 --port 8008

The endpoint accepts JSON {"text": "...", "candidate_labels": ["a","b"]}
and returns {"labels": [...], "scores": [...], "top_label": "..."}.

Notes:
- This is a template. For production or higher throughput, replace the model
  with a smaller/faster local model or a dedicated runtime.
- To change classifier model, set BACKEND_MODEL env var before running.

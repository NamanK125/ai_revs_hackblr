"""FastEmbed sidecar for the EngineerOps Voice backend.

Exposes a single POST /embed endpoint that batch-embeds short texts using
BAAI/bge-small-en-v1.5 (384-dim) on CPU. The model is loaded once at module
import time so per-request latency stays in the ~30ms/chunk range.
"""

from typing import List

from fastapi import FastAPI, HTTPException
from fastembed import TextEmbedding
from pydantic import BaseModel, Field

MODEL_NAME = "BAAI/bge-small-en-v1.5"
EMBED_DIM = 384
MAX_BATCH = 256
MAX_TEXT_CHARS = 8000

# Singleton: load the model once at import time.
_model = TextEmbedding(MODEL_NAME)

app = FastAPI(title="EngineerOps Embedder", version="0.1.0")


class EmbedRequest(BaseModel):
    texts: List[str] = Field(..., description="Texts to embed (batch).")


class EmbedResponse(BaseModel):
    vectors: List[List[float]]
    model: str
    dim: int


@app.get("/health")
def health() -> dict:
    return {"ok": True}


@app.post("/embed", response_model=EmbedResponse)
def embed(req: EmbedRequest) -> EmbedResponse:
    texts = req.texts
    if not texts:
        raise HTTPException(status_code=422, detail="texts must be non-empty")
    if len(texts) > MAX_BATCH:
        raise HTTPException(
            status_code=422,
            detail=f"too many texts: {len(texts)} > {MAX_BATCH}",
        )
    for i, t in enumerate(texts):
        if not isinstance(t, str):
            raise HTTPException(status_code=422, detail=f"texts[{i}] is not a string")
        if len(t) > MAX_TEXT_CHARS:
            raise HTTPException(
                status_code=422,
                detail=f"texts[{i}] too long: {len(t)} > {MAX_TEXT_CHARS} chars",
            )

    # fastembed returns a generator of numpy arrays; materialize as plain lists for JSON.
    vectors = [vec.tolist() for vec in _model.embed(texts)]

    return EmbedResponse(vectors=vectors, model=MODEL_NAME, dim=EMBED_DIM)

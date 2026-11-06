"""App assembly for `uvicorn app.main:app` (requires OPENCODE_API_KEY in env)."""

from .api import create_app
from .graph import build_graph

app = create_app()

if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8000)

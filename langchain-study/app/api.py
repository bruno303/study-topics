"""FastAPI layer: thin HTTP wrappers around the compiled LangGraph pipeline.

All decision logic lives in the graph; this module only:
- translates request bodies into graph state inputs,
- passes a thread_id in the config (checkpointer key),
- formats results back to HTTP responses / SSE events.
"""

import json
import uuid
from collections.abc import AsyncIterator

from fastapi import FastAPI, HTTPException
from fastapi.responses import StreamingResponse

from .graph import build_graph
from .schemas import (
    AnalysisRequest,
    AnalysisResponse,
    FollowupRequest,
    FollowupResponse,
)


def _plain(value):
    """Make graph values JSON-safe inside SSE payloads."""
    if hasattr(value, "model_dump"):
        return value.model_dump(mode="json")
    return value


def _sse(payload: dict) -> str:
    return f"data: {json.dumps(payload, default=str)}\n\n"


def create_app(graph=None) -> FastAPI:
    """App factory: pass a graph built with a fake LLM for offline tests."""
    graph = graph or build_graph()

    app = FastAPI(
        title="langchain-study",
        description="Sequential LangGraph prompt chain (car analyzer) over the OpenCode inference API.",
        version="0.1.0",
    )

    @app.get("/health")
    async def health() -> dict:
        return {"status": "ok"}

    @app.post("/analyses", response_model=AnalysisResponse)
    async def create_analysis(req: AnalysisRequest) -> AnalysisResponse:
        thread_id = req.thread_id or str(uuid.uuid4())
        config = {"configurable": {"thread_id": thread_id}}
        final = await graph.ainvoke(
            {
                "car_model": req.car_model,
                "year": req.year,
                "question": None,  # clear any prior follow-up question
            },
            config=config,
        )
        return AnalysisResponse(
            thread_id=thread_id,
            car_info=final.get("car_info"),
            market_info=final.get("market_info"),
            report=final.get("report"),
            error=final.get("error"),
        )

    @app.post("/analyses/stream")
    async def stream_analysis(req: AnalysisRequest) -> StreamingResponse:
        thread_id = req.thread_id or str(uuid.uuid4())
        config = {"configurable": {"thread_id": thread_id}}
        state_input = {
            "car_model": req.car_model,
            "year": req.year,
            "question": None,
        }

        async def event_stream() -> AsyncIterator[str]:
            # stream_mode="updates" emits {node_name: state_delta} per superstep,
            # which makes sequence/progress visible to the client.
            try:
                yield _sse({"event": "start", "thread_id": thread_id})
                async for chunk in graph.astream(state_input, config=config, stream_mode="updates"):
                    for node_name, update in (chunk or {}).items():
                        if node_name.startswith("__"):
                            continue
                        yield _sse(
                            {
                                "event": "node_update",
                                "node": node_name,
                                "data": {
                                    key: _plain(value) for key, value in (update or {}).items()
                                },
                            }
                        )
                        if node_name == "report" and update.get("report"):
                            yield _sse({"event": "report", "content": update["report"]})
                snapshot = await graph.aget_state(config)
                final = snapshot.values or {}
                yield _sse(
                    {
                        "event": "done",
                        "thread_id": thread_id,
                        "car_info": _plain(final.get("car_info")),
                        "market_info": _plain(final.get("market_info")),
                        "error": final.get("error"),
                    }
                )
            except Exception as exc:  # noqa: BLE001 - SSE needs the error as an event
                yield _sse({"event": "error", "message": str(exc)})

        return StreamingResponse(
            event_stream(),
            media_type="text/event-stream",
            headers={"Cache-Control": "no-cache", "X-Accel-Buffering": "no"},
        )

    @app.post("/analyses/{thread_id}/followup", response_model=FollowupResponse)
    async def followup(thread_id: str, req: FollowupRequest) -> FollowupResponse:
        config = {"configurable": {"thread_id": thread_id}}
        snapshot = await graph.aget_state(config)
        if not (snapshot.values or {}).get("car_info"):
            raise HTTPException(
                status_code=404,
                detail="No analysis found for this thread_id; run POST /analyses first.",
            )
        final = await graph.ainvoke({"question": req.question}, config=config)
        return FollowupResponse(
            thread_id=thread_id,
            question=req.question,
            answer=final["last_answer"],
        )

    @app.get("/analyses/{thread_id}", response_model=AnalysisResponse)
    async def get_analysis(thread_id: str) -> AnalysisResponse:
        config = {"configurable": {"thread_id": thread_id}}
        snapshot = await graph.aget_state(config)
        if not snapshot.values:
            raise HTTPException(status_code=404, detail="Unknown thread_id.")
        values = snapshot.values
        return AnalysisResponse(
            thread_id=thread_id,
            car_info=values.get("car_info"),
            market_info=values.get("market_info"),
            report=values.get("report"),
            error=values.get("error"),
        )

    return app

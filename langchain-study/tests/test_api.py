"""HTTP layer tests: FastAPI routes against a graph wired to a fake LLM."""

import json

from fastapi.testclient import TestClient

from app.api import create_app
from app.graph import build_graph
from app.schemas import CarInfo
from tests.fakes import make_llm


def make_client():
    llm, _, _ = make_llm()
    graph = build_graph(llm=llm)
    app = create_app(graph=graph)
    return TestClient(app), llm


def test_health():
    client, _ = make_client()
    assert client.get("/health").json() == {"status": "ok"}


def test_create_analysis_returns_full_report():
    client, _ = make_client()
    response = client.post("/analyses", json={"car_model": "Toyota Corolla", "year": 2022})

    assert response.status_code == 200
    body = response.json()
    assert body["car_info"]["make"] == "Toyota"
    assert body["market_info"]["price_high"] == 27500.0
    assert body["report"] == "REPORT TEXT"
    assert body["error"] is None
    assert body["thread_id"]  # generated when not supplied


def test_followup_endpoint_uses_persisted_state():
    client, _ = make_client()
    created = client.post("/analyses", json={"car_model": "Toyota Corolla"}).json()

    response = client.post(
        f"/analyses/{created['thread_id']}/followup",
        json={"question": "Any common problems?"},
    )

    assert response.status_code == 200
    assert response.json() == {
        "thread_id": created["thread_id"],
        "question": "Any common problems?",
        "answer": "FOLLOWUP ANSWER",
    }


def test_followup_unknown_thread_returns_404():
    client, _ = make_client()
    response = client.post(
        "/analyses/no-such-thread/followup",
        json={"question": "hello?"},
    )
    assert response.status_code == 404


def test_get_analysis_reads_checkpointer_state():
    client, _ = make_client()
    created = client.post(
        "/analyses",
        json={"car_model": "Toyota Corolla", "thread_id": "fixed-id"},
    ).json()

    response = client.get(f"/analyses/{created['thread_id']}")

    assert response.status_code == 200
    assert response.json()["car_info"]["model"] == "Corolla"
    assert response.json()["thread_id"] == "fixed-id"


def test_get_analysis_unknown_thread_returns_404():
    client, _ = make_client()
    assert client.get("/analyses/unknown").status_code == 404


def _parse_sse(text: str) -> list[dict]:
    return [json.loads(line[len("data: "):]) for line in text.splitlines() if line.startswith("data: ")]


def test_stream_emits_progress_events_in_order():
    client, _ = make_client()
    with client.stream("POST", "/analyses/stream", json={"car_model": "Toyota Corolla"}) as response:
        assert response.status_code == 200
        assert response.headers["content-type"].startswith("text/event-stream")
        text = "".join(response.iter_text())

    events = _parse_sse(text)
    kinds = [event["event"] for event in events]
    assert kinds[0] == "start"
    assert "report" in kinds
    assert "done" in kinds

    nodes = [event["node"] for event in events if event["event"] == "node_update"]
    assert nodes == ["identify_specs", "market_research", "report"]

    done = events[-1]
    assert done["error"] is None


def test_stream_reports_unidentified_car_without_market_steps():
    llm, _, _ = make_llm()
    llm._structured[CarInfo] = [CarInfo(identified=False)]
    graph = build_graph(llm=llm)
    client = TestClient(create_app(graph=graph))

    with client.stream("POST", "/analyses/stream", json={"car_model": "Flying Falcon 3000X"}) as response:
        text = "".join(response.iter_text())

    events = _parse_sse(text)
    nodes = [event["node"] for event in events if event["event"] == "node_update"]
    assert nodes == ["identify_specs", "not_found"]
    assert events[-1]["error"]

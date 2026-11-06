"""Offline unit tests for the LangGraph pipeline wiring (no network, no server)."""

from app.graph import build_graph
from app.schemas import CarInfo, MarketInfo
from tests.fakes import FakeLLM, make_llm


def make_graph(llm: FakeLLM):
    # Uses the app-default checkpointer (MemorySaver + allowlisted schemas).
    return build_graph(llm=llm)


def make_config(thread_id: str) -> dict:
    return {"configurable": {"thread_id": thread_id}}


def test_happy_path_runs_all_steps_in_order():
    llm, car, market = make_llm()
    graph = make_graph(llm)

    final = graph.invoke(
        {"car_model": "Toyota Corolla", "year": 2022, "question": None},
        config=make_config("t-happy"),
    )

    assert final["car_info"] is car
    assert final["market_info"] is market
    assert final["report"] == "REPORT TEXT"
    assert final.get("error") is None
    # steps: structured(CarInfo) -> structured(MarketInfo) -> plain text (report)
    assert llm.called_schemas() == ["CarInfo", "MarketInfo", None]


def test_step_two_prompt_embeds_step_one_output():
    llm, _, _ = make_llm()
    graph = make_graph(llm)

    graph.invoke(
        {"car_model": "Toyota Corolla", "year": 2022, "question": None},
        config=make_config("t-embed"),
    )

    market_prompts = llm.human_text_for("MarketInfo")
    assert "2.0L petrol and hybrid options" in market_prompts  # from step 1
    assert "Toyota Corolla" in market_prompts


def test_unidentified_car_routes_to_not_found_and_skips_rest():
    llm, _, _ = make_llm()
    llm._structured[CarInfo] = [CarInfo(identified=False)]
    llm._structured[MarketInfo] = []
    llm._texts = []
    graph = make_graph(llm)

    final = graph.invoke(
        {"car_model": "Flying Falcon 3000X", "question": None},
        config=make_config("t-notfound"),
    )

    assert final["error"] is not None
    assert final.get("market_info") is None
    assert final.get("report") is None
    # only step 1 ran: no MarketInfo or plain-text calls
    assert llm.called_schemas() == ["CarInfo"]


def test_structured_output_objects_are_preserved_in_state():
    llm, _, _ = make_llm()
    graph = make_graph(llm)

    final = graph.invoke(
        {"car_model": "Honda Civic", "question": None},
        config=make_config("t-types"),
    )

    assert isinstance(final["car_info"], CarInfo)
    assert isinstance(final["market_info"], MarketInfo)

"""Checkpointer / multi-turn behavior: a follow-up continues the same thread."""

from app.graph import build_graph
from tests.fakes import make_llm


def test_followup_no_rerun_of_previous_steps():
    llm, _, _ = make_llm()
    graph = build_graph(llm=llm)
    config = {"configurable": {"thread_id": "t-follow"}}

    first = graph.invoke(
        {"car_model": "Toyota Corolla", "year": 2022, "question": None},
        config=config,
    )
    assert first["report"] == "REPORT TEXT"
    structured_calls_before = llm.called_schemas().count("CarInfo")

    second = graph.invoke({"question": "Any common problems?"}, config=config)

    assert second["last_answer"] == "FOLLOWUP ANSWER"
    # the previous steps were NOT re-executed: step 1 ran exactly once
    assert llm.called_schemas().count("CarInfo") == structured_calls_before
    # the four calls for this conversation: specs, market, report, followup
    assert llm.called_schemas() == ["CarInfo", "MarketInfo", None, None]


def test_checkpointer_roundtrip_has_no_unregistered_type_warning(caplog):
    """The default serializer must cleanly revive CarInfo/MarketInfo checkpoints.

    Regression guard for the langgraph "Deserializing unregistered type ... will
    be blocked in a future version" warning seen on the live run.
    """
    llm, _, _ = make_llm()
    graph = build_graph(llm=llm)
    config = {"configurable": {"thread_id": "t-serde"}}

    graph.invoke(
        {"car_model": "Toyota Corolla", "year": 2022, "question": None},
        config=config,
    )
    graph.invoke({"question": "How does it handle snow?"}, config=config)

    assert not any(
        "unregistered type" in record.getMessage() for record in caplog.records
    )


def test_followup_prompt_contains_persisted_context():
    llm, _, _ = make_llm()
    graph = build_graph(llm=llm)
    config = {"configurable": {"thread_id": "t-context"}}

    graph.invoke(
        {"car_model": "Toyota Corolla", "year": 2022, "question": None},
        config=config,
    )
    graph.invoke({"question": "How does it handle snow?"}, config=config)

    followup_prompts = llm.human_text_for(None)
    # follow-up sees step 1 and step 2 outputs stored by the checkpointer
    assert "2.0L petrol and hybrid options" in followup_prompts
    assert "27500.0" in followup_prompts
    assert "How does it handle snow?" in followup_prompts

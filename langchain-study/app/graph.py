"""LangGraph orchestration: the sequential car-analysis pipeline.

Graph wiring (first turn):

    START -> identify_specs --identified--> market_research -> report -> END
                     \\--not identified--> not_found -----------------> END

Follow-up turn (same thread_id, previous state restored by the checkpointer):

    START -> followup -> END

Concepts demonstrated:
- StateGraph state: every node returns a dict of state updates, merged by LangGraph.
- Structured output: ChatOpenAI.with_structured_output(PydanticModel).
- Conditional edge: `start_router` (fresh analysis vs follow-up) and the
  `route_after_specs` router after step 1.
- Checkpointer: MemorySaver (schemas allowlisted in its serializer) + thread_id
  makes the graph state survive between HTTP calls (multi-turn behavior).
"""

from typing import Literal, TypedDict

from langchain_core.language_models.chat_models import BaseChatModel
from langgraph.checkpoint.memory import MemorySaver
from langgraph.checkpoint.serde.jsonplus import JsonPlusSerializer
from langgraph.graph import END, START, StateGraph
from langgraph.graph.state import CompiledStateGraph

from .models import chat_model
from .schemas import CarInfo, MarketInfo

# Checkpoints store our Pydantic result objects. Explicitly allowlisting them
# keeps the checkpointer round-trip warning-free and strict (instead of
# langgraph's permissive default, which will become a hard block in futures).
_CHECKPOINT_SERIALIZER = JsonPlusSerializer(allowed_msgpack_modules=[CarInfo, MarketInfo])


class CarAnalysisState(TypedDict):
    """Shared state flowing through every node; nodes update only their keys."""

    car_model: str
    year: int | None
    question: str | None          # set only on follow-up turns
    car_info: CarInfo | None      # step 1
    market_info: MarketInfo | None   # step 2 (built from car_info)
    report: str | None            # step 3 (merges car_info + market_info)
    error: str | None             # set by not_found
    last_answer: str | None       # set by followup


def build_graph(
    llm: BaseChatModel | None = None,
    checkpointer: MemorySaver | None = None,
) -> CompiledStateGraph:
    """Compile the analysis graph.

    `llm` and `checkpointer` are injectable so tests can run the whole pipeline
    offline; defaults wire the real OpenCode model and an in-memory checkpointer.
    """
    llm = llm or chat_model()
    checkpointer = checkpointer or MemorySaver(serde=_CHECKPOINT_SERIALIZER)

    def _car_label(state: CarAnalysisState) -> str:
        year = state.get("year")
        label = state.get("car_model", "")
        return f"{year} {label}".strip() if year else label

    def _specs_text(state: CarAnalysisState) -> str:
        return state["car_info"].model_dump_json(indent=2)

    # -- nodes ---------------------------------------------------------------

    def identify_specs(state: CarAnalysisState) -> dict:
        """Step 1: extract facts about the car from just its name/year."""
        car_info = llm.with_structured_output(CarInfo).invoke(
            [
                (
                    "system",
                    "You are a car analyst with encyclopedic knowledge of vehicles. "
                    "If the requested name does not clearly match a real, specific car "
                    "model, set identified=false and leave the other fields empty.",
                ),
                (
                    "human",
                    f"Car requested: {_car_label(state)}\n"
                    "Extract the key facts about this car.",
                ),
            ]
        )
        return {"car_info": car_info}

    def route_after_specs(state: CarAnalysisState) -> Literal["market_research", "not_found"]:
        return "market_research" if state["car_info"].identified else "not_found"

    def market_research(state: CarAnalysisState) -> dict:
        """Step 2: prompt *builds on* step 1's structured output."""
        market_info = llm.with_structured_output(MarketInfo).invoke(
            [
                (
                    "system",
                    "You are a market analyst for used and new cars. "
                    "Always estimate price ranges conservatively and mention a caveat "
                    "that these are knowledge-based estimates.",
                ),
                (
                    "human",
                    "Specs from the previous analysis step:\n"
                    f"{_specs_text(state)}\n\n"
                    f"Based on those specs, give the recent market context for the "
                    f"{_car_label(state)}.",
                ),
            ]
        )
        return {"market_info": market_info}

    def report(state: CarAnalysisState) -> dict:
        """Step 3: plain-text invocation that merges both earlier results."""
        response = llm.invoke(
            [
                (
                    "system",
                    "You write concise car-analysis reports in markdown.",
                ),
                (
                    "human",
                    f"Car: {_car_label(state)}\n"
                    f"Extracted specs:\n{_specs_text(state)}\n"
                    f"Market data:\n{state['market_info'].model_dump_json(indent=2)}\n\n"
                    "Write a short markdown report combining both sections.",
                ),
            ]
        )
        return {"report": response.content}

    def not_found(state: CarAnalysisState) -> dict:
        return {
            "error": (
                f"Could not identify the car '{_car_label(state)}' as a real model; "
                "the market-research and report steps were skipped."
            )
        }

    def followup(state: CarAnalysisState) -> dict:
        """Follow-up turn: answer a question using the persisted state."""
        specs = _specs_text(state) if state.get("car_info") else "n/a"
        market = (
            state["market_info"].model_dump_json(indent=2)
            if state.get("market_info")
            else "n/a"
        )
        answer = llm.invoke(
            [
                (
                    "system",
                    "You continue an ongoing car-analysis conversation. Answer using "
                    "only the context and your general car knowledge; say so when the "
                    "context does not cover the question.",
                ),
                (
                    "human",
                    f"Car info:\n{specs}\n\nMarket info:\n{market}\n\n"
                    f"Question: {state['question']}",
                ),
            ]
        )
        return {"last_answer": answer.content}

    def start_router(state: CarAnalysisState) -> Literal["identify_specs", "followup"]:
        """Fresh analysis when there is no question; otherwise resume the thread."""
        return "followup" if state.get("question") else "identify_specs"

    # -- wiring ---------------------------------------------------------------

    builder = StateGraph(CarAnalysisState)
    builder.add_node("identify_specs", identify_specs)
    builder.add_node("market_research", market_research)
    builder.add_node("report", report)
    builder.add_node("not_found", not_found)
    builder.add_node("followup", followup)

    builder.add_conditional_edges(
        START,
        start_router,
        {"identify_specs": "identify_specs", "followup": "followup"},
    )
    builder.add_conditional_edges(
        "identify_specs",
        route_after_specs,
        {"market_research": "market_research", "not_found": "not_found"},
    )
    builder.add_edge("market_research", "report")
    for node in ("report", "not_found", "followup"):
        builder.add_edge(node, END)

    return builder.compile(checkpointer=checkpointer)

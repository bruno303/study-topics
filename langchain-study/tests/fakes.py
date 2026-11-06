"""Test doubles shared by the offline test suite."""

from types import SimpleNamespace

from app.schemas import CarInfo, MarketInfo


class FakeStructuredOutput:
    """Mirrors the real shape of `llm.with_structured_output(schema)`: a chain
    object exposing `.invoke(messages)` that returns the structured result."""

    def __init__(self, llm: "FakeLLM", schema: type):
        self._llm = llm
        self._schema = schema

    def invoke(self, messages):
        self._llm.calls.append((self._schema.__name__, messages))
        return self._llm._structured[self._schema].pop(0)


class FakeLLM:
    """Duck-typed stand-in for a LangChain chat model.

    - `with_structured_output(schema)` returns a callable whose `.invoke(...)`
      pops the next pre-wired result for that schema (like the real structured
      output chain, without any network call).
    - `invoke(...)` (plain text calls) pops entries from `texts` and wraps them
      in an object exposing `.content`, exactly what the graph nodes use.
    - Every call is recorded in `self.calls` as (schema_name_or_None, messages)
      so tests can assert node ordering and prompt contents.
    """

    def __init__(self, structured: dict[type, list], texts: list[str]):
        self._structured = {key: list(value) for key, value in structured.items()}
        self._texts = list(texts)
        self.calls: list[tuple[str | None, list]] = []

    def with_structured_output(self, schema):
        return FakeStructuredOutput(self, schema)

    def invoke(self, messages):
        self.calls.append((None, messages))
        return SimpleNamespace(content=self._texts.pop(0))

    # -- assertions helpers ---------------------------------------------------

    def called_schemas(self) -> list[str | None]:
        return [name for name, _ in self.calls]

    def human_text_for(self, schema_name: str | None) -> str:
        """Concatenated request text of all calls made with a given schema."""
        chunks = []
        for name, messages in self.calls:
            if name != schema_name:
                continue
            for message in messages:
                if isinstance(message, tuple):
                    chunks.append(message[1])
                else:
                    chunks.append(str(message))
        return "\n".join(chunks)


def make_llm() -> tuple[FakeLLM, CarInfo, MarketInfo]:
    """Standard happy-path fixtures: one CarInfo, one MarketInfo, two texts."""
    car = CarInfo(
        identified=True,
        make="Toyota",
        model="Corolla",
        year=2022,
        segment="compact sedan",
        powertrain="2.0L petrol and hybrid options",
        highlights=["low running costs", "high reliability"],
        summary="A pragmatic compact sedan.",
    )
    market = MarketInfo(
        currency="USD",
        price_low=19000.0,
        price_high=27500.0,
        market_trend="stable",
        competitors=["Honda Civic", "Hyundai Elantra"],
    )
    llm = FakeLLM(
        {CarInfo: [car], MarketInfo: [market]},
        texts=["REPORT TEXT", "FOLLOWUP ANSWER"],
    )
    return llm, car, market

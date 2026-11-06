"""Pydantic models: LLM structured outputs (steps 1 and 2) + API request/response bodies."""

from pydantic import BaseModel, Field


# ---------------------------------------------------------------------------
# LLM structured outputs (used with llm.with_structured_output(...))
# ---------------------------------------------------------------------------

class CarInfo(BaseModel):
    """Step 1 output: factual breakdown of the requested car."""

    identified: bool = Field(
        description="True when the text refers to a real, specific car model.",
    )
    make: str | None = Field(None, description="Manufacturer, e.g. 'Toyota'.")
    model: str | None = Field(None, description="Model name, e.g. 'Corolla'.")
    year: int | None = Field(None, description="Model year when known.")
    segment: str | None = Field(None, description="Market segment, e.g. 'compact sedan'.")
    powertrain: str | None = Field(
        None,
        description="Engines / fuel options offered for this car.",
    )
    highlights: list[str] = Field(
        default_factory=list,
        description="Key facts worth mentioning in an analysis.",
    )
    summary: str | None = Field(None, description="One-paragraph description of the car.")


class MarketInfo(BaseModel):
    """Step 2 output: market context, built on top of step 1's CarInfo."""

    currency: str = Field(default="USD")
    price_low: float | None = Field(
        None,
        description="Typical recent low price for this car/segment.",
    )
    price_high: float | None = Field(
        None,
        description="Typical recent high price for this car/segment.",
    )
    market_trend: str | None = Field(
        None,
        description="Qualitative trend, e.g. 'stable', 'depreciating fast'.",
    )
    competitors: list[str] = Field(default_factory=list)
    caveat: str = Field(
        default="Prices are estimates from the model's training knowledge, not live market data.",
    )


# ---------------------------------------------------------------------------
# API bodies
# ---------------------------------------------------------------------------

class AnalysisRequest(BaseModel):
    """POST /analyses and POST /analyses/stream body."""

    car_model: str = Field(min_length=1, examples=["Toyota Corolla"])
    year: int | None = Field(None, ge=1900, le=2100)
    # Optional: pass an existing thread_id to re-run the analysis and keep
    # following up in the same graph state.
    thread_id: str | None = None


class AnalysisResponse(BaseModel):
    thread_id: str
    car_info: CarInfo | None = None
    market_info: MarketInfo | None = None
    report: str | None = None
    error: str | None = None


class FollowupRequest(BaseModel):
    """POST /analyses/{thread_id}/followup body."""

    question: str = Field(min_length=1)


class FollowupResponse(BaseModel):
    thread_id: str
    question: str
    answer: str

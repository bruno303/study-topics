# langchain-study — sequential LangGraph prompt chain (car analyzer)

A small **learning** project that orchestrates a sequence of LLM calls with
[LangGraph](https://langchain-ai.github.io/langgraph/) and
[LangChain](https://docs.langchain.com), served over a FastAPI API:

> 📘 **Want to understand *why* LangChain/LangGraph are here?** Read
> [`docs/understanding-langchain.md`](docs/understanding-langchain.md) — what each
> library does (and what you would hand-write without it), with diagrams.

1. **Step 1 — specs**: a prompt extracts facts about the car model you send.
2. **Step 2 — market**: a second prompt *builds on step 1's structured output*
   to give recent market context (price range, competitors, trend).
3. **Step 3 — report**: a third prompt merges both into a markdown report.

All three calls hit the **OpenCode inference API** (OpenAI-compatible), but the
orchestration lives entirely in the LangGraph state graph.

## Learning map (concept → file)

| LangGraph concept | Where |
|---|---|
| State (`TypedDict`) + state updates | `app/graph.py` → `CarAnalysisState` |
| Nodes & sequential edges | `build_graph` wiring in `app/graph.py` |
| Conditional edge (identify/`not_found`) | `route_after_specs` router |
| Structured output (Pydantic) | `identify_specs` / `market_research` in `app/graph.py`, schemas in `app/schemas.py` |
| Checkpointer / multi-turn (`thread_id`) | `MemorySaver` in `build_graph` (schemas allowlisted in its serializer), `followup` route + `/analyses/{thread_id}/followup` |
| Streaming (`astream` + SSE) | `/analyses/stream` in `app/api.py` |
| LLM client (OpenAI-compatible endpoint, Go-recommended headers) | `app/models.py` |

## Setup

```sh
cd langchain-study
uv sync --extra dev          # or: python3 -m venv .venv && .venv/bin/pip install -e ".[dev]"
cp .env.example .env         # then fill OPENCODE_API_KEY
```

Config notes (all in `.env`):

- **With an OpenCode Go subscription**: keep the Go base URL
  `https://opencode.ai/inference/go/openai/v1` and set `OPENCODE_MODEL` to a Go
  model id — e.g. `glm-5.3-flash` (largest static monthly allowance, cheapest).
  List the live Go models with `curl https://opencode.ai/inference/go/v1/models`;
  stick to chat-completions-family ids (`glm-*`, `kimi-*`, `longcat-*`, ...).
- **Without Go** (pay-as-you-go Console key): switch `OPENCODE_BASE_URL` to
  `https://opencode.ai/inference/openai/v1` and use a paid Chat Completions
  model such as `glm-5.3`. Free-tier chat models only work **inside** OpenCode
  itself — external clients get `HTTP 403`.
- `OPENCODE_MODEL` and `OPENCODE_BASE_URL` are both read from `.env`. Defaults
  are Go-first: `glm-5.3-flash` on `https://opencode.ai/inference/go/openai/v1`.
  For **pay-as-you-go** Console keys (no Go), switch `OPENCODE_BASE_URL` to
  `https://opencode.ai/inference/openai/v1` and use a paid Chat Completions
  model such as `glm-5.3`. Free-tier chat models only work **inside** OpenCode
  itself — external clients get `HTTP 403`.
- `OPENCODE_BASE_URL` must never include `/chat/completions` — LangChain appends
  the path itself; including it causes confusing 404s.
- OpenCode Go recommends clients send identifying headers on every request;
  `app/models.py` always sends both:
  - `User-Agent: langchain-study/0.1.0` (self-identified, not the generic SDK name);
    override with `OPENCODE_USER_AGENT`.
  - `x-opencode-session`: a stable conversation id (from `OPENCODE_SESSION_ID`
    when set, else generated once per process).
  Note: per-HTTP-analysis mapping (one session id per `thread_id`) would need a
  custom client — a nice v2 exercise. Responses also report usage against your
  plan (`opencode.billing_source`, `opencode.allowance` in the usage payload).

## Run

```sh
uv run uvicorn app.main:app --reload    # or: .venv/bin/uvicorn app.main:app --reload
```

## Endpoints

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/analyses` | Run the full pipeline synchronously, returns JSON report |
| `POST` | `/analyses/stream` | Same run but streamed as SSE progress events |
| `POST` | `/analyses/{thread_id}/followup` | Ask a follow-up; state is restored from the checkpointer |
| `GET` | `/analyses/{thread_id}` | Inspect the persisted graph state for a thread |
| `GET` | `/health` | Liveness |

Examples:

```sh
# 1. full analysis (car name, optional year)
curl -s -X POST localhost:8000/analyses \
  -H 'Content-Type: application/json' \
  -d '{"car_model": "Toyota Corolla", "year": 2022}' | jq

# 2. streamed version (watch the nodes fire one by one)
curl -N -X POST localhost:8000/analyses/stream \
  -H 'Content-Type: application/json' \
  -d '{"car_model": "Toyota Corolla", "year": 2022}'

# 3. follow up inside the same thread (uses the checkpointer)
curl -s -X POST localhost:8000/analyses/<thread_id>/followup \
  -H 'Content-Type: application/json' \
  -d '{"question": "Is the hybrid variant cheaper to maintain?"}' | jq

# 4. inspect persisted state
curl -s localhost:8000/analyses/<thread_id> | jq
```

SSE events (one JSON object per `data:` line): `start` → `node_update` (one per
node, with its state delta) → `report` (full text) → `done` (final snapshot), or
`error` if a prompt call fails. Unknown models route to the `not_found` node and
you get an `error` field instead of market data.

## Tests (fully offline — no API key/network needed)

```sh
uv run pytest -q
```

The LLM is faked by `tests/fakes.py` (`FakeLLM`), injected through the
`build_graph(llm=..., checkpointer=...)` factory, so the graph wiring, prompts,
conditional edges, checkpointer behavior and HTTP/SSE layers are all tested
without the network.

## Known scope limits (on purpose)

- Prices are **knowledge-based estimates** — no live market data (v2 idea: add a
  web-search tool node).
- `MemorySaver` keeps threads in RAM only; state is lost when the process exits.
- Errors from the LLM endpoint surface as a 500 (JSON) or an SSE `error` event;
  no retries beyond `max_retries=1`.

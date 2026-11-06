"""LangChain chat model factory: points ChatOpenAI at the OpenCode inference API.

The OpenCode inference API is OpenAI-compatible, so the plain ChatOpenAI class
is all we need; only base_url / api_key / model change. LangChain appends
`/chat/completions` to base_url itself, so base_url must be the bare root.

Headers (what the OpenCode Go docs ask from direct clients):
- `Authorization: Bearer <key>`  -> handled by the `api_key` argument.
- `x-opencode-session`           -> stable conversation id per process, for
  routing/prompt-cache optimization. Taken from OPENCODE_SESSION_ID when set,
  otherwise generated once per process.
- `User-Agent`                   -> identify the client yourself instead of the
  generic SDK name; defaults to this project's name, OPENCODE_USER_AGENT overrides.
"""

import os
import uuid

from langchain_openai import ChatOpenAI

# Defaults match an OpenCode Go subscription; pay-as-you-go keys should set
# OPENCODE_BASE_URL=https://opencode.ai/inference/openai/v1 in .env.
DEFAULT_BASE_URL = "https://opencode.ai/inference/go/openai/v1"
DEFAULT_MODEL = "glm-5.3-flash"
DEFAULT_USER_AGENT = "langchain-study/0.1.0"


def _default_headers() -> dict:
    """Headers OpenCode Go recommends identifying yourself with."""
    return {
        "User-Agent": os.getenv("OPENCODE_USER_AGENT") or DEFAULT_USER_AGENT,
        "x-opencode-session": os.getenv("OPENCODE_SESSION_ID")
        or f"langchain-study-{uuid.uuid4()}",
    }


def chat_model() -> ChatOpenAI:
    """Build a ChatOpenAI configured from env vars (see .env.example)."""
    api_key = os.environ.get("OPENCODE_API_KEY")
    if not api_key:
        raise RuntimeError(
            "OPENCODE_API_KEY is not set. Copy .env.example to .env and fill it in "
            "(get a personal API key at https://opencode.ai/console)."
        )
    return ChatOpenAI(
        model=os.getenv("OPENCODE_MODEL", DEFAULT_MODEL),
        base_url=os.getenv("OPENCODE_BASE_URL", DEFAULT_BASE_URL),
        api_key=api_key,
        temperature=0.2,
        timeout=120,
        max_retries=1,
        default_headers=_default_headers(),
    )

"""Unit tests for the OpenCode chat-model factory (offline, no network)."""

import pytest

from app.models import DEFAULT_MODEL, DEFAULT_USER_AGENT, chat_model


@pytest.fixture
def api_key(monkeypatch):
    monkeypatch.setenv("OPENCODE_API_KEY", "test-key")


def test_model_and_base_url_read_from_env(api_key, monkeypatch):
    monkeypatch.setenv("OPENCODE_MODEL", "kimi-k3")
    monkeypatch.setenv(
        "OPENCODE_BASE_URL",
        "https://opencode.ai/inference/go/openai/v1",
    )
    model = chat_model()

    assert model.model_name == "kimi-k3"
    assert model.openai_api_base == "https://opencode.ai/inference/go/openai/v1"


def test_glm_5_3_flash_is_the_default_model(api_key):
    assert DEFAULT_MODEL == "glm-5.3-flash"
    assert chat_model().model_name == "glm-5.3-flash"


def test_required_headers_are_always_sent(api_key):
    headers = chat_model().root_client.default_headers

    session = headers.get("x-opencode-session")
    assert session  # env-provided or generated per process
    assert headers.get("User-Agent", "").startswith("langchain-study/")


def test_header_overrides_from_env(api_key, monkeypatch):
    monkeypatch.setenv("OPENCODE_SESSION_ID", "my-stable-analysis-session")
    monkeypatch.setenv("OPENCODE_USER_AGENT", "my-car-analyzer/1.0")

    headers = chat_model().root_client.default_headers

    assert headers["x-opencode-session"] == "my-stable-analysis-session"
    assert headers["User-Agent"] == "my-car-analyzer/1.0"


def test_missing_api_key_fails_fast(monkeypatch):
    monkeypatch.delenv("OPENCODE_API_KEY", raising=False)
    with pytest.raises(RuntimeError, match="OPENCODE_API_KEY"):
        chat_model()


def test_user_agent_default_is_self_identifying(api_key):
    assert DEFAULT_USER_AGENT.startswith("langchain-study/")

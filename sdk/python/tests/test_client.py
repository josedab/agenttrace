"""Tests for the AgentTrace client."""

import pytest
from unittest.mock import Mock, patch

from agenttrace import AgentTrace, observe, generation
from agenttrace.context import get_client, get_current_trace, set_client
from agenttrace.prompt import PromptVersion


class TestAgentTraceClient:
    """Tests for the AgentTrace client."""

    def test_client_initialization(self):
        """Test client initializes correctly."""
        with patch("agenttrace.client.HttpTransport"):
            with patch("agenttrace.client.BatchQueue"):
                client = AgentTrace(
                    api_key="test-key",
                    host="https://test.example.com",
                    enabled=True,
                )

                assert client.api_key == "test-key"
                assert client.host == "https://test.example.com"
                assert client.enabled is True

    def test_client_disabled(self):
        """Test client can be disabled."""
        with patch("agenttrace.client.HttpTransport"):
            with patch("agenttrace.client.BatchQueue"):
                client = AgentTrace(
                    api_key="test-key",
                    enabled=False,
                )

                assert client.enabled is False

    def test_trace_creation(self):
        """Test trace can be created."""
        with patch("agenttrace.client.HttpTransport"):
            with patch("agenttrace.client.BatchQueue") as mock_queue:
                client = AgentTrace(api_key="test-key")

                trace = client.trace(name="test-trace")

                assert trace.name == "test-trace"
                assert trace.id is not None


class TestObserveDecorator:
    """Tests for the @observe decorator."""

    def test_observe_sync_function(self):
        """Test observe decorator with sync function."""
        set_client(None)  # Clear any existing client

        @observe
        def my_function(x: int) -> int:
            return x * 2

        # Without a client, function should still work
        result = my_function(5)
        assert result == 10

    def test_observe_with_name(self):
        """Test observe decorator with custom name."""
        @observe(name="custom-name")
        def my_function():
            return "result"

        result = my_function()
        assert result == "result"

    @pytest.mark.asyncio
    async def test_observe_async_function(self):
        """Test observe decorator with async function."""
        @observe
        async def my_async_function(x: int) -> int:
            return x * 2

        result = await my_async_function(5)
        assert result == 10


class TestPromptVersion:
    """Tests for prompt functionality."""

    def test_prompt_compile(self):
        """Test prompt compilation with variables."""
        prompt = PromptVersion(
            id="test-id",
            version=1,
            prompt="Hello, {{name}}! Welcome to {{place}}.",
            config={},
            labels=[],
            created_at="2024-01-01T00:00:00Z",
        )

        result = prompt.compile(name="Alice", place="AgentTrace")
        assert result == "Hello, Alice! Welcome to AgentTrace."

    def test_prompt_compile_single_brace(self):
        """Test prompt compilation with single brace syntax."""
        prompt = PromptVersion(
            id="test-id",
            version=1,
            prompt="Hello, {name}!",
            config={},
            labels=[],
            created_at="2024-01-01T00:00:00Z",
        )

        result = prompt.compile(name="Bob")
        assert result == "Hello, Bob!"

    def test_prompt_compile_chat(self):
        """Test prompt compilation as chat messages."""
        prompt = PromptVersion(
            id="test-id",
            version=1,
            prompt="system: You are a helpful assistant.\nuser: Hello, {{name}}!",
            config={},
            labels=[],
            created_at="2024-01-01T00:00:00Z",
        )

        messages = prompt.compile_chat(name="Charlie")

        assert len(messages) == 2
        assert messages[0]["role"] == "system"
        assert messages[0]["content"] == "You are a helpful assistant."
        assert messages[1]["role"] == "user"
        assert messages[1]["content"] == "Hello, Charlie!"

    def test_prompt_get_variables(self):
        """Test extracting variables from prompt."""
        prompt = PromptVersion(
            id="test-id",
            version=1,
            prompt="Hello, {{name}}! Your score is {score}.",
            config={},
            labels=[],
            created_at="2024-01-01T00:00:00Z",
        )

        variables = prompt.get_variables()

        assert "name" in variables
        assert "score" in variables


class TestGeneration:
    """Tests for generation context manager."""

    def test_generation_context_without_client(self):
        """Test generation context manager without client."""
        set_client(None)

        with generation(name="test-gen", model="gpt-4") as gen:
            gen.update(output="test output")

        # Should complete without error


class TestSerializeValue:
    """Tests for value serialization."""

    def test_serialize_primitives(self):
        """Test serializing primitive values."""
        from agenttrace.decorators import _serialize_value

        assert _serialize_value("string") == "string"
        assert _serialize_value(42) == 42
        assert _serialize_value(3.14) == 3.14
        assert _serialize_value(True) is True
        assert _serialize_value(None) is None

    def test_serialize_list(self):
        """Test serializing lists."""
        from agenttrace.decorators import _serialize_value

        result = _serialize_value([1, "two", 3.0])
        assert result == [1, "two", 3.0]

    def test_serialize_dict(self):
        """Test serializing dictionaries."""
        from agenttrace.decorators import _serialize_value

        result = _serialize_value({"key": "value", "num": 42})
        assert result == {"key": "value", "num": 42}

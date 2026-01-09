"""
OpenAI auto-instrumentation for AgentTrace.

Automatically traces OpenAI API calls when enabled.
"""

from __future__ import annotations

import functools
import logging
from typing import Any, Callable, Dict, List, Optional, TypeVar

from agenttrace.context import get_client, get_current_trace, get_current_observation
from agenttrace.generation import start_generation

logger = logging.getLogger("agenttrace")

F = TypeVar("F", bound=Callable[..., Any])


class OpenAIInstrumentation:
    """
    Auto-instrumentation for OpenAI SDK.

    Example:
        from agenttrace import AgentTrace
        from agenttrace.integrations import OpenAIInstrumentation

        client = AgentTrace(api_key="...")
        OpenAIInstrumentation.enable()

        # Now all OpenAI calls are automatically traced
        import openai
        response = openai.chat.completions.create(...)
    """

    _original_create: Optional[Callable] = None
    _original_acreate: Optional[Callable] = None
    _original_completions_create: Optional[Callable] = None
    _original_completions_acreate: Optional[Callable] = None
    _enabled: bool = False

    @classmethod
    def enable(cls) -> None:
        """Enable OpenAI auto-instrumentation."""
        if cls._enabled:
            return

        try:
            import openai
        except ImportError:
            logger.warning("openai package not installed, skipping instrumentation")
            return

        cls._patch_chat_completions(openai)
        cls._patch_completions(openai)
        cls._enabled = True
        logger.debug("OpenAI instrumentation enabled")

    @classmethod
    def disable(cls) -> None:
        """Disable OpenAI auto-instrumentation."""
        if not cls._enabled:
            return

        try:
            import openai
            cls._unpatch_chat_completions(openai)
            cls._unpatch_completions(openai)
        except ImportError:
            pass

        cls._enabled = False
        logger.debug("OpenAI instrumentation disabled")

    @classmethod
    def _patch_chat_completions(cls, openai) -> None:
        """Patch chat completions API."""
        try:
            # OpenAI v1.x
            if hasattr(openai, "resources"):
                chat_module = openai.resources.chat.completions
                if hasattr(chat_module, "Completions"):
                    completions_cls = chat_module.Completions

                    # Save original
                    cls._original_create = completions_cls.create

                    # Create wrapped version
                    @functools.wraps(cls._original_create)
                    def wrapped_create(self, *args, **kwargs):
                        return cls._trace_chat_completion(
                            cls._original_create, self, *args, **kwargs
                        )

                    completions_cls.create = wrapped_create

                    # Async version
                    if hasattr(chat_module, "AsyncCompletions"):
                        async_completions_cls = chat_module.AsyncCompletions
                        cls._original_acreate = async_completions_cls.create

                        @functools.wraps(cls._original_acreate)
                        async def wrapped_acreate(self, *args, **kwargs):
                            return await cls._trace_chat_completion_async(
                                cls._original_acreate, self, *args, **kwargs
                            )

                        async_completions_cls.create = wrapped_acreate

        except Exception as e:
            logger.warning(f"Failed to patch chat completions: {e}")

    @classmethod
    def _unpatch_chat_completions(cls, openai) -> None:
        """Restore original chat completions API."""
        try:
            if hasattr(openai, "resources") and cls._original_create:
                chat_module = openai.resources.chat.completions
                if hasattr(chat_module, "Completions"):
                    chat_module.Completions.create = cls._original_create
                if hasattr(chat_module, "AsyncCompletions") and cls._original_acreate:
                    chat_module.AsyncCompletions.create = cls._original_acreate
        except Exception as e:
            logger.warning(f"Failed to unpatch chat completions: {e}")

    @classmethod
    def _patch_completions(cls, openai) -> None:
        """Patch legacy completions API."""
        try:
            if hasattr(openai, "resources") and hasattr(openai.resources, "completions"):
                completions_module = openai.resources.completions
                if hasattr(completions_module, "Completions"):
                    completions_cls = completions_module.Completions
                    cls._original_completions_create = completions_cls.create

                    @functools.wraps(cls._original_completions_create)
                    def wrapped_create(self, *args, **kwargs):
                        return cls._trace_completion(
                            cls._original_completions_create, self, *args, **kwargs
                        )

                    completions_cls.create = wrapped_create
        except Exception as e:
            logger.debug(f"Legacy completions not available: {e}")

    @classmethod
    def _unpatch_completions(cls, openai) -> None:
        """Restore original completions API."""
        try:
            if cls._original_completions_create and hasattr(openai, "resources"):
                openai.resources.completions.Completions.create = cls._original_completions_create
        except Exception as e:
            logger.debug(f"Failed to unpatch completions: {e}")

    @classmethod
    def _trace_chat_completion(
        cls,
        original_fn: Callable,
        instance: Any,
        *args,
        **kwargs,
    ) -> Any:
        """Trace a chat completion call."""
        client = get_client()
        if client is None or not client.enabled:
            return original_fn(instance, *args, **kwargs)

        model = kwargs.get("model", "unknown")
        messages = kwargs.get("messages", [])

        # Extract model parameters
        model_params = {}
        for param in ["temperature", "max_tokens", "top_p", "frequency_penalty",
                      "presence_penalty", "stop", "seed"]:
            if param in kwargs:
                model_params[param] = kwargs[param]

        gen = start_generation(
            name="openai.chat.completions.create",
            model=model,
            model_parameters=model_params,
            input={"messages": messages},
        )

        try:
            response = original_fn(instance, *args, **kwargs)

            # Handle streaming responses
            if kwargs.get("stream", False):
                return cls._wrap_stream(response, gen)

            # Extract output and usage
            output = cls._extract_chat_output(response)
            usage = cls._extract_usage(response)

            gen.end(output=output, usage=usage, model=getattr(response, "model", model))
            return response

        except Exception as e:
            gen.end(output={"error": str(e)})
            raise

    @classmethod
    async def _trace_chat_completion_async(
        cls,
        original_fn: Callable,
        instance: Any,
        *args,
        **kwargs,
    ) -> Any:
        """Trace an async chat completion call."""
        client = get_client()
        if client is None or not client.enabled:
            return await original_fn(instance, *args, **kwargs)

        model = kwargs.get("model", "unknown")
        messages = kwargs.get("messages", [])

        model_params = {}
        for param in ["temperature", "max_tokens", "top_p", "frequency_penalty",
                      "presence_penalty", "stop", "seed"]:
            if param in kwargs:
                model_params[param] = kwargs[param]

        gen = start_generation(
            name="openai.chat.completions.create",
            model=model,
            model_parameters=model_params,
            input={"messages": messages},
        )

        try:
            response = await original_fn(instance, *args, **kwargs)

            if kwargs.get("stream", False):
                return cls._wrap_stream_async(response, gen)

            output = cls._extract_chat_output(response)
            usage = cls._extract_usage(response)

            gen.end(output=output, usage=usage, model=getattr(response, "model", model))
            return response

        except Exception as e:
            gen.end(output={"error": str(e)})
            raise

    @classmethod
    def _trace_completion(
        cls,
        original_fn: Callable,
        instance: Any,
        *args,
        **kwargs,
    ) -> Any:
        """Trace a legacy completion call."""
        client = get_client()
        if client is None or not client.enabled:
            return original_fn(instance, *args, **kwargs)

        model = kwargs.get("model", "unknown")
        prompt = kwargs.get("prompt", "")

        model_params = {}
        for param in ["temperature", "max_tokens", "top_p", "frequency_penalty",
                      "presence_penalty", "stop"]:
            if param in kwargs:
                model_params[param] = kwargs[param]

        gen = start_generation(
            name="openai.completions.create",
            model=model,
            model_parameters=model_params,
            input={"prompt": prompt},
        )

        try:
            response = original_fn(instance, *args, **kwargs)

            output = cls._extract_completion_output(response)
            usage = cls._extract_usage(response)

            gen.end(output=output, usage=usage, model=getattr(response, "model", model))
            return response

        except Exception as e:
            gen.end(output={"error": str(e)})
            raise

    @classmethod
    def _extract_chat_output(cls, response) -> Dict[str, Any]:
        """Extract output from chat completion response."""
        try:
            choices = getattr(response, "choices", [])
            if choices:
                message = getattr(choices[0], "message", None)
                if message:
                    return {
                        "role": getattr(message, "role", "assistant"),
                        "content": getattr(message, "content", None),
                        "tool_calls": cls._extract_tool_calls(message),
                    }
            return {}
        except Exception:
            return {}

    @classmethod
    def _extract_tool_calls(cls, message) -> Optional[List[Dict[str, Any]]]:
        """Extract tool calls from message."""
        tool_calls = getattr(message, "tool_calls", None)
        if not tool_calls:
            return None

        return [
            {
                "id": getattr(tc, "id", ""),
                "type": getattr(tc, "type", "function"),
                "function": {
                    "name": getattr(getattr(tc, "function", None), "name", ""),
                    "arguments": getattr(getattr(tc, "function", None), "arguments", ""),
                },
            }
            for tc in tool_calls
        ]

    @classmethod
    def _extract_completion_output(cls, response) -> Dict[str, Any]:
        """Extract output from legacy completion response."""
        try:
            choices = getattr(response, "choices", [])
            if choices:
                return {"text": getattr(choices[0], "text", "")}
            return {}
        except Exception:
            return {}

    @classmethod
    def _extract_usage(cls, response) -> Optional[Dict[str, int]]:
        """Extract usage from response."""
        try:
            usage = getattr(response, "usage", None)
            if usage:
                return {
                    "input_tokens": getattr(usage, "prompt_tokens", 0),
                    "output_tokens": getattr(usage, "completion_tokens", 0),
                    "total_tokens": getattr(usage, "total_tokens", 0),
                }
            return None
        except Exception:
            return None

    @classmethod
    def _wrap_stream(cls, stream, gen):
        """Wrap a streaming response to capture output."""
        chunks = []
        role = "assistant"
        tool_calls: List[Dict[str, Any]] = []

        try:
            for chunk in stream:
                chunks.append(chunk)
                yield chunk

            # After stream completes, extract the full response
            content = cls._combine_stream_chunks(chunks)
            output = {"role": role, "content": content}
            if tool_calls:
                output["tool_calls"] = tool_calls

            # Try to get usage from the last chunk
            usage = None
            if chunks:
                usage = cls._extract_usage(chunks[-1])

            gen.end(output=output, usage=usage)

        except Exception as e:
            gen.end(output={"error": str(e)})
            raise

    @classmethod
    async def _wrap_stream_async(cls, stream, gen):
        """Wrap an async streaming response."""
        chunks = []

        try:
            async for chunk in stream:
                chunks.append(chunk)
                yield chunk

            content = cls._combine_stream_chunks(chunks)
            output = {"role": "assistant", "content": content}

            usage = None
            if chunks:
                usage = cls._extract_usage(chunks[-1])

            gen.end(output=output, usage=usage)

        except Exception as e:
            gen.end(output={"error": str(e)})
            raise

    @classmethod
    def _combine_stream_chunks(cls, chunks) -> str:
        """Combine streaming chunks into complete content."""
        content_parts = []
        for chunk in chunks:
            try:
                choices = getattr(chunk, "choices", [])
                if choices:
                    delta = getattr(choices[0], "delta", None)
                    if delta:
                        content = getattr(delta, "content", None)
                        if content:
                            content_parts.append(content)
            except Exception:
                pass
        return "".join(content_parts)

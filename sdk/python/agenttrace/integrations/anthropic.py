"""
Anthropic auto-instrumentation for AgentTrace.

Automatically traces Anthropic API calls when enabled.
"""

from __future__ import annotations

import functools
import logging
from typing import Any, Callable, Dict, List, Optional, TypeVar

from agenttrace.context import get_client
from agenttrace.generation import start_generation

logger = logging.getLogger("agenttrace")

F = TypeVar("F", bound=Callable[..., Any])


class AnthropicInstrumentation:
    """
    Auto-instrumentation for Anthropic SDK.

    Example:
        from agenttrace import AgentTrace
        from agenttrace.integrations import AnthropicInstrumentation

        client = AgentTrace(api_key="...")
        AnthropicInstrumentation.enable()

        # Now all Anthropic calls are automatically traced
        import anthropic
        client = anthropic.Anthropic()
        response = client.messages.create(...)
    """

    _original_create: Optional[Callable] = None
    _original_acreate: Optional[Callable] = None
    _enabled: bool = False

    @classmethod
    def enable(cls) -> None:
        """Enable Anthropic auto-instrumentation."""
        if cls._enabled:
            return

        try:
            import anthropic
        except ImportError:
            logger.warning("anthropic package not installed, skipping instrumentation")
            return

        cls._patch_messages(anthropic)
        cls._enabled = True
        logger.debug("Anthropic instrumentation enabled")

    @classmethod
    def disable(cls) -> None:
        """Disable Anthropic auto-instrumentation."""
        if not cls._enabled:
            return

        try:
            import anthropic
            cls._unpatch_messages(anthropic)
        except ImportError:
            pass

        cls._enabled = False
        logger.debug("Anthropic instrumentation disabled")

    @classmethod
    def _patch_messages(cls, anthropic) -> None:
        """Patch messages API."""
        try:
            # Anthropic SDK structure
            if hasattr(anthropic, "resources") and hasattr(anthropic.resources, "messages"):
                messages_module = anthropic.resources.messages

                if hasattr(messages_module, "Messages"):
                    messages_cls = messages_module.Messages

                    # Save original
                    cls._original_create = messages_cls.create

                    # Create wrapped version
                    @functools.wraps(cls._original_create)
                    def wrapped_create(self, *args, **kwargs):
                        return cls._trace_message(
                            cls._original_create, self, *args, **kwargs
                        )

                    messages_cls.create = wrapped_create

                # Async version
                if hasattr(messages_module, "AsyncMessages"):
                    async_messages_cls = messages_module.AsyncMessages
                    cls._original_acreate = async_messages_cls.create

                    @functools.wraps(cls._original_acreate)
                    async def wrapped_acreate(self, *args, **kwargs):
                        return await cls._trace_message_async(
                            cls._original_acreate, self, *args, **kwargs
                        )

                    async_messages_cls.create = wrapped_acreate

        except Exception as e:
            logger.warning(f"Failed to patch Anthropic messages: {e}")

    @classmethod
    def _unpatch_messages(cls, anthropic) -> None:
        """Restore original messages API."""
        try:
            if hasattr(anthropic, "resources") and hasattr(anthropic.resources, "messages"):
                messages_module = anthropic.resources.messages

                if cls._original_create and hasattr(messages_module, "Messages"):
                    messages_module.Messages.create = cls._original_create

                if cls._original_acreate and hasattr(messages_module, "AsyncMessages"):
                    messages_module.AsyncMessages.create = cls._original_acreate
        except Exception as e:
            logger.warning(f"Failed to unpatch Anthropic messages: {e}")

    @classmethod
    def _trace_message(
        cls,
        original_fn: Callable,
        instance: Any,
        *args,
        **kwargs,
    ) -> Any:
        """Trace a message creation call."""
        client = get_client()
        if client is None or not client.enabled:
            return original_fn(instance, *args, **kwargs)

        model = kwargs.get("model", "unknown")
        messages = kwargs.get("messages", [])
        system = kwargs.get("system")

        # Build input
        input_data: Dict[str, Any] = {"messages": messages}
        if system:
            input_data["system"] = system

        # Extract model parameters
        model_params = {}
        for param in ["max_tokens", "temperature", "top_p", "top_k", "stop_sequences"]:
            if param in kwargs:
                model_params[param] = kwargs[param]

        gen = start_generation(
            name="anthropic.messages.create",
            model=model,
            model_parameters=model_params,
            input=input_data,
        )

        try:
            response = original_fn(instance, *args, **kwargs)

            # Handle streaming
            if kwargs.get("stream", False):
                return cls._wrap_stream(response, gen, model)

            # Extract output and usage
            output = cls._extract_output(response)
            usage = cls._extract_usage(response)

            gen.end(output=output, usage=usage, model=getattr(response, "model", model))
            return response

        except Exception as e:
            gen.end(output={"error": str(e)})
            raise

    @classmethod
    async def _trace_message_async(
        cls,
        original_fn: Callable,
        instance: Any,
        *args,
        **kwargs,
    ) -> Any:
        """Trace an async message creation call."""
        client = get_client()
        if client is None or not client.enabled:
            return await original_fn(instance, *args, **kwargs)

        model = kwargs.get("model", "unknown")
        messages = kwargs.get("messages", [])
        system = kwargs.get("system")

        input_data: Dict[str, Any] = {"messages": messages}
        if system:
            input_data["system"] = system

        model_params = {}
        for param in ["max_tokens", "temperature", "top_p", "top_k", "stop_sequences"]:
            if param in kwargs:
                model_params[param] = kwargs[param]

        gen = start_generation(
            name="anthropic.messages.create",
            model=model,
            model_parameters=model_params,
            input=input_data,
        )

        try:
            response = await original_fn(instance, *args, **kwargs)

            if kwargs.get("stream", False):
                return cls._wrap_stream_async(response, gen, model)

            output = cls._extract_output(response)
            usage = cls._extract_usage(response)

            gen.end(output=output, usage=usage, model=getattr(response, "model", model))
            return response

        except Exception as e:
            gen.end(output={"error": str(e)})
            raise

    @classmethod
    def _extract_output(cls, response) -> Dict[str, Any]:
        """Extract output from message response."""
        try:
            content = getattr(response, "content", [])
            stop_reason = getattr(response, "stop_reason", None)

            # Convert content blocks to serializable format
            output_content = []
            for block in content:
                block_type = getattr(block, "type", "text")
                if block_type == "text":
                    output_content.append({
                        "type": "text",
                        "text": getattr(block, "text", ""),
                    })
                elif block_type == "tool_use":
                    output_content.append({
                        "type": "tool_use",
                        "id": getattr(block, "id", ""),
                        "name": getattr(block, "name", ""),
                        "input": getattr(block, "input", {}),
                    })

            return {
                "content": output_content,
                "stop_reason": stop_reason,
            }
        except Exception:
            return {}

    @classmethod
    def _extract_usage(cls, response) -> Optional[Dict[str, int]]:
        """Extract usage from response."""
        try:
            usage = getattr(response, "usage", None)
            if usage:
                return {
                    "input_tokens": getattr(usage, "input_tokens", 0),
                    "output_tokens": getattr(usage, "output_tokens", 0),
                    "total_tokens": (
                        getattr(usage, "input_tokens", 0) +
                        getattr(usage, "output_tokens", 0)
                    ),
                }
            return None
        except Exception:
            return None

    @classmethod
    def _wrap_stream(cls, stream, gen, model: str):
        """Wrap a streaming response to capture output."""
        content_parts: List[str] = []
        tool_use_blocks: List[Dict[str, Any]] = []
        input_tokens = 0
        output_tokens = 0
        stop_reason = None

        try:
            for event in stream:
                yield event

                # Process different event types
                event_type = getattr(event, "type", "")

                if event_type == "message_start":
                    message = getattr(event, "message", None)
                    if message:
                        usage = getattr(message, "usage", None)
                        if usage:
                            input_tokens = getattr(usage, "input_tokens", 0)

                elif event_type == "content_block_delta":
                    delta = getattr(event, "delta", None)
                    if delta:
                        delta_type = getattr(delta, "type", "")
                        if delta_type == "text_delta":
                            text = getattr(delta, "text", "")
                            content_parts.append(text)
                        elif delta_type == "input_json_delta":
                            # Tool use
                            pass

                elif event_type == "message_delta":
                    delta = getattr(event, "delta", None)
                    if delta:
                        stop_reason = getattr(delta, "stop_reason", stop_reason)
                    usage = getattr(event, "usage", None)
                    if usage:
                        output_tokens = getattr(usage, "output_tokens", 0)

            # Build final output
            output_content = []
            if content_parts:
                output_content.append({
                    "type": "text",
                    "text": "".join(content_parts),
                })
            output_content.extend(tool_use_blocks)

            output = {
                "content": output_content,
                "stop_reason": stop_reason,
            }

            usage = {
                "input_tokens": input_tokens,
                "output_tokens": output_tokens,
                "total_tokens": input_tokens + output_tokens,
            } if input_tokens or output_tokens else None

            gen.end(output=output, usage=usage, model=model)

        except Exception as e:
            gen.end(output={"error": str(e)})
            raise

    @classmethod
    async def _wrap_stream_async(cls, stream, gen, model: str):
        """Wrap an async streaming response."""
        content_parts: List[str] = []
        input_tokens = 0
        output_tokens = 0
        stop_reason = None

        try:
            async for event in stream:
                yield event

                event_type = getattr(event, "type", "")

                if event_type == "message_start":
                    message = getattr(event, "message", None)
                    if message:
                        usage = getattr(message, "usage", None)
                        if usage:
                            input_tokens = getattr(usage, "input_tokens", 0)

                elif event_type == "content_block_delta":
                    delta = getattr(event, "delta", None)
                    if delta:
                        delta_type = getattr(delta, "type", "")
                        if delta_type == "text_delta":
                            text = getattr(delta, "text", "")
                            content_parts.append(text)

                elif event_type == "message_delta":
                    delta = getattr(event, "delta", None)
                    if delta:
                        stop_reason = getattr(delta, "stop_reason", stop_reason)
                    usage = getattr(event, "usage", None)
                    if usage:
                        output_tokens = getattr(usage, "output_tokens", 0)

            output_content = []
            if content_parts:
                output_content.append({
                    "type": "text",
                    "text": "".join(content_parts),
                })

            output = {
                "content": output_content,
                "stop_reason": stop_reason,
            }

            usage = {
                "input_tokens": input_tokens,
                "output_tokens": output_tokens,
                "total_tokens": input_tokens + output_tokens,
            } if input_tokens or output_tokens else None

            gen.end(output=output, usage=usage, model=model)

        except Exception as e:
            gen.end(output={"error": str(e)})
            raise

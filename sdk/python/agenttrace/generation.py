"""
Generation context manager for tracking LLM calls.
"""

from __future__ import annotations

from contextlib import contextmanager, asynccontextmanager
from datetime import datetime
from typing import Any, Dict, Optional, Generator, AsyncGenerator

from agenttrace.context import (
    get_client,
    get_current_trace,
    get_current_observation,
    set_current_observation,
)


class GenerationContext:
    """
    Represents an active generation within the context manager.

    Provides methods to update the generation with output, usage, etc.
    """

    def __init__(
        self,
        generation,
        prev_observation,
    ):
        self._generation = generation
        self._prev_observation = prev_observation
        self._ended = False

    @property
    def id(self) -> str:
        """Get the generation ID."""
        return self._generation.id if self._generation else ""

    @property
    def trace_id(self) -> str:
        """Get the trace ID."""
        return self._generation.trace_id if self._generation else ""

    def update(
        self,
        output: Optional[Any] = None,
        usage: Optional[Dict[str, int]] = None,
        model: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None,
    ) -> "GenerationContext":
        """
        Update the generation with additional data.

        Args:
            output: The output/response from the LLM
            usage: Token usage (input_tokens, output_tokens, total_tokens)
            model: Model name/identifier (if changed)
            metadata: Additional metadata

        Returns:
            Self for chaining
        """
        if self._generation and not self._ended:
            if output is not None:
                self._generation.output = output
            if usage is not None:
                self._generation.usage = usage
            if model is not None:
                self._generation.model = model
            if metadata is not None:
                self._generation.metadata.update(metadata)

        return self

    def end(
        self,
        output: Optional[Any] = None,
        usage: Optional[Dict[str, int]] = None,
        model: Optional[str] = None,
    ) -> None:
        """
        End the generation.

        Args:
            output: The final output
            usage: Token usage
            model: Model used
        """
        if self._generation and not self._ended:
            self._generation.end(output=output, usage=usage, model=model)
            self._ended = True
            set_current_observation(self._prev_observation)


@contextmanager
def generation(
    name: str,
    model: Optional[str] = None,
    model_parameters: Optional[Dict[str, Any]] = None,
    input: Optional[Any] = None,
    metadata: Optional[Dict[str, Any]] = None,
    level: str = "DEFAULT",
) -> Generator[GenerationContext, None, None]:
    """
    Context manager for tracking an LLM generation.

    Example:
        with generation(name="chat", model="gpt-4", input=messages) as gen:
            response = openai.chat.completions.create(
                model="gpt-4",
                messages=messages,
            )
            gen.update(
                output=response.choices[0].message.content,
                usage={
                    "input_tokens": response.usage.prompt_tokens,
                    "output_tokens": response.usage.completion_tokens,
                }
            )

    Args:
        name: Name of the generation
        model: Model name/identifier
        model_parameters: Model parameters (temperature, max_tokens, etc.)
        input: Input prompt or messages
        metadata: Additional metadata
        level: Log level (DEBUG, DEFAULT, WARNING, ERROR)

    Yields:
        GenerationContext object for updating the generation
    """
    client = get_client()
    if client is None or not client.enabled:
        # Return a no-op context
        yield GenerationContext(None, None)
        return

    trace = get_current_trace()
    if trace is None:
        # Create a new trace for this generation
        trace = client.trace(name=name, input=input)

    parent_observation = get_current_observation()
    parent_id = parent_observation.id if parent_observation else None

    gen = trace.generation(
        name=name,
        model=model,
        model_parameters=model_parameters,
        input=input,
        metadata=metadata,
        parent_observation_id=parent_id,
        level=level,
    )

    # Set as current observation
    set_current_observation(gen)

    ctx = GenerationContext(gen, parent_observation)
    try:
        yield ctx
    except Exception as e:
        # Record error and re-raise
        if not ctx._ended:
            gen.end(output={"error": str(e)})
            ctx._ended = True
            set_current_observation(parent_observation)
        raise
    finally:
        # Ensure generation is ended
        if not ctx._ended:
            gen.end()
            set_current_observation(parent_observation)


@asynccontextmanager
async def ageneration(
    name: str,
    model: Optional[str] = None,
    model_parameters: Optional[Dict[str, Any]] = None,
    input: Optional[Any] = None,
    metadata: Optional[Dict[str, Any]] = None,
    level: str = "DEFAULT",
) -> AsyncGenerator[GenerationContext, None]:
    """
    Async context manager for tracking an LLM generation.

    Example:
        async with ageneration(name="chat", model="gpt-4", input=messages) as gen:
            response = await openai.chat.completions.create(
                model="gpt-4",
                messages=messages,
            )
            gen.update(
                output=response.choices[0].message.content,
                usage={
                    "input_tokens": response.usage.prompt_tokens,
                    "output_tokens": response.usage.completion_tokens,
                }
            )

    Args:
        name: Name of the generation
        model: Model name/identifier
        model_parameters: Model parameters
        input: Input prompt or messages
        metadata: Additional metadata
        level: Log level

    Yields:
        GenerationContext object for updating the generation
    """
    client = get_client()
    if client is None or not client.enabled:
        yield GenerationContext(None, None)
        return

    trace = get_current_trace()
    if trace is None:
        trace = client.trace(name=name, input=input)

    parent_observation = get_current_observation()
    parent_id = parent_observation.id if parent_observation else None

    gen = trace.generation(
        name=name,
        model=model,
        model_parameters=model_parameters,
        input=input,
        metadata=metadata,
        parent_observation_id=parent_id,
        level=level,
    )

    set_current_observation(gen)

    ctx = GenerationContext(gen, parent_observation)
    try:
        yield ctx
    except Exception as e:
        if not ctx._ended:
            gen.end(output={"error": str(e)})
            ctx._ended = True
            set_current_observation(parent_observation)
        raise
    finally:
        if not ctx._ended:
            gen.end()
            set_current_observation(parent_observation)


def start_generation(
    name: str,
    model: Optional[str] = None,
    model_parameters: Optional[Dict[str, Any]] = None,
    input: Optional[Any] = None,
    metadata: Optional[Dict[str, Any]] = None,
    level: str = "DEFAULT",
) -> GenerationContext:
    """
    Start a generation without a context manager.

    This is useful when you need more control over when to end the generation.
    Remember to call .end() when done!

    Example:
        gen = start_generation(name="chat", model="gpt-4", input=messages)
        try:
            response = openai.chat.completions.create(...)
            gen.end(
                output=response.choices[0].message.content,
                usage={"input_tokens": ..., "output_tokens": ...}
            )
        except Exception as e:
            gen.end(output={"error": str(e)})
            raise

    Args:
        name: Name of the generation
        model: Model name/identifier
        model_parameters: Model parameters
        input: Input prompt or messages
        metadata: Additional metadata
        level: Log level

    Returns:
        GenerationContext object
    """
    client = get_client()
    if client is None or not client.enabled:
        return GenerationContext(None, None)

    trace = get_current_trace()
    if trace is None:
        trace = client.trace(name=name, input=input)

    parent_observation = get_current_observation()
    parent_id = parent_observation.id if parent_observation else None

    gen = trace.generation(
        name=name,
        model=model,
        model_parameters=model_parameters,
        input=input,
        metadata=metadata,
        parent_observation_id=parent_id,
        level=level,
    )

    set_current_observation(gen)

    return GenerationContext(gen, parent_observation)

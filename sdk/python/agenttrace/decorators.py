"""
Decorators for automatic tracing.
"""

from __future__ import annotations

import asyncio
import functools
import inspect
from typing import Any, Callable, Optional, TypeVar, Union, overload

from agenttrace.context import get_client, get_current_observation, get_current_trace

F = TypeVar("F", bound=Callable[..., Any])


@overload
def observe(func: F) -> F: ...


@overload
def observe(
    *,
    name: Optional[str] = None,
    as_type: Optional[str] = None,
    capture_input: bool = True,
    capture_output: bool = True,
) -> Callable[[F], F]: ...


def observe(
    func: Optional[F] = None,
    *,
    name: Optional[str] = None,
    as_type: Optional[str] = None,
    capture_input: bool = True,
    capture_output: bool = True,
) -> Union[F, Callable[[F], F]]:
    """
    Decorator to automatically trace a function.

    Can be used with or without arguments:

        @observe
        def my_function():
            pass

        @observe(name="custom-name", as_type="generation")
        def my_llm_function():
            pass

    Args:
        name: Custom name for the observation (defaults to function name)
        as_type: Type of observation ("span" or "generation")
        capture_input: Whether to capture function arguments
        capture_output: Whether to capture function return value

    Returns:
        Decorated function
    """

    def decorator(fn: F) -> F:
        observation_name = name or fn.__name__
        observation_type = as_type or "span"

        if asyncio.iscoroutinefunction(fn):
            @functools.wraps(fn)
            async def async_wrapper(*args: Any, **kwargs: Any) -> Any:
                return await _trace_async(
                    fn,
                    args,
                    kwargs,
                    observation_name,
                    observation_type,
                    capture_input,
                    capture_output,
                )

            return async_wrapper  # type: ignore
        else:
            @functools.wraps(fn)
            def sync_wrapper(*args: Any, **kwargs: Any) -> Any:
                return _trace_sync(
                    fn,
                    args,
                    kwargs,
                    observation_name,
                    observation_type,
                    capture_input,
                    capture_output,
                )

            return sync_wrapper  # type: ignore

    if func is not None:
        return decorator(func)

    return decorator


def _trace_sync(
    fn: Callable[..., Any],
    args: tuple[Any, ...],
    kwargs: dict[str, Any],
    name: str,
    observation_type: str,
    capture_input: bool,
    capture_output: bool,
) -> Any:
    """Synchronous tracing wrapper."""
    client = get_client()
    if client is None or not client.enabled:
        return fn(*args, **kwargs)

    # Get or create trace
    trace = get_current_trace()
    parent_observation = get_current_observation()

    # Prepare input
    input_data = None
    if capture_input:
        input_data = _capture_args(fn, args, kwargs)

    # Create observation
    if trace is None:
        # Create a new trace
        trace = client.trace(name=name, input=input_data)
        observation = None
    else:
        parent_id = parent_observation.id if parent_observation else None
        if observation_type == "generation":
            observation = trace.generation(
                name=name,
                input=input_data,
                parent_observation_id=parent_id,
            )
        else:
            observation = trace.span(
                name=name,
                input=input_data,
                parent_observation_id=parent_id,
            )

    try:
        # Execute the function
        result = fn(*args, **kwargs)

        # Capture output
        output = result if capture_output else None

        if observation:
            observation.end(output=output)
        else:
            trace.end(output=output)

        return result

    except Exception as e:
        # Record error
        if observation:
            observation.end(output={"error": str(e)})
        else:
            trace.end(output={"error": str(e)})
        raise


async def _trace_async(
    fn: Callable[..., Any],
    args: tuple[Any, ...],
    kwargs: dict[str, Any],
    name: str,
    observation_type: str,
    capture_input: bool,
    capture_output: bool,
) -> Any:
    """Asynchronous tracing wrapper."""
    client = get_client()
    if client is None or not client.enabled:
        return await fn(*args, **kwargs)

    trace = get_current_trace()
    parent_observation = get_current_observation()

    input_data = None
    if capture_input:
        input_data = _capture_args(fn, args, kwargs)

    if trace is None:
        trace = client.trace(name=name, input=input_data)
        observation = None
    else:
        parent_id = parent_observation.id if parent_observation else None
        if observation_type == "generation":
            observation = trace.generation(
                name=name,
                input=input_data,
                parent_observation_id=parent_id,
            )
        else:
            observation = trace.span(
                name=name,
                input=input_data,
                parent_observation_id=parent_id,
            )

    try:
        result = await fn(*args, **kwargs)

        output = result if capture_output else None

        if observation:
            observation.end(output=output)
        else:
            trace.end(output=output)

        return result

    except Exception as e:
        if observation:
            observation.end(output={"error": str(e)})
        else:
            trace.end(output={"error": str(e)})
        raise


def _capture_args(
    fn: Callable[..., Any],
    args: tuple[Any, ...],
    kwargs: dict[str, Any],
) -> dict[str, Any]:
    """Capture function arguments as a dictionary."""
    try:
        sig = inspect.signature(fn)
        bound = sig.bind(*args, **kwargs)
        bound.apply_defaults()

        # Filter out self/cls
        result = {}
        for key, value in bound.arguments.items():
            if key not in ("self", "cls"):
                result[key] = _serialize_value(value)

        return result
    except Exception:
        # Fallback to simple representation
        return {
            "args": [_serialize_value(a) for a in args],
            "kwargs": {k: _serialize_value(v) for k, v in kwargs.items()},
        }


def _serialize_value(value: Any) -> Any:
    """Serialize a value for JSON encoding."""
    if isinstance(value, (str, int, float, bool, type(None))):
        return value
    if isinstance(value, (list, tuple)):
        return [_serialize_value(v) for v in value]
    if isinstance(value, dict):
        return {k: _serialize_value(v) for k, v in value.items()}
    try:
        return str(value)
    except Exception:
        return "<unserializable>"

"""
Context propagation for tracing.
"""

from __future__ import annotations

import contextvars
from typing import TYPE_CHECKING, Optional

if TYPE_CHECKING:
    from agenttrace.client import AgentTrace, Trace, Span, Generation

# Context variables for thread-safe and async-safe context propagation
_client_var: contextvars.ContextVar[Optional["AgentTrace"]] = contextvars.ContextVar(
    "agenttrace_client", default=None
)
_trace_var: contextvars.ContextVar[Optional["Trace"]] = contextvars.ContextVar(
    "agenttrace_trace", default=None
)
_observation_var: contextvars.ContextVar[Optional["Span | Generation"]] = contextvars.ContextVar(
    "agenttrace_observation", default=None
)


def set_client(client: Optional["AgentTrace"]) -> None:
    """Set the global AgentTrace client."""
    _client_var.set(client)


def get_client() -> Optional["AgentTrace"]:
    """Get the global AgentTrace client."""
    return _client_var.get()


def set_current_trace(trace: Optional["Trace"]) -> None:
    """Set the current trace in context."""
    _trace_var.set(trace)


def get_current_trace() -> Optional["Trace"]:
    """Get the current trace from context."""
    return _trace_var.get()


def set_current_observation(observation: Optional["Span | Generation"]) -> None:
    """Set the current observation in context."""
    _observation_var.set(observation)


def get_current_observation() -> Optional["Span | Generation"]:
    """Get the current observation from context."""
    return _observation_var.get()


class TraceContext:
    """
    Context manager for scoping a trace.

    Example:
        with TraceContext(trace):
            # Code here runs with trace as current context
            pass
    """

    def __init__(self, trace: "Trace"):
        self.trace = trace
        self._token: Optional[contextvars.Token] = None

    def __enter__(self) -> "Trace":
        self._token = _trace_var.set(self.trace)
        return self.trace

    def __exit__(self, exc_type, exc_val, exc_tb) -> None:
        if self._token is not None:
            _trace_var.reset(self._token)


class ObservationContext:
    """
    Context manager for scoping an observation.

    Example:
        with ObservationContext(span):
            # Code here runs with span as current observation
            pass
    """

    def __init__(self, observation: "Span | Generation"):
        self.observation = observation
        self._token: Optional[contextvars.Token] = None

    def __enter__(self) -> "Span | Generation":
        self._token = _observation_var.set(self.observation)
        return self.observation

    def __exit__(self, exc_type, exc_val, exc_tb) -> None:
        if self._token is not None:
            _observation_var.reset(self._token)


def run_in_trace_context(trace: "Trace", func, *args, **kwargs):
    """
    Run a function within a trace context.

    This is useful for running callbacks or parallel tasks while
    maintaining the trace context.

    Example:
        def my_callback():
            # This runs in the trace context
            current = get_current_trace()

        run_in_trace_context(trace, my_callback)
    """
    token = _trace_var.set(trace)
    try:
        return func(*args, **kwargs)
    finally:
        _trace_var.reset(token)


async def run_in_trace_context_async(trace: "Trace", coro):
    """
    Run an async coroutine within a trace context.

    Example:
        async def my_async_task():
            current = get_current_trace()

        await run_in_trace_context_async(trace, my_async_task())
    """
    token = _trace_var.set(trace)
    try:
        return await coro
    finally:
        _trace_var.reset(token)

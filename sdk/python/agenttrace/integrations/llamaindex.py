"""
LlamaIndex auto-instrumentation for AgentTrace.

Automatically traces LlamaIndex LLM calls, queries, and retrievals when enabled.
"""

from __future__ import annotations

import logging
import uuid
from typing import Any, Dict, List, Optional
from datetime import datetime

from agenttrace.context import get_client
from agenttrace.span import start_span
from agenttrace.generation import start_generation

logger = logging.getLogger("agenttrace")


class LlamaIndexInstrumentation:
    """
    Auto-instrumentation for LlamaIndex.

    Example:
        from agenttrace import AgentTrace
        from agenttrace.integrations import LlamaIndexInstrumentation

        client = AgentTrace(api_key="...")
        LlamaIndexInstrumentation.enable()

        # Now all LlamaIndex calls are automatically traced
        from llama_index.core import VectorStoreIndex
        index = VectorStoreIndex.from_documents(documents)
        response = index.as_query_engine().query("What is the answer?")
    """

    _enabled: bool = False
    _callback_handler: Optional["AgentTraceCallbackHandler"] = None

    @classmethod
    def enable(cls) -> None:
        """Enable LlamaIndex auto-instrumentation."""
        if cls._enabled:
            return

        try:
            from llama_index.core.callbacks import CallbackManager
            from llama_index.core import Settings
        except ImportError:
            try:
                from llama_index.callbacks import CallbackManager
                from llama_index import Settings
            except ImportError:
                logger.warning(
                    "llama-index or llama-index-core not installed, skipping instrumentation"
                )
                return

        cls._callback_handler = AgentTraceCallbackHandler()

        # Add to global settings
        try:
            if Settings.callback_manager is None:
                Settings.callback_manager = CallbackManager([cls._callback_handler])
            else:
                Settings.callback_manager.add_handler(cls._callback_handler)
        except Exception as e:
            logger.debug(f"Could not set global callback manager: {e}")

        cls._enabled = True
        logger.debug("LlamaIndex instrumentation enabled")

    @classmethod
    def disable(cls) -> None:
        """Disable LlamaIndex auto-instrumentation."""
        if not cls._enabled:
            return

        try:
            from llama_index.core import Settings

            if Settings.callback_manager and cls._callback_handler:
                handlers = Settings.callback_manager.handlers
                if cls._callback_handler in handlers:
                    handlers.remove(cls._callback_handler)
        except Exception:
            pass

        cls._callback_handler = None
        cls._enabled = False
        logger.debug("LlamaIndex instrumentation disabled")

    @classmethod
    def get_callback_handler(cls) -> Optional["AgentTraceCallbackHandler"]:
        """
        Get the callback handler to pass to LlamaIndex.

        Example:
            from agenttrace.integrations import LlamaIndexInstrumentation

            LlamaIndexInstrumentation.enable()
            handler = LlamaIndexInstrumentation.get_callback_handler()

            # Use with LlamaIndex
            from llama_index.core.callbacks import CallbackManager
            callback_manager = CallbackManager([handler])
        """
        if not cls._enabled:
            cls.enable()
        return cls._callback_handler


class AgentTraceCallbackHandler:
    """
    LlamaIndex callback handler that sends traces to AgentTrace.

    Can be used directly:
        from agenttrace.integrations.llamaindex import AgentTraceCallbackHandler
        from llama_index.core.callbacks import CallbackManager

        handler = AgentTraceCallbackHandler()
        callback_manager = CallbackManager([handler])
    """

    def __init__(self):
        self._event_map: Dict[str, Any] = {}
        self._trace_map: Dict[str, Any] = {}

    def _get_event_id(self, event_id: Any) -> str:
        """Convert event_id to string."""
        if isinstance(event_id, uuid.UUID):
            return str(event_id)
        return str(event_id)

    # Event handlers
    def on_event_start(
        self,
        event_type: str,
        payload: Optional[Dict[str, Any]] = None,
        event_id: str = "",
        parent_id: str = "",
        **kwargs: Any,
    ) -> str:
        """Called when an event starts."""
        client = get_client()
        if client is None or not client.enabled:
            return event_id

        event_id_str = self._get_event_id(event_id) or str(uuid.uuid4())
        payload = payload or {}

        # Handle different event types
        if event_type in ("llm", "LLM"):
            self._start_llm_event(event_id_str, payload, parent_id)
        elif event_type in ("embedding", "EMBEDDING"):
            self._start_embedding_event(event_id_str, payload, parent_id)
        elif event_type in ("query", "QUERY"):
            self._start_query_event(event_id_str, payload, parent_id)
        elif event_type in ("retrieve", "RETRIEVE"):
            self._start_retrieve_event(event_id_str, payload, parent_id)
        elif event_type in ("synthesize", "SYNTHESIZE"):
            self._start_synthesize_event(event_id_str, payload, parent_id)
        elif event_type in ("tree", "TREE"):
            self._start_tree_event(event_id_str, payload, parent_id)
        elif event_type in ("sub_question", "SUB_QUESTION"):
            self._start_sub_question_event(event_id_str, payload, parent_id)
        elif event_type in ("reranking", "RERANKING"):
            self._start_reranking_event(event_id_str, payload, parent_id)
        elif event_type in ("agent_step", "AGENT_STEP"):
            self._start_agent_step_event(event_id_str, payload, parent_id)
        elif event_type in ("function_call", "FUNCTION_CALL"):
            self._start_function_call_event(event_id_str, payload, parent_id)
        else:
            self._start_generic_event(event_id_str, event_type, payload, parent_id)

        return event_id_str

    def on_event_end(
        self,
        event_type: str,
        payload: Optional[Dict[str, Any]] = None,
        event_id: str = "",
        **kwargs: Any,
    ) -> None:
        """Called when an event ends."""
        event_id_str = self._get_event_id(event_id)
        event_data = self._event_map.pop(event_id_str, None)

        if event_data is None:
            return

        payload = payload or {}

        # Handle different event types
        event_obj = event_data.get("span") or event_data.get("generation")
        if event_obj is None:
            return

        if event_type in ("llm", "LLM"):
            self._end_llm_event(event_data, payload)
        elif event_type in ("embedding", "EMBEDDING"):
            self._end_embedding_event(event_data, payload)
        elif event_type in ("retrieve", "RETRIEVE"):
            self._end_retrieve_event(event_data, payload)
        else:
            self._end_generic_event(event_data, payload)

    # LLM Events
    def _start_llm_event(
        self, event_id: str, payload: Dict[str, Any], parent_id: str
    ) -> None:
        """Start tracing an LLM event."""
        # Extract model info
        model = payload.get("model_dict", {}).get("model", "unknown")
        messages = payload.get("messages", [])

        # Format messages
        formatted_messages = []
        for msg in messages:
            if hasattr(msg, "content"):
                formatted_messages.append({
                    "role": getattr(msg, "role", "user"),
                    "content": msg.content,
                })
            elif isinstance(msg, dict):
                formatted_messages.append(msg)
            else:
                formatted_messages.append({"content": str(msg)})

        # Extract model parameters
        model_params = {}
        serialized = payload.get("serialized", {})
        for param in ["temperature", "max_tokens", "top_p"]:
            if param in serialized:
                model_params[param] = serialized[param]

        gen = start_generation(
            name="llamaindex.llm",
            model=model,
            model_parameters=model_params,
            input={"messages": formatted_messages},
            metadata={
                "llamaindex.event_id": event_id,
                "llamaindex.parent_id": parent_id,
            },
        )

        self._event_map[event_id] = {
            "generation": gen,
            "type": "llm",
            "start_time": datetime.utcnow(),
            "model": model,
        }

    def _end_llm_event(self, event_data: Dict[str, Any], payload: Dict[str, Any]) -> None:
        """End tracing an LLM event."""
        gen = event_data["generation"]

        # Extract response
        response = payload.get("response", {})
        message = payload.get("messages", [{}])[-1] if payload.get("messages") else {}

        output = {}
        if hasattr(response, "message"):
            msg = response.message
            output = {
                "role": getattr(msg, "role", "assistant"),
                "content": getattr(msg, "content", str(msg)),
            }
            if hasattr(msg, "additional_kwargs"):
                output["additional_kwargs"] = msg.additional_kwargs
        elif message:
            output = {
                "role": message.get("role", "assistant"),
                "content": message.get("content", str(message)),
            }
        else:
            output = {"raw": str(response)}

        # Extract usage
        usage = None
        raw_response = getattr(response, "raw", None)
        if raw_response and hasattr(raw_response, "usage"):
            u = raw_response.usage
            usage = {
                "input_tokens": getattr(u, "prompt_tokens", 0),
                "output_tokens": getattr(u, "completion_tokens", 0),
                "total_tokens": getattr(u, "total_tokens", 0),
            }

        gen.end(output=output, usage=usage, model=event_data.get("model"))

    # Embedding Events
    def _start_embedding_event(
        self, event_id: str, payload: Dict[str, Any], parent_id: str
    ) -> None:
        """Start tracing an embedding event."""
        model = payload.get("model_dict", {}).get("model_name", "unknown")

        gen = start_generation(
            name="llamaindex.embedding",
            model=model,
            input={"chunks": payload.get("chunks", [])},
            metadata={
                "llamaindex.event_id": event_id,
                "llamaindex.parent_id": parent_id,
            },
        )

        self._event_map[event_id] = {
            "generation": gen,
            "type": "embedding",
            "start_time": datetime.utcnow(),
            "model": model,
        }

    def _end_embedding_event(
        self, event_data: Dict[str, Any], payload: Dict[str, Any]
    ) -> None:
        """End tracing an embedding event."""
        gen = event_data["generation"]

        embeddings = payload.get("embeddings", [])
        output = {
            "embedding_count": len(embeddings),
            "dimensions": len(embeddings[0]) if embeddings else 0,
        }

        gen.end(output=output, model=event_data.get("model"))

    # Query Events
    def _start_query_event(
        self, event_id: str, payload: Dict[str, Any], parent_id: str
    ) -> None:
        """Start tracing a query event."""
        query_str = payload.get("query_str", "")

        span = start_span(
            name="llamaindex.query",
            input={"query": query_str},
            metadata={
                "llamaindex.event_id": event_id,
                "llamaindex.parent_id": parent_id,
            },
        )

        self._event_map[event_id] = {
            "span": span,
            "type": "query",
            "start_time": datetime.utcnow(),
        }

    # Retrieve Events
    def _start_retrieve_event(
        self, event_id: str, payload: Dict[str, Any], parent_id: str
    ) -> None:
        """Start tracing a retrieve event."""
        query_str = payload.get("query_str", "")

        span = start_span(
            name="llamaindex.retrieve",
            input={"query": query_str},
            metadata={
                "llamaindex.event_id": event_id,
                "llamaindex.parent_id": parent_id,
            },
        )

        self._event_map[event_id] = {
            "span": span,
            "type": "retrieve",
            "start_time": datetime.utcnow(),
        }

    def _end_retrieve_event(
        self, event_data: Dict[str, Any], payload: Dict[str, Any]
    ) -> None:
        """End tracing a retrieve event."""
        span = event_data["span"]

        nodes = payload.get("nodes", [])
        formatted_nodes = []
        for node in nodes:
            node_dict = {
                "score": getattr(node, "score", None),
            }
            if hasattr(node, "node"):
                inner_node = node.node
                node_dict["text"] = getattr(inner_node, "text", str(inner_node))[:500]
                node_dict["metadata"] = getattr(inner_node, "metadata", {})
            else:
                node_dict["text"] = str(node)[:500]
            formatted_nodes.append(node_dict)

        span.end(output={"nodes": formatted_nodes, "count": len(nodes)})

    # Synthesize Events
    def _start_synthesize_event(
        self, event_id: str, payload: Dict[str, Any], parent_id: str
    ) -> None:
        """Start tracing a synthesize event."""
        query_str = payload.get("query_str", "")

        span = start_span(
            name="llamaindex.synthesize",
            input={"query": query_str},
            metadata={
                "llamaindex.event_id": event_id,
                "llamaindex.parent_id": parent_id,
            },
        )

        self._event_map[event_id] = {
            "span": span,
            "type": "synthesize",
            "start_time": datetime.utcnow(),
        }

    # Tree Events
    def _start_tree_event(
        self, event_id: str, payload: Dict[str, Any], parent_id: str
    ) -> None:
        """Start tracing a tree traversal event."""
        span = start_span(
            name="llamaindex.tree",
            input=payload,
            metadata={
                "llamaindex.event_id": event_id,
                "llamaindex.parent_id": parent_id,
            },
        )

        self._event_map[event_id] = {
            "span": span,
            "type": "tree",
            "start_time": datetime.utcnow(),
        }

    # Sub Question Events
    def _start_sub_question_event(
        self, event_id: str, payload: Dict[str, Any], parent_id: str
    ) -> None:
        """Start tracing a sub question event."""
        sub_question = payload.get("sub_question", {})
        question = getattr(sub_question, "sub_question", str(sub_question))

        span = start_span(
            name="llamaindex.sub_question",
            input={"question": question},
            metadata={
                "llamaindex.event_id": event_id,
                "llamaindex.parent_id": parent_id,
            },
        )

        self._event_map[event_id] = {
            "span": span,
            "type": "sub_question",
            "start_time": datetime.utcnow(),
        }

    # Reranking Events
    def _start_reranking_event(
        self, event_id: str, payload: Dict[str, Any], parent_id: str
    ) -> None:
        """Start tracing a reranking event."""
        query_str = payload.get("query_str", "")
        nodes = payload.get("nodes", [])

        span = start_span(
            name="llamaindex.reranking",
            input={"query": query_str, "node_count": len(nodes)},
            metadata={
                "llamaindex.event_id": event_id,
                "llamaindex.parent_id": parent_id,
            },
        )

        self._event_map[event_id] = {
            "span": span,
            "type": "reranking",
            "start_time": datetime.utcnow(),
        }

    # Agent Step Events
    def _start_agent_step_event(
        self, event_id: str, payload: Dict[str, Any], parent_id: str
    ) -> None:
        """Start tracing an agent step event."""
        task_id = payload.get("task_id", "")
        step_input = payload.get("input", "")

        span = start_span(
            name="llamaindex.agent_step",
            input={"task_id": task_id, "input": step_input},
            metadata={
                "llamaindex.event_id": event_id,
                "llamaindex.parent_id": parent_id,
            },
        )

        self._event_map[event_id] = {
            "span": span,
            "type": "agent_step",
            "start_time": datetime.utcnow(),
        }

    # Function Call Events
    def _start_function_call_event(
        self, event_id: str, payload: Dict[str, Any], parent_id: str
    ) -> None:
        """Start tracing a function call event."""
        tool = payload.get("tool", {})
        tool_name = getattr(tool, "name", tool.get("name", "unknown")) if isinstance(tool, dict) or hasattr(tool, "name") else str(tool)
        tool_input = payload.get("function_call", "")

        span = start_span(
            name=f"llamaindex.tool.{tool_name}",
            input={"tool": tool_name, "arguments": tool_input},
            metadata={
                "llamaindex.event_id": event_id,
                "llamaindex.parent_id": parent_id,
                "llamaindex.tool_name": tool_name,
            },
        )

        self._event_map[event_id] = {
            "span": span,
            "type": "function_call",
            "start_time": datetime.utcnow(),
        }

    # Generic Events
    def _start_generic_event(
        self, event_id: str, event_type: str, payload: Dict[str, Any], parent_id: str
    ) -> None:
        """Start tracing a generic event."""
        span = start_span(
            name=f"llamaindex.{event_type.lower()}",
            input=payload,
            metadata={
                "llamaindex.event_id": event_id,
                "llamaindex.event_type": event_type,
                "llamaindex.parent_id": parent_id,
            },
        )

        self._event_map[event_id] = {
            "span": span,
            "type": "generic",
            "start_time": datetime.utcnow(),
        }

    def _end_generic_event(
        self, event_data: Dict[str, Any], payload: Dict[str, Any]
    ) -> None:
        """End tracing a generic event."""
        span_or_gen = event_data.get("span") or event_data.get("generation")

        if span_or_gen:
            # Try to extract response
            response = payload.get("response")
            if response:
                output = {"response": str(response)[:1000]}
            else:
                output = payload

            span_or_gen.end(output=output)

    # Trace methods (for compatibility with LlamaIndex's trace context)
    def start_trace(self, trace_id: Optional[str] = None) -> None:
        """Called when a trace starts."""
        client = get_client()
        if client is None or not client.enabled:
            return

        trace_id = trace_id or str(uuid.uuid4())
        self._trace_map[trace_id] = {
            "start_time": datetime.utcnow(),
        }

    def end_trace(
        self,
        trace_id: Optional[str] = None,
        trace_map: Optional[Dict[str, Any]] = None,
    ) -> None:
        """Called when a trace ends."""
        if trace_id:
            self._trace_map.pop(trace_id, None)


# Convenience function to create a callback handler
def get_llamaindex_callback() -> AgentTraceCallbackHandler:
    """
    Get a LlamaIndex callback handler for AgentTrace.

    Example:
        from agenttrace.integrations.llamaindex import get_llamaindex_callback
        from llama_index.core.callbacks import CallbackManager

        handler = get_llamaindex_callback()
        callback_manager = CallbackManager([handler])
    """
    return AgentTraceCallbackHandler()

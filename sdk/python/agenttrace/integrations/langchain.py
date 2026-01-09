"""
LangChain auto-instrumentation for AgentTrace.

Automatically traces LangChain LLM calls, chains, and agents when enabled.
"""

from __future__ import annotations

import logging
import uuid
from typing import Any, Dict, List, Optional, Union
from datetime import datetime

from agenttrace.context import get_client, get_current_trace
from agenttrace.span import start_span
from agenttrace.generation import start_generation

logger = logging.getLogger("agenttrace")


class LangChainInstrumentation:
    """
    Auto-instrumentation for LangChain.

    Example:
        from agenttrace import AgentTrace
        from agenttrace.integrations import LangChainInstrumentation

        client = AgentTrace(api_key="...")
        LangChainInstrumentation.enable()

        # Now all LangChain calls are automatically traced
        from langchain_openai import ChatOpenAI
        llm = ChatOpenAI()
        response = llm.invoke("Hello!")
    """

    _enabled: bool = False
    _callback_handler: Optional["AgentTraceCallbackHandler"] = None

    @classmethod
    def enable(cls) -> None:
        """Enable LangChain auto-instrumentation."""
        if cls._enabled:
            return

        try:
            from langchain_core.callbacks import BaseCallbackHandler
            from langchain_core.callbacks.manager import CallbackManager
        except ImportError:
            try:
                from langchain.callbacks.base import BaseCallbackHandler
                from langchain.callbacks.manager import CallbackManager
            except ImportError:
                logger.warning(
                    "langchain or langchain-core not installed, skipping instrumentation"
                )
                return

        cls._callback_handler = AgentTraceCallbackHandler()
        cls._enabled = True
        logger.debug("LangChain instrumentation enabled")

    @classmethod
    def disable(cls) -> None:
        """Disable LangChain auto-instrumentation."""
        if not cls._enabled:
            return

        cls._callback_handler = None
        cls._enabled = False
        logger.debug("LangChain instrumentation disabled")

    @classmethod
    def get_callback_handler(cls) -> Optional["AgentTraceCallbackHandler"]:
        """
        Get the callback handler to pass to LangChain.

        This handler should be passed to LangChain operations:

        Example:
            from agenttrace.integrations import LangChainInstrumentation

            LangChainInstrumentation.enable()
            handler = LangChainInstrumentation.get_callback_handler()

            # Use with LangChain
            llm = ChatOpenAI(callbacks=[handler])

            # Or pass to invoke
            chain.invoke(input, config={"callbacks": [handler]})
        """
        if not cls._enabled:
            cls.enable()
        return cls._callback_handler


class AgentTraceCallbackHandler:
    """
    LangChain callback handler that sends traces to AgentTrace.

    Can be used directly:
        from agenttrace.integrations.langchain import AgentTraceCallbackHandler

        handler = AgentTraceCallbackHandler()
        llm = ChatOpenAI(callbacks=[handler])
    """

    def __init__(self):
        self._run_map: Dict[str, Any] = {}
        self._llm_runs: Dict[str, Any] = {}
        self._chain_runs: Dict[str, Any] = {}
        self._tool_runs: Dict[str, Any] = {}

    def _get_run_id(self, run_id: Any) -> str:
        """Convert run_id to string."""
        if isinstance(run_id, uuid.UUID):
            return str(run_id)
        return str(run_id)

    # LLM Callbacks
    def on_llm_start(
        self,
        serialized: Dict[str, Any],
        prompts: List[str],
        *,
        run_id: Any,
        parent_run_id: Optional[Any] = None,
        tags: Optional[List[str]] = None,
        metadata: Optional[Dict[str, Any]] = None,
        **kwargs: Any,
    ) -> None:
        """Called when LLM starts running."""
        client = get_client()
        if client is None or not client.enabled:
            return

        run_id_str = self._get_run_id(run_id)

        # Extract model info from serialized
        model_name = serialized.get("name", "unknown")
        model_id = serialized.get("id", ["langchain", "llms", "unknown"])
        if isinstance(model_id, list):
            model_name = model_id[-1] if model_id else model_name

        # Extract model from kwargs if available
        invocation_params = kwargs.get("invocation_params", {})
        model = invocation_params.get("model", invocation_params.get("model_name", model_name))

        gen = start_generation(
            name=f"langchain.llm.{model_name}",
            model=model,
            input={"prompts": prompts},
            metadata={
                "langchain.run_id": run_id_str,
                "langchain.tags": tags or [],
                **(metadata or {}),
            },
        )

        self._llm_runs[run_id_str] = {
            "generation": gen,
            "start_time": datetime.utcnow(),
            "model": model,
        }

    def on_llm_end(
        self,
        response: Any,
        *,
        run_id: Any,
        parent_run_id: Optional[Any] = None,
        **kwargs: Any,
    ) -> None:
        """Called when LLM ends running."""
        run_id_str = self._get_run_id(run_id)
        run_data = self._llm_runs.pop(run_id_str, None)

        if run_data is None:
            return

        gen = run_data["generation"]

        # Extract output
        output = self._extract_llm_output(response)
        usage = self._extract_llm_usage(response)

        gen.end(output=output, usage=usage, model=run_data.get("model"))

    def on_llm_error(
        self,
        error: BaseException,
        *,
        run_id: Any,
        parent_run_id: Optional[Any] = None,
        **kwargs: Any,
    ) -> None:
        """Called when LLM errors."""
        run_id_str = self._get_run_id(run_id)
        run_data = self._llm_runs.pop(run_id_str, None)

        if run_data is None:
            return

        gen = run_data["generation"]
        gen.end(output={"error": str(error)})

    # Chat Model Callbacks
    def on_chat_model_start(
        self,
        serialized: Dict[str, Any],
        messages: List[List[Any]],
        *,
        run_id: Any,
        parent_run_id: Optional[Any] = None,
        tags: Optional[List[str]] = None,
        metadata: Optional[Dict[str, Any]] = None,
        **kwargs: Any,
    ) -> None:
        """Called when chat model starts running."""
        client = get_client()
        if client is None or not client.enabled:
            return

        run_id_str = self._get_run_id(run_id)

        # Extract model info
        model_name = serialized.get("name", "unknown")
        model_id = serialized.get("id", ["langchain", "chat_models", "unknown"])
        if isinstance(model_id, list):
            model_name = model_id[-1] if model_id else model_name

        # Get model from kwargs
        invocation_params = kwargs.get("invocation_params", {})
        model = invocation_params.get("model", invocation_params.get("model_name", model_name))

        # Convert messages to serializable format
        formatted_messages = self._format_messages(messages)

        # Extract model parameters
        model_params = {}
        for param in ["temperature", "max_tokens", "top_p", "frequency_penalty", "presence_penalty"]:
            if param in invocation_params:
                model_params[param] = invocation_params[param]

        gen = start_generation(
            name=f"langchain.chat.{model_name}",
            model=model,
            model_parameters=model_params,
            input={"messages": formatted_messages},
            metadata={
                "langchain.run_id": run_id_str,
                "langchain.tags": tags or [],
                **(metadata or {}),
            },
        )

        self._llm_runs[run_id_str] = {
            "generation": gen,
            "start_time": datetime.utcnow(),
            "model": model,
        }

    # Chain Callbacks
    def on_chain_start(
        self,
        serialized: Dict[str, Any],
        inputs: Dict[str, Any],
        *,
        run_id: Any,
        parent_run_id: Optional[Any] = None,
        tags: Optional[List[str]] = None,
        metadata: Optional[Dict[str, Any]] = None,
        **kwargs: Any,
    ) -> None:
        """Called when chain starts running."""
        client = get_client()
        if client is None or not client.enabled:
            return

        run_id_str = self._get_run_id(run_id)

        # Extract chain info
        chain_name = serialized.get("name", "unknown")
        chain_id = serialized.get("id", ["langchain", "chains", "unknown"])
        if isinstance(chain_id, list):
            chain_name = chain_id[-1] if chain_id else chain_name

        span = start_span(
            name=f"langchain.chain.{chain_name}",
            input=inputs,
            metadata={
                "langchain.run_id": run_id_str,
                "langchain.chain_type": chain_name,
                "langchain.tags": tags or [],
                **(metadata or {}),
            },
        )

        self._chain_runs[run_id_str] = {
            "span": span,
            "start_time": datetime.utcnow(),
        }

    def on_chain_end(
        self,
        outputs: Dict[str, Any],
        *,
        run_id: Any,
        parent_run_id: Optional[Any] = None,
        **kwargs: Any,
    ) -> None:
        """Called when chain ends running."""
        run_id_str = self._get_run_id(run_id)
        run_data = self._chain_runs.pop(run_id_str, None)

        if run_data is None:
            return

        span = run_data["span"]
        span.end(output=outputs)

    def on_chain_error(
        self,
        error: BaseException,
        *,
        run_id: Any,
        parent_run_id: Optional[Any] = None,
        **kwargs: Any,
    ) -> None:
        """Called when chain errors."""
        run_id_str = self._get_run_id(run_id)
        run_data = self._chain_runs.pop(run_id_str, None)

        if run_data is None:
            return

        span = run_data["span"]
        span.end(output={"error": str(error)}, level="ERROR")

    # Tool Callbacks
    def on_tool_start(
        self,
        serialized: Dict[str, Any],
        input_str: str,
        *,
        run_id: Any,
        parent_run_id: Optional[Any] = None,
        tags: Optional[List[str]] = None,
        metadata: Optional[Dict[str, Any]] = None,
        **kwargs: Any,
    ) -> None:
        """Called when tool starts running."""
        client = get_client()
        if client is None or not client.enabled:
            return

        run_id_str = self._get_run_id(run_id)

        tool_name = serialized.get("name", "unknown_tool")

        span = start_span(
            name=f"langchain.tool.{tool_name}",
            input={"input": input_str},
            metadata={
                "langchain.run_id": run_id_str,
                "langchain.tool_name": tool_name,
                "langchain.tags": tags or [],
                **(metadata or {}),
            },
        )

        self._tool_runs[run_id_str] = {
            "span": span,
            "start_time": datetime.utcnow(),
        }

    def on_tool_end(
        self,
        output: str,
        *,
        run_id: Any,
        parent_run_id: Optional[Any] = None,
        **kwargs: Any,
    ) -> None:
        """Called when tool ends running."""
        run_id_str = self._get_run_id(run_id)
        run_data = self._tool_runs.pop(run_id_str, None)

        if run_data is None:
            return

        span = run_data["span"]
        span.end(output={"output": output})

    def on_tool_error(
        self,
        error: BaseException,
        *,
        run_id: Any,
        parent_run_id: Optional[Any] = None,
        **kwargs: Any,
    ) -> None:
        """Called when tool errors."""
        run_id_str = self._get_run_id(run_id)
        run_data = self._tool_runs.pop(run_id_str, None)

        if run_data is None:
            return

        span = run_data["span"]
        span.end(output={"error": str(error)}, level="ERROR")

    # Agent Callbacks
    def on_agent_action(
        self,
        action: Any,
        *,
        run_id: Any,
        parent_run_id: Optional[Any] = None,
        **kwargs: Any,
    ) -> None:
        """Called when agent takes an action."""
        client = get_client()
        if client is None or not client.enabled:
            return

        # Log agent action as metadata update on current span
        run_id_str = self._get_run_id(run_id)

        # Try to find the parent chain/agent run
        if parent_run_id:
            parent_id_str = self._get_run_id(parent_run_id)
            run_data = self._chain_runs.get(parent_id_str)
            if run_data:
                span = run_data["span"]
                # Update span with action info
                action_dict = {
                    "tool": getattr(action, "tool", str(action)),
                    "tool_input": getattr(action, "tool_input", ""),
                    "log": getattr(action, "log", ""),
                }
                span.event(name="agent_action", metadata=action_dict)

    def on_agent_finish(
        self,
        finish: Any,
        *,
        run_id: Any,
        parent_run_id: Optional[Any] = None,
        **kwargs: Any,
    ) -> None:
        """Called when agent finishes."""
        client = get_client()
        if client is None or not client.enabled:
            return

        # Log agent finish
        if parent_run_id:
            parent_id_str = self._get_run_id(parent_run_id)
            run_data = self._chain_runs.get(parent_id_str)
            if run_data:
                span = run_data["span"]
                finish_dict = {
                    "output": getattr(finish, "return_values", str(finish)),
                    "log": getattr(finish, "log", ""),
                }
                span.event(name="agent_finish", metadata=finish_dict)

    # Retriever Callbacks
    def on_retriever_start(
        self,
        serialized: Dict[str, Any],
        query: str,
        *,
        run_id: Any,
        parent_run_id: Optional[Any] = None,
        tags: Optional[List[str]] = None,
        metadata: Optional[Dict[str, Any]] = None,
        **kwargs: Any,
    ) -> None:
        """Called when retriever starts running."""
        client = get_client()
        if client is None or not client.enabled:
            return

        run_id_str = self._get_run_id(run_id)

        retriever_name = serialized.get("name", "unknown_retriever")
        retriever_id = serialized.get("id", ["langchain", "retrievers", "unknown"])
        if isinstance(retriever_id, list):
            retriever_name = retriever_id[-1] if retriever_id else retriever_name

        span = start_span(
            name=f"langchain.retriever.{retriever_name}",
            input={"query": query},
            metadata={
                "langchain.run_id": run_id_str,
                "langchain.retriever_type": retriever_name,
                "langchain.tags": tags or [],
                **(metadata or {}),
            },
        )

        self._run_map[run_id_str] = {
            "span": span,
            "type": "retriever",
            "start_time": datetime.utcnow(),
        }

    def on_retriever_end(
        self,
        documents: List[Any],
        *,
        run_id: Any,
        parent_run_id: Optional[Any] = None,
        **kwargs: Any,
    ) -> None:
        """Called when retriever ends running."""
        run_id_str = self._get_run_id(run_id)
        run_data = self._run_map.pop(run_id_str, None)

        if run_data is None:
            return

        span = run_data["span"]

        # Format documents
        formatted_docs = []
        for doc in documents:
            formatted_docs.append({
                "page_content": getattr(doc, "page_content", str(doc)),
                "metadata": getattr(doc, "metadata", {}),
            })

        span.end(output={"documents": formatted_docs, "count": len(documents)})

    def on_retriever_error(
        self,
        error: BaseException,
        *,
        run_id: Any,
        parent_run_id: Optional[Any] = None,
        **kwargs: Any,
    ) -> None:
        """Called when retriever errors."""
        run_id_str = self._get_run_id(run_id)
        run_data = self._run_map.pop(run_id_str, None)

        if run_data is None:
            return

        span = run_data["span"]
        span.end(output={"error": str(error)}, level="ERROR")

    # Helper methods
    def _format_messages(self, messages: List[List[Any]]) -> List[List[Dict[str, Any]]]:
        """Format messages to serializable format."""
        formatted = []
        for message_list in messages:
            formatted_list = []
            for msg in message_list:
                formatted_msg = {
                    "type": type(msg).__name__,
                    "content": getattr(msg, "content", str(msg)),
                }
                if hasattr(msg, "role"):
                    formatted_msg["role"] = msg.role
                if hasattr(msg, "additional_kwargs"):
                    formatted_msg["additional_kwargs"] = msg.additional_kwargs
                formatted_list.append(formatted_msg)
            formatted.append(formatted_list)
        return formatted

    def _extract_llm_output(self, response: Any) -> Dict[str, Any]:
        """Extract output from LLM response."""
        try:
            generations = getattr(response, "generations", [])
            llm_output = getattr(response, "llm_output", {})

            output = {
                "generations": [],
                "llm_output": llm_output,
            }

            for gen_list in generations:
                gen_outputs = []
                for gen in gen_list:
                    gen_dict = {
                        "text": getattr(gen, "text", ""),
                    }
                    if hasattr(gen, "message"):
                        msg = gen.message
                        gen_dict["message"] = {
                            "type": type(msg).__name__,
                            "content": getattr(msg, "content", ""),
                        }
                        if hasattr(msg, "tool_calls"):
                            gen_dict["message"]["tool_calls"] = [
                                {
                                    "name": tc.get("name", ""),
                                    "args": tc.get("args", {}),
                                    "id": tc.get("id", ""),
                                }
                                for tc in (msg.tool_calls or [])
                            ]
                    if hasattr(gen, "generation_info"):
                        gen_dict["generation_info"] = gen.generation_info
                    gen_outputs.append(gen_dict)
                output["generations"].append(gen_outputs)

            return output
        except Exception as e:
            logger.debug(f"Error extracting LLM output: {e}")
            return {"raw": str(response)}

    def _extract_llm_usage(self, response: Any) -> Optional[Dict[str, int]]:
        """Extract usage from LLM response."""
        try:
            llm_output = getattr(response, "llm_output", {}) or {}

            # Try different usage formats
            if "token_usage" in llm_output:
                usage = llm_output["token_usage"]
                return {
                    "input_tokens": usage.get("prompt_tokens", 0),
                    "output_tokens": usage.get("completion_tokens", 0),
                    "total_tokens": usage.get("total_tokens", 0),
                }

            if "usage" in llm_output:
                usage = llm_output["usage"]
                return {
                    "input_tokens": usage.get("input_tokens", usage.get("prompt_tokens", 0)),
                    "output_tokens": usage.get("output_tokens", usage.get("completion_tokens", 0)),
                    "total_tokens": usage.get("total_tokens", 0),
                }

            return None
        except Exception:
            return None


# Convenience function to create a callback handler
def get_langchain_callback() -> AgentTraceCallbackHandler:
    """
    Get a LangChain callback handler for AgentTrace.

    Example:
        from agenttrace.integrations.langchain import get_langchain_callback

        callback = get_langchain_callback()
        llm = ChatOpenAI(callbacks=[callback])
    """
    return AgentTraceCallbackHandler()

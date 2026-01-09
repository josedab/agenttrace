"""
Integrations for auto-instrumenting LLM providers and frameworks.
"""

from agenttrace.integrations.openai import OpenAIInstrumentation
from agenttrace.integrations.anthropic import AnthropicInstrumentation
from agenttrace.integrations.langchain import (
    LangChainInstrumentation,
    AgentTraceCallbackHandler as LangChainCallbackHandler,
    get_langchain_callback,
)
from agenttrace.integrations.llamaindex import (
    LlamaIndexInstrumentation,
    AgentTraceCallbackHandler as LlamaIndexCallbackHandler,
    get_llamaindex_callback,
)

__all__ = [
    "OpenAIInstrumentation",
    "AnthropicInstrumentation",
    "LangChainInstrumentation",
    "LangChainCallbackHandler",
    "get_langchain_callback",
    "LlamaIndexInstrumentation",
    "LlamaIndexCallbackHandler",
    "get_llamaindex_callback",
]

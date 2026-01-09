"""
AgentTrace Python SDK - Observability for AI Coding Agents

Example usage:
    from agenttrace import AgentTrace, observe, generation

    client = AgentTrace(api_key="your-api-key")

    @observe()
    def my_function():
        # Your code here
        pass

    # Or use the generation context manager
    with generation(name="llm-call", model="gpt-4") as gen:
        response = openai.chat.completions.create(...)
        gen.update(output=response.choices[0].message.content)

    # Enable auto-instrumentation
    from agenttrace.integrations import OpenAIInstrumentation
    OpenAIInstrumentation.enable()
"""

from agenttrace.client import AgentTrace, Trace, Span, Generation
from agenttrace.decorators import observe
from agenttrace.context import (
    get_client,
    get_current_trace,
    get_current_observation,
    set_client,
    set_current_trace,
    set_current_observation,
    TraceContext,
    ObservationContext,
)
from agenttrace.generation import (
    generation,
    ageneration,
    start_generation,
    GenerationContext,
)
from agenttrace.prompt import Prompt, PromptVersion, get_prompt

# Agent-specific features
from agenttrace.checkpoint import (
    CheckpointClient,
    CheckpointType,
    CheckpointInfo,
    checkpoint_scope,
)
from agenttrace.git import (
    GitClient,
    GitLinkType,
    GitLinkInfo,
)
from agenttrace.fileops import (
    FileOperationClient,
    FileOperationType,
    FileOperationInfo,
    file_op_scope,
)
from agenttrace.terminal import (
    TerminalClient,
    TerminalCommandInfo,
    run as terminal_run,
    terminal_scope,
)

__version__ = "0.1.0"

__all__ = [
    # Main client
    "AgentTrace",
    "Trace",
    "Span",
    "Generation",
    # Decorators
    "observe",
    # Context
    "get_client",
    "get_current_trace",
    "get_current_observation",
    "set_client",
    "set_current_trace",
    "set_current_observation",
    "TraceContext",
    "ObservationContext",
    # Generation helpers
    "generation",
    "ageneration",
    "start_generation",
    "GenerationContext",
    # Prompts
    "Prompt",
    "PromptVersion",
    "get_prompt",
    # Checkpoints
    "CheckpointClient",
    "CheckpointType",
    "CheckpointInfo",
    "checkpoint_scope",
    # Git links
    "GitClient",
    "GitLinkType",
    "GitLinkInfo",
    # File operations
    "FileOperationClient",
    "FileOperationType",
    "FileOperationInfo",
    "file_op_scope",
    # Terminal commands
    "TerminalClient",
    "TerminalCommandInfo",
    "terminal_run",
    "terminal_scope",
]

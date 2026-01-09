"""
AutoGen Multi-Agent Chat with AgentTrace Tracing

This example shows how to trace AutoGen agents having a conversation
to solve a coding task.
"""

import os
from typing import Dict, Any

# AgentTrace imports
from agenttrace import AgentTrace
from agenttrace.span import start_span
from agenttrace.generation import start_generation

# AutoGen imports
from autogen import AssistantAgent, UserProxyAgent, config_list_from_json


# Initialize AgentTrace
client = AgentTrace(
    api_key=os.environ.get("AGENTTRACE_API_KEY", "your-api-key"),
    host=os.environ.get("AGENTTRACE_HOST", "http://localhost:8080"),
)


class TracedAssistantAgent(AssistantAgent):
    """AssistantAgent with AgentTrace instrumentation."""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self._current_span = None

    def generate_reply(
        self,
        messages: list = None,
        sender: "Agent" = None,
        **kwargs,
    ) -> str | dict | None:
        """Generate a reply with tracing."""

        # Start a span for this generation
        span = start_span(
            name=f"autogen.{self.name}.generate_reply",
            input={
                "messages": [
                    {"role": m.get("role", "user"), "content": m.get("content", "")[:200]}
                    for m in (messages or [])[-3:]  # Last 3 messages
                ],
                "sender": sender.name if sender else None,
            },
            metadata={
                "agent_name": self.name,
                "agent_type": "assistant",
            },
        )

        try:
            # Call the original method
            reply = super().generate_reply(messages, sender, **kwargs)

            # Record the output
            if isinstance(reply, dict):
                output = {"type": "dict", "content": str(reply)[:500]}
            else:
                output = {"type": "string", "content": str(reply)[:500] if reply else None}

            span.end(output=output)
            return reply

        except Exception as e:
            span.end(output={"error": str(e)}, level="ERROR")
            raise


class TracedUserProxyAgent(UserProxyAgent):
    """UserProxyAgent with AgentTrace instrumentation."""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def generate_reply(
        self,
        messages: list = None,
        sender: "Agent" = None,
        **kwargs,
    ) -> str | dict | None:
        """Generate a reply with tracing."""

        span = start_span(
            name=f"autogen.{self.name}.generate_reply",
            input={
                "messages": [
                    {"role": m.get("role", "user"), "content": m.get("content", "")[:200]}
                    for m in (messages or [])[-3:]
                ],
                "sender": sender.name if sender else None,
            },
            metadata={
                "agent_name": self.name,
                "agent_type": "user_proxy",
            },
        )

        try:
            reply = super().generate_reply(messages, sender, **kwargs)

            if isinstance(reply, dict):
                output = {"type": "dict", "content": str(reply)[:500]}
            else:
                output = {"type": "string", "content": str(reply)[:500] if reply else None}

            span.end(output=output)
            return reply

        except Exception as e:
            span.end(output={"error": str(e)}, level="ERROR")
            raise


def create_agents():
    """Create the AutoGen agents."""

    # Configure LLM
    llm_config = {
        "model": "gpt-4o-mini",
        "api_key": os.environ.get("OPENAI_API_KEY"),
    }

    # Create assistant agent
    assistant = TracedAssistantAgent(
        name="coding_assistant",
        system_message="""You are a helpful AI assistant skilled in Python programming.
        You write clean, efficient, and well-documented code.
        When asked to write code, provide complete, runnable solutions.
        Always explain your code briefly.""",
        llm_config=llm_config,
    )

    # Create user proxy agent
    user_proxy = TracedUserProxyAgent(
        name="user_proxy",
        human_input_mode="NEVER",  # Automatic mode for demo
        max_consecutive_auto_reply=3,
        code_execution_config={
            "work_dir": "workspace",
            "use_docker": False,
        },
        llm_config=llm_config,
    )

    # Create critic agent
    critic = TracedAssistantAgent(
        name="code_critic",
        system_message="""You are a code reviewer. Your job is to review code
        and suggest improvements for:
        1. Code quality and readability
        2. Performance optimizations
        3. Error handling
        4. Best practices

        Be constructive and specific in your feedback.""",
        llm_config=llm_config,
    )

    return user_proxy, assistant, critic


def main():
    """Run the AutoGen example with tracing."""

    print("\nInitializing AutoGen agents...\n")

    user_proxy, assistant, critic = create_agents()

    # The task to solve
    task = """
    Create a Python function that:
    1. Generates a list of 100 random numbers between 1 and 1000
    2. Finds the top 10 numbers
    3. Calculates their mean, median, and standard deviation
    4. Returns a dictionary with the results

    Please write clean, well-documented code.
    """

    print(f"Task: {task}")
    print("=" * 50)

    # Start tracing
    with client.trace(name="autogen-coding-chat") as trace:
        trace.update(
            metadata={
                "task_type": "code_generation",
                "agents": ["user_proxy", "coding_assistant", "code_critic"],
                "model": "gpt-4o-mini",
            }
        )

        # Track the conversation
        with start_span(name="autogen.conversation") as conv_span:
            conv_span.update(input={"task": task})

            # Start the conversation
            chat_result = user_proxy.initiate_chat(
                assistant,
                message=task,
                summary_method="reflection_with_llm",
            )

            # Record conversation summary
            conv_span.end(
                output={
                    "summary": chat_result.summary if hasattr(chat_result, "summary") else str(chat_result),
                    "message_count": len(chat_result.chat_history) if hasattr(chat_result, "chat_history") else 0,
                }
            )

        # Update trace with final result
        trace.update(
            output={
                "summary": chat_result.summary[:500] if hasattr(chat_result, "summary") and chat_result.summary else "No summary",
                "success": True,
            }
        )

    print("\n" + "=" * 50)
    print(f"\nTrace ID: {trace.id}")
    print(f"View trace at: {client.host}/traces/{trace.id}")


if __name__ == "__main__":
    main()

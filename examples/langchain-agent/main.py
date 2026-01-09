"""
LangChain ReAct Agent with AgentTrace Tracing

This example shows how to trace a LangChain agent that uses tools
to answer complex questions.
"""

import os
from datetime import datetime

# AgentTrace imports
from agenttrace import AgentTrace
from agenttrace.integrations import LangChainInstrumentation, get_langchain_callback

# LangChain imports
from langchain_openai import ChatOpenAI
from langchain.agents import AgentExecutor, create_react_agent
from langchain.tools import Tool
from langchain_core.prompts import PromptTemplate


# Initialize AgentTrace
client = AgentTrace(
    api_key=os.environ.get("AGENTTRACE_API_KEY", "your-api-key"),
    host=os.environ.get("AGENTTRACE_HOST", "http://localhost:8080"),
)

# Enable automatic LangChain instrumentation
LangChainInstrumentation.enable()


# Define some example tools
def get_current_time(query: str) -> str:
    """Get the current date and time."""
    return datetime.now().strftime("%Y-%m-%d %H:%M:%S")


def calculate(expression: str) -> str:
    """Evaluate a mathematical expression."""
    try:
        # Simple calculator - in production use a safer evaluation method
        allowed_chars = set("0123456789+-*/(). ")
        if all(c in allowed_chars for c in expression):
            result = eval(expression)
            return str(result)
        return "Error: Invalid expression"
    except Exception as e:
        return f"Error: {str(e)}"


def search_knowledge(query: str) -> str:
    """Search for information (mock implementation)."""
    # In a real app, this would query a vector store or API
    knowledge_base = {
        "python": "Python is a high-level programming language created by Guido van Rossum in 1991.",
        "langchain": "LangChain is a framework for developing applications powered by language models.",
        "agenttrace": "AgentTrace is an observability platform for AI coding agents.",
    }

    query_lower = query.lower()
    for key, value in knowledge_base.items():
        if key in query_lower:
            return value

    return "No specific information found. Please try a different query."


# Create tools
tools = [
    Tool(
        name="CurrentTime",
        func=get_current_time,
        description="Get the current date and time. Use this when you need to know what time or date it is.",
    ),
    Tool(
        name="Calculator",
        func=calculate,
        description="Perform mathematical calculations. Input should be a mathematical expression like '2 + 2' or '(5 * 10) / 2'.",
    ),
    Tool(
        name="KnowledgeSearch",
        func=search_knowledge,
        description="Search for information about programming topics. Use this to find facts about Python, LangChain, AgentTrace, etc.",
    ),
]

# ReAct prompt template
react_prompt = PromptTemplate.from_template("""Answer the following questions as best you can. You have access to the following tools:

{tools}

Use the following format:

Question: the input question you must answer
Thought: you should always think about what to do
Action: the action to take, should be one of [{tool_names}]
Action Input: the input to the action
Observation: the result of the action
... (this Thought/Action/Action Input/Observation can repeat N times)
Thought: I now know the final answer
Final Answer: the final answer to the original input question

Begin!

Question: {input}
Thought:{agent_scratchpad}""")


def main():
    """Run the LangChain agent with tracing."""

    # Get the callback handler for explicit passing
    callback = get_langchain_callback()

    # Create the LLM
    llm = ChatOpenAI(
        model="gpt-4o-mini",
        temperature=0,
        callbacks=[callback],  # Attach callback
    )

    # Create the agent
    agent = create_react_agent(llm, tools, react_prompt)

    # Create the agent executor
    agent_executor = AgentExecutor(
        agent=agent,
        tools=tools,
        verbose=True,
        handle_parsing_errors=True,
        callbacks=[callback],  # Attach callback
    )

    # Start a trace for this interaction
    with client.trace(name="langchain-react-agent") as trace:
        # Add some metadata
        trace.update(
            metadata={
                "model": "gpt-4o-mini",
                "agent_type": "react",
                "tools": [t.name for t in tools],
            }
        )

        # Ask the agent a question that requires using multiple tools
        question = """
        I need help with a few things:
        1. What time is it right now?
        2. What is 25 * 4 + 100?
        3. What is LangChain?
        """

        print(f"\nQuestion: {question}\n")
        print("-" * 50)

        # Run the agent
        result = agent_executor.invoke(
            {"input": question},
            config={"callbacks": [callback]},
        )

        print("-" * 50)
        print(f"\nFinal Answer: {result['output']}\n")

        # Update trace with the result
        trace.update(
            output={"answer": result["output"]},
        )

    print(f"Trace ID: {trace.id}")
    print(f"View trace at: {client.host}/traces/{trace.id}")


if __name__ == "__main__":
    main()

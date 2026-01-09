"""
CrewAI Multi-Agent Team with AgentTrace Tracing

This example shows how to trace a CrewAI team of agents working together
to research and create content.
"""

import os

# AgentTrace imports
from agenttrace import AgentTrace
from agenttrace.decorators import observe

# CrewAI imports
from crewai import Agent, Task, Crew, Process


# Initialize AgentTrace
client = AgentTrace(
    api_key=os.environ.get("AGENTTRACE_API_KEY", "your-api-key"),
    host=os.environ.get("AGENTTRACE_HOST", "http://localhost:8080"),
)


# Define agents
researcher = Agent(
    role="Senior Research Analyst",
    goal="Uncover cutting-edge developments and insights about the given topic",
    backstory="""You work at a leading tech think tank.
    Your expertise lies in identifying emerging trends and technologies.
    You have a knack for dissecting complex data and presenting
    actionable insights.""",
    verbose=True,
    allow_delegation=False,
)

writer = Agent(
    role="Tech Content Strategist",
    goal="Craft compelling content about tech topics based on research",
    backstory="""You are a renowned Content Strategist, known for
    your insightful and engaging articles. You transform complex concepts
    into compelling narratives that educate and inspire readers.""",
    verbose=True,
    allow_delegation=False,
)

editor = Agent(
    role="Senior Editor",
    goal="Review and polish content for publication",
    backstory="""You are an experienced editor with an eye for detail.
    You ensure all content is accurate, engaging, and follows best practices
    for readability and SEO.""",
    verbose=True,
    allow_delegation=False,
)


@observe(name="run_content_crew")
def run_content_crew(topic: str) -> str:
    """Run the content creation crew with tracing."""

    # Define tasks
    research_task = Task(
        description=f"""Conduct comprehensive research about {topic}.

        Your research should cover:
        1. Current state of the technology
        2. Key players and innovations
        3. Recent developments and trends
        4. Future predictions

        Compile your findings into a detailed research brief.""",
        expected_output="A comprehensive research brief with key findings and data points",
        agent=researcher,
    )

    writing_task = Task(
        description=f"""Using the research provided, write an engaging blog post about {topic}.

        Your article should:
        1. Have a compelling headline
        2. Include an engaging introduction
        3. Present the key findings in an accessible way
        4. Include practical insights for readers
        5. End with a thought-provoking conclusion

        Target length: 500-700 words.""",
        expected_output="A well-structured blog post in markdown format",
        agent=writer,
        context=[research_task],
    )

    editing_task = Task(
        description="""Review and edit the blog post for publication.

        Your review should:
        1. Check for factual accuracy
        2. Improve clarity and readability
        3. Fix any grammatical errors
        4. Ensure consistent tone and style
        5. Add any missing context

        Return the polished final version.""",
        expected_output="The final, publication-ready blog post",
        agent=editor,
        context=[writing_task],
    )

    # Create crew
    crew = Crew(
        agents=[researcher, writer, editor],
        tasks=[research_task, writing_task, editing_task],
        process=Process.sequential,
        verbose=True,
    )

    # Execute the crew
    result = crew.kickoff()

    return str(result)


def main():
    """Run the CrewAI example with tracing."""

    topic = "AI Coding Agents and Developer Productivity in 2024"

    print(f"\nTopic: {topic}")
    print("=" * 50)
    print("\nStarting CrewAI content team...\n")

    # Run with a trace wrapper
    with client.trace(name="crewai-content-team") as trace:
        trace.update(
            metadata={
                "topic": topic,
                "agents": ["researcher", "writer", "editor"],
                "process": "sequential",
            }
        )

        result = run_content_crew(topic)

        trace.update(
            output={"content": result[:500] + "..." if len(result) > 500 else result},
        )

    print("\n" + "=" * 50)
    print("FINAL RESULT:")
    print("=" * 50)
    print(result)

    print(f"\nTrace ID: {trace.id}")
    print(f"View trace at: {client.host}/traces/{trace.id}")


if __name__ == "__main__":
    main()

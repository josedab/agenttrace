"""
Terminal command tracking for AgentTrace.

Track terminal/shell commands executed during agent sessions for
full visibility and debugging capabilities.
"""

from __future__ import annotations

import os
import subprocess
import uuid
from contextlib import contextmanager
from dataclasses import dataclass
from datetime import datetime
from typing import Any, Dict, Generator, List, Optional, Tuple, TYPE_CHECKING

if TYPE_CHECKING:
    from agenttrace.client import AgentTrace


@dataclass
class TerminalCommandInfo:
    """Information about a tracked terminal command."""
    id: str
    trace_id: str
    observation_id: Optional[str]
    command: str
    args: List[str]
    working_directory: str
    exit_code: int
    stdout: str
    stderr: str
    success: bool
    duration_ms: int
    started_at: datetime
    completed_at: datetime


class TerminalClient:
    """
    Client for tracking terminal commands.

    Example:
        client = AgentTrace(api_key="...")
        trace = client.trace(name="my-agent")

        # Track a command manually
        trace.terminal_cmd(
            command="npm",
            args=["test"],
            exit_code=0,
            stdout="All tests passed"
        )

        # Or run and track automatically
        result = trace.run_cmd("npm test")

        # Or use subprocess wrapper
        from agenttrace.terminal import run
        result = run(["npm", "test"], trace=trace)
    """

    def __init__(self, client: "AgentTrace"):
        self._client = client

    def track(
        self,
        trace_id: str,
        command: str,
        args: Optional[List[str]] = None,
        observation_id: Optional[str] = None,
        working_directory: Optional[str] = None,
        shell: Optional[str] = None,
        env_vars: Optional[Dict[str, str]] = None,
        exit_code: int = 0,
        stdout: Optional[str] = None,
        stderr: Optional[str] = None,
        stdout_truncated: bool = False,
        stderr_truncated: bool = False,
        timed_out: bool = False,
        killed: bool = False,
        max_memory_bytes: Optional[int] = None,
        cpu_time_ms: Optional[int] = None,
        tool_name: Optional[str] = None,
        reason: Optional[str] = None,
        started_at: Optional[datetime] = None,
        completed_at: Optional[datetime] = None,
        success: Optional[bool] = None,
    ) -> TerminalCommandInfo:
        """
        Track a terminal command.

        Args:
            trace_id: ID of the trace
            command: The command that was executed
            args: Command arguments
            observation_id: Optional observation ID
            working_directory: Working directory
            shell: Shell used (bash, zsh, etc.)
            env_vars: Environment variables (JSON string)
            exit_code: Command exit code
            stdout: Standard output
            stderr: Standard error
            stdout_truncated: Whether stdout was truncated
            stderr_truncated: Whether stderr was truncated
            timed_out: Whether the command timed out
            killed: Whether the command was killed
            max_memory_bytes: Maximum memory used
            cpu_time_ms: CPU time used
            tool_name: Name of the tool that ran the command
            reason: Reason for running the command
            started_at: When the command started
            completed_at: When the command completed
            success: Whether the command succeeded (defaults to exit_code == 0)

        Returns:
            TerminalCommandInfo with details about the tracked command
        """
        cmd_id = str(uuid.uuid4())
        now = datetime.utcnow()

        if started_at is None:
            started_at = now
        if completed_at is None:
            completed_at = now

        duration_ms = int((completed_at - started_at).total_seconds() * 1000)

        if working_directory is None:
            working_directory = os.getcwd()

        if success is None:
            success = exit_code == 0

        args = args or []
        stdout = stdout or ""
        stderr = stderr or ""

        # Convert env_vars dict to JSON string
        env_vars_str = None
        if env_vars:
            import json
            env_vars_str = json.dumps(env_vars)

        # Send to API
        if self._client.enabled:
            self._client._batch_queue.add({
                "type": "terminal-command-create",
                "body": {
                    "id": cmd_id,
                    "traceId": trace_id,
                    "observationId": observation_id,
                    "command": command,
                    "args": args,
                    "workingDirectory": working_directory,
                    "shell": shell,
                    "envVars": env_vars_str,
                    "startedAt": started_at.isoformat() + "Z",
                    "completedAt": completed_at.isoformat() + "Z",
                    "durationMs": duration_ms,
                    "exitCode": exit_code,
                    "stdout": stdout,
                    "stderr": stderr,
                    "stdoutTruncated": stdout_truncated,
                    "stderrTruncated": stderr_truncated,
                    "success": success,
                    "timedOut": timed_out,
                    "killed": killed,
                    "maxMemoryBytes": max_memory_bytes,
                    "cpuTimeMs": cpu_time_ms,
                    "toolName": tool_name,
                    "reason": reason,
                },
            })

        return TerminalCommandInfo(
            id=cmd_id,
            trace_id=trace_id,
            observation_id=observation_id,
            command=command,
            args=args,
            working_directory=working_directory,
            exit_code=exit_code,
            stdout=stdout,
            stderr=stderr,
            success=success,
            duration_ms=duration_ms,
            started_at=started_at,
            completed_at=completed_at,
        )

    def run(
        self,
        trace_id: str,
        command: str,
        args: Optional[List[str]] = None,
        observation_id: Optional[str] = None,
        working_directory: Optional[str] = None,
        env: Optional[Dict[str, str]] = None,
        timeout: Optional[float] = None,
        shell: bool = False,
        tool_name: Optional[str] = None,
        reason: Optional[str] = None,
        max_output_bytes: int = 100000,
    ) -> Tuple[TerminalCommandInfo, subprocess.CompletedProcess]:
        """
        Run a command and track it.

        Args:
            trace_id: ID of the trace
            command: Command to run
            args: Command arguments
            observation_id: Optional observation ID
            working_directory: Working directory
            env: Environment variables
            timeout: Timeout in seconds
            shell: Whether to run in shell
            tool_name: Name of the tool running the command
            reason: Reason for running the command
            max_output_bytes: Maximum output bytes to capture

        Returns:
            Tuple of (TerminalCommandInfo, CompletedProcess)
        """
        started_at = datetime.utcnow()

        if working_directory is None:
            working_directory = os.getcwd()

        # Build command
        if args:
            cmd = [command] + args
        else:
            cmd = command if shell else [command]

        timed_out = False
        killed = False
        stdout = ""
        stderr = ""
        exit_code = 0

        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                cwd=working_directory,
                env=env,
                timeout=timeout,
                shell=shell,
            )
            stdout = result.stdout[:max_output_bytes] if result.stdout else ""
            stderr = result.stderr[:max_output_bytes] if result.stderr else ""
            exit_code = result.returncode
        except subprocess.TimeoutExpired as e:
            timed_out = True
            killed = True
            stdout = e.stdout[:max_output_bytes] if e.stdout else ""
            stderr = e.stderr[:max_output_bytes] if e.stderr else ""
            exit_code = -1
            result = subprocess.CompletedProcess(cmd, exit_code, stdout, stderr)
        except Exception as e:
            stderr = str(e)
            exit_code = -1
            result = subprocess.CompletedProcess(cmd, exit_code, stdout, stderr)

        completed_at = datetime.utcnow()

        stdout_truncated = len(result.stdout or "") > max_output_bytes if hasattr(result, 'stdout') else False
        stderr_truncated = len(result.stderr or "") > max_output_bytes if hasattr(result, 'stderr') else False

        info = self.track(
            trace_id=trace_id,
            command=command,
            args=args,
            observation_id=observation_id,
            working_directory=working_directory,
            exit_code=exit_code,
            stdout=stdout,
            stderr=stderr,
            stdout_truncated=stdout_truncated,
            stderr_truncated=stderr_truncated,
            timed_out=timed_out,
            killed=killed,
            tool_name=tool_name,
            reason=reason,
            started_at=started_at,
            completed_at=completed_at,
        )

        return info, result


def run(
    cmd: List[str] | str,
    client: Optional["AgentTrace"] = None,
    trace_id: Optional[str] = None,
    **kwargs: Any,
) -> Tuple[TerminalCommandInfo, subprocess.CompletedProcess]:
    """
    Run a command and track it with AgentTrace.

    This is a convenience function that uses the global client.

    Args:
        cmd: Command to run (list or string for shell)
        client: Optional AgentTrace client
        trace_id: Optional trace ID
        **kwargs: Additional arguments passed to TerminalClient.run()

    Returns:
        Tuple of (TerminalCommandInfo, CompletedProcess)
    """
    from agenttrace.context import get_client, get_current_trace

    if client is None:
        client = get_client()

    if client is None:
        raise RuntimeError("No AgentTrace client available. Initialize one first.")

    if trace_id is None:
        trace = get_current_trace()
        if trace is None:
            raise RuntimeError("No active trace. Create a trace first.")
        trace_id = trace.id

    term_client = TerminalClient(client)

    if isinstance(cmd, str):
        return term_client.run(trace_id, cmd, shell=True, **kwargs)
    else:
        return term_client.run(trace_id, cmd[0], args=cmd[1:] if len(cmd) > 1 else None, **kwargs)


@contextmanager
def terminal_scope(
    client: "AgentTrace",
    trace_id: str,
    command: str,
    args: Optional[List[str]] = None,
    observation_id: Optional[str] = None,
    tool_name: Optional[str] = None,
    reason: Optional[str] = None,
) -> Generator[Dict[str, Any], None, None]:
    """
    Context manager for tracking terminal commands manually.

    Use this when you need to run commands yourself but want tracking.

    Example:
        with terminal_scope(client, trace_id, "npm", ["test"]) as ctx:
            result = subprocess.run(["npm", "test"], capture_output=True)
            ctx["exit_code"] = result.returncode
            ctx["stdout"] = result.stdout
            ctx["stderr"] = result.stderr
    """
    started_at = datetime.utcnow()
    context: Dict[str, Any] = {
        "exit_code": 0,
        "stdout": "",
        "stderr": "",
        "success": True,
    }

    try:
        yield context
    except Exception as e:
        context["success"] = False
        context["stderr"] = str(e)
        context["exit_code"] = -1
        raise
    finally:
        completed_at = datetime.utcnow()
        term_client = TerminalClient(client)
        term_client.track(
            trace_id=trace_id,
            command=command,
            args=args,
            observation_id=observation_id,
            exit_code=context.get("exit_code", 0),
            stdout=context.get("stdout", ""),
            stderr=context.get("stderr", ""),
            success=context.get("success", True),
            tool_name=tool_name,
            reason=reason,
            started_at=started_at,
            completed_at=completed_at,
        )

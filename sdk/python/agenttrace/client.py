"""
AgentTrace client for tracing and observability.
"""

from __future__ import annotations

import atexit
import uuid
from datetime import datetime
from typing import Any, Dict, List, Optional

from agenttrace.context import set_client, set_current_trace
from agenttrace.transport.http import HttpTransport
from agenttrace.transport.batch import BatchQueue
from agenttrace.checkpoint import CheckpointClient, CheckpointType, CheckpointInfo
from agenttrace.git import GitClient, GitLinkType, GitLinkInfo
from agenttrace.fileops import FileOperationClient, FileOperationType, FileOperationInfo
from agenttrace.terminal import TerminalClient, TerminalCommandInfo


class AgentTrace:
    """
    Main AgentTrace client for tracing and observability.

    Example:
        client = AgentTrace(
            api_key="your-api-key",
            host="https://api.agenttrace.io"
        )

        # Start a trace
        trace = client.trace(name="my-trace")

        # Create a generation
        generation = trace.generation(
            name="llm-call",
            model="gpt-4",
            input={"query": "Hello"},
        )
        generation.end(output="Hi there!")

        # End the trace
        trace.end()

        # Flush remaining events
        client.flush()
    """

    def __init__(
        self,
        api_key: str,
        host: str = "https://api.agenttrace.io",
        public_key: Optional[str] = None,
        project_id: Optional[str] = None,
        enabled: bool = True,
        flush_at: int = 20,
        flush_interval: float = 5.0,
        max_retries: int = 3,
        timeout: float = 10.0,
    ):
        """
        Initialize the AgentTrace client.

        Args:
            api_key: Your AgentTrace API key
            host: AgentTrace API host URL
            public_key: Optional public key for client-side usage
            project_id: Optional project ID override
            enabled: Whether tracing is enabled
            flush_at: Number of events before auto-flush
            flush_interval: Seconds between auto-flush
            max_retries: Number of retries for failed requests
            timeout: Request timeout in seconds
        """
        self.api_key = api_key
        self.host = host.rstrip("/")
        self.public_key = public_key
        self.project_id = project_id
        self.enabled = enabled

        # Initialize transport
        self._transport = HttpTransport(
            host=self.host,
            api_key=self.api_key,
            timeout=timeout,
            max_retries=max_retries,
        )

        # Initialize batch queue
        self._batch_queue = BatchQueue(
            transport=self._transport,
            flush_at=flush_at,
            flush_interval=flush_interval,
        )

        # Register flush on exit
        atexit.register(self.flush)

        # Set as global client
        set_client(self)

    def trace(
        self,
        name: str,
        id: Optional[str] = None,
        user_id: Optional[str] = None,
        session_id: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None,
        tags: Optional[List[str]] = None,
        input: Optional[Any] = None,
        public: bool = False,
    ) -> "Trace":
        """
        Create a new trace.

        Args:
            name: Name of the trace
            id: Optional trace ID (generated if not provided)
            user_id: Optional user identifier
            session_id: Optional session identifier
            metadata: Optional metadata dictionary
            tags: Optional list of tags
            input: Optional input data
            public: Whether the trace is publicly accessible

        Returns:
            Trace object
        """
        trace = Trace(
            client=self,
            id=id or str(uuid.uuid4()),
            name=name,
            user_id=user_id,
            session_id=session_id,
            metadata=metadata or {},
            tags=tags or [],
            input=input,
            public=public,
        )

        # Set as current trace in context
        set_current_trace(trace)

        return trace

    def score(
        self,
        trace_id: str,
        name: str,
        value: float | bool | str,
        observation_id: Optional[str] = None,
        data_type: str = "NUMERIC",
        comment: Optional[str] = None,
    ) -> None:
        """
        Submit a score for a trace or observation.

        Args:
            trace_id: ID of the trace
            name: Score name
            value: Score value (0-1 for numeric, bool, or categorical string)
            observation_id: Optional observation ID
            data_type: Score data type (NUMERIC, BOOLEAN, CATEGORICAL)
            comment: Optional comment
        """
        if not self.enabled:
            return

        self._batch_queue.add({
            "type": "score-create",
            "body": {
                "id": str(uuid.uuid4()),
                "traceId": trace_id,
                "observationId": observation_id,
                "name": name,
                "value": value,
                "dataType": data_type,
                "comment": comment,
                "source": "API",
            },
        })

    def flush(self) -> None:
        """Flush all pending events to the server."""
        self._batch_queue.flush()

    def shutdown(self) -> None:
        """Shutdown the client and flush remaining events."""
        self.flush()
        self._batch_queue.stop()


class Trace:
    """
    Represents a trace in AgentTrace.
    """

    def __init__(
        self,
        client: AgentTrace,
        id: str,
        name: str,
        user_id: Optional[str] = None,
        session_id: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None,
        tags: Optional[List[str]] = None,
        input: Optional[Any] = None,
        public: bool = False,
    ):
        self._client = client
        self.id = id
        self.name = name
        self.user_id = user_id
        self.session_id = session_id
        self.metadata = metadata or {}
        self.tags = tags or []
        self.input = input
        self.public = public
        self.start_time = datetime.utcnow()
        self.end_time: Optional[datetime] = None
        self.output: Optional[Any] = None
        self._ended = False

        # Send trace create event
        self._send_create()

    def _send_create(self) -> None:
        """Send trace creation event."""
        if not self._client.enabled:
            return

        self._client._batch_queue.add({
            "type": "trace-create",
            "body": {
                "id": self.id,
                "name": self.name,
                "userId": self.user_id,
                "sessionId": self.session_id,
                "metadata": self.metadata,
                "tags": self.tags,
                "input": self.input,
                "public": self.public,
                "timestamp": self.start_time.isoformat() + "Z",
            },
        })

    def span(
        self,
        name: str,
        id: Optional[str] = None,
        parent_observation_id: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None,
        input: Optional[Any] = None,
        level: str = "DEFAULT",
    ) -> "Span":
        """
        Create a span within this trace.

        Args:
            name: Name of the span
            id: Optional span ID
            parent_observation_id: Optional parent observation ID
            metadata: Optional metadata
            input: Optional input data
            level: Log level (DEBUG, DEFAULT, WARNING, ERROR)

        Returns:
            Span object
        """
        return Span(
            client=self._client,
            trace_id=self.id,
            id=id or str(uuid.uuid4()),
            name=name,
            parent_observation_id=parent_observation_id,
            metadata=metadata,
            input=input,
            level=level,
        )

    def generation(
        self,
        name: str,
        id: Optional[str] = None,
        parent_observation_id: Optional[str] = None,
        model: Optional[str] = None,
        model_parameters: Optional[Dict[str, Any]] = None,
        input: Optional[Any] = None,
        metadata: Optional[Dict[str, Any]] = None,
        level: str = "DEFAULT",
    ) -> "Generation":
        """
        Create a generation (LLM call) within this trace.

        Args:
            name: Name of the generation
            id: Optional generation ID
            parent_observation_id: Optional parent observation ID
            model: Model name/identifier
            model_parameters: Model parameters (temperature, etc.)
            input: Input prompt or messages
            metadata: Optional metadata
            level: Log level

        Returns:
            Generation object
        """
        return Generation(
            client=self._client,
            trace_id=self.id,
            id=id or str(uuid.uuid4()),
            name=name,
            parent_observation_id=parent_observation_id,
            model=model,
            model_parameters=model_parameters,
            input=input,
            metadata=metadata,
            level=level,
        )

    def update(
        self,
        name: Optional[str] = None,
        user_id: Optional[str] = None,
        session_id: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None,
        tags: Optional[List[str]] = None,
        input: Optional[Any] = None,
        output: Optional[Any] = None,
        public: Optional[bool] = None,
    ) -> "Trace":
        """Update trace properties."""
        if name is not None:
            self.name = name
        if user_id is not None:
            self.user_id = user_id
        if session_id is not None:
            self.session_id = session_id
        if metadata is not None:
            self.metadata.update(metadata)
        if tags is not None:
            self.tags = tags
        if input is not None:
            self.input = input
        if output is not None:
            self.output = output
        if public is not None:
            self.public = public

        # Send update event
        if self._client.enabled:
            self._client._batch_queue.add({
                "type": "trace-update",
                "body": {
                    "id": self.id,
                    "name": self.name,
                    "userId": self.user_id,
                    "sessionId": self.session_id,
                    "metadata": self.metadata,
                    "tags": self.tags,
                    "input": self.input,
                    "output": self.output,
                    "public": self.public,
                },
            })

        return self

    def end(self, output: Optional[Any] = None) -> None:
        """End the trace."""
        if self._ended:
            return

        self._ended = True
        self.end_time = datetime.utcnow()
        if output is not None:
            self.output = output

        # Send final update
        self.update(output=self.output)

        # Clear from context
        set_current_trace(None)

    def score(
        self,
        name: str,
        value: float | bool | str,
        data_type: str = "NUMERIC",
        comment: Optional[str] = None,
    ) -> None:
        """Add a score to this trace."""
        self._client.score(
            trace_id=self.id,
            name=name,
            value=value,
            data_type=data_type,
            comment=comment,
        )

    def checkpoint(
        self,
        name: str,
        checkpoint_type: CheckpointType = CheckpointType.MANUAL,
        observation_id: Optional[str] = None,
        description: Optional[str] = None,
        files: Optional[List[str]] = None,
        include_git_info: bool = True,
    ) -> CheckpointInfo:
        """
        Create a checkpoint for this trace.

        Args:
            name: Human-readable name for the checkpoint
            checkpoint_type: Type of checkpoint (MANUAL, AUTO, TOOL_CALL, etc.)
            observation_id: Optional observation ID to link to
            description: Optional description
            files: List of files to include in the checkpoint
            include_git_info: Whether to auto-detect git info

        Returns:
            CheckpointInfo with details about the created checkpoint
        """
        cp_client = CheckpointClient(self._client)
        return cp_client.create(
            trace_id=self.id,
            name=name,
            checkpoint_type=checkpoint_type,
            observation_id=observation_id,
            description=description,
            files=files,
            include_git_info=include_git_info,
        )

    def git_link(
        self,
        link_type: GitLinkType = GitLinkType.COMMIT,
        observation_id: Optional[str] = None,
        commit_sha: Optional[str] = None,
        branch: Optional[str] = None,
        repo_url: Optional[str] = None,
        commit_message: Optional[str] = None,
        files_changed: Optional[List[str]] = None,
        auto_detect: bool = True,
    ) -> GitLinkInfo:
        """
        Create a git link for this trace.

        Args:
            link_type: Type of git link (START, COMMIT, RESTORE, etc.)
            observation_id: Optional observation ID to link
            commit_sha: Git commit SHA (auto-detected if not provided)
            branch: Git branch name (auto-detected if not provided)
            repo_url: Repository URL (auto-detected if not provided)
            commit_message: Commit message (auto-detected if not provided)
            files_changed: List of changed files
            auto_detect: Whether to auto-detect git info

        Returns:
            GitLinkInfo with details about the created link
        """
        git_client = GitClient(self._client)
        return git_client.link(
            trace_id=self.id,
            observation_id=observation_id,
            link_type=link_type,
            commit_sha=commit_sha,
            branch=branch,
            repo_url=repo_url,
            commit_message=commit_message,
            files_changed=files_changed,
            auto_detect=auto_detect,
        )

    def file_op(
        self,
        operation: FileOperationType,
        file_path: str,
        observation_id: Optional[str] = None,
        new_path: Optional[str] = None,
        content_before: Optional[str] = None,
        content_after: Optional[str] = None,
        lines_added: Optional[int] = None,
        lines_removed: Optional[int] = None,
        diff_preview: Optional[str] = None,
        tool_name: Optional[str] = None,
        reason: Optional[str] = None,
        success: bool = True,
        error_message: Optional[str] = None,
    ) -> FileOperationInfo:
        """
        Track a file operation for this trace.

        Args:
            operation: Type of file operation (CREATE, READ, UPDATE, DELETE, etc.)
            file_path: Path to the file
            observation_id: Optional observation ID
            new_path: New path for rename/move operations
            content_before: File content before operation
            content_after: File content after operation
            lines_added: Number of lines added
            lines_removed: Number of lines removed
            diff_preview: Preview of the diff
            tool_name: Name of the tool that performed the operation
            reason: Reason for the operation
            success: Whether the operation succeeded
            error_message: Error message if failed

        Returns:
            FileOperationInfo with details about the tracked operation
        """
        file_client = FileOperationClient(self._client)
        return file_client.track(
            trace_id=self.id,
            operation=operation,
            file_path=file_path,
            observation_id=observation_id,
            new_path=new_path,
            content_before=content_before,
            content_after=content_after,
            lines_added=lines_added,
            lines_removed=lines_removed,
            diff_preview=diff_preview,
            tool_name=tool_name,
            reason=reason,
            success=success,
            error_message=error_message,
        )

    def terminal_cmd(
        self,
        command: str,
        args: Optional[List[str]] = None,
        observation_id: Optional[str] = None,
        working_directory: Optional[str] = None,
        exit_code: int = 0,
        stdout: Optional[str] = None,
        stderr: Optional[str] = None,
        tool_name: Optional[str] = None,
        reason: Optional[str] = None,
        success: Optional[bool] = None,
    ) -> TerminalCommandInfo:
        """
        Track a terminal command for this trace.

        Args:
            command: The command that was executed
            args: Command arguments
            observation_id: Optional observation ID
            working_directory: Working directory
            exit_code: Command exit code
            stdout: Standard output
            stderr: Standard error
            tool_name: Name of the tool that ran the command
            reason: Reason for running the command
            success: Whether the command succeeded

        Returns:
            TerminalCommandInfo with details about the tracked command
        """
        term_client = TerminalClient(self._client)
        return term_client.track(
            trace_id=self.id,
            command=command,
            args=args,
            observation_id=observation_id,
            working_directory=working_directory,
            exit_code=exit_code,
            stdout=stdout,
            stderr=stderr,
            tool_name=tool_name,
            reason=reason,
            success=success,
        )

    def run_cmd(
        self,
        command: str,
        args: Optional[List[str]] = None,
        observation_id: Optional[str] = None,
        working_directory: Optional[str] = None,
        env: Optional[Dict[str, str]] = None,
        timeout: Optional[float] = None,
        shell: bool = False,
        tool_name: Optional[str] = None,
        reason: Optional[str] = None,
    ):
        """
        Run a command and track it.

        Args:
            command: Command to run
            args: Command arguments
            observation_id: Optional observation ID
            working_directory: Working directory
            env: Environment variables
            timeout: Timeout in seconds
            shell: Whether to run in shell
            tool_name: Name of the tool running the command
            reason: Reason for running the command

        Returns:
            Tuple of (TerminalCommandInfo, CompletedProcess)
        """
        term_client = TerminalClient(self._client)
        return term_client.run(
            trace_id=self.id,
            command=command,
            args=args,
            observation_id=observation_id,
            working_directory=working_directory,
            env=env,
            timeout=timeout,
            shell=shell,
            tool_name=tool_name,
            reason=reason,
        )


class Span:
    """Represents a span within a trace."""

    def __init__(
        self,
        client: AgentTrace,
        trace_id: str,
        id: str,
        name: str,
        parent_observation_id: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None,
        input: Optional[Any] = None,
        level: str = "DEFAULT",
    ):
        self._client = client
        self.trace_id = trace_id
        self.id = id
        self.name = name
        self.parent_observation_id = parent_observation_id
        self.metadata = metadata or {}
        self.input = input
        self.level = level
        self.output: Optional[Any] = None
        self.start_time = datetime.utcnow()
        self.end_time: Optional[datetime] = None
        self._ended = False

        self._send_create()

    def _send_create(self) -> None:
        if not self._client.enabled:
            return

        self._client._batch_queue.add({
            "type": "span-create",
            "body": {
                "id": self.id,
                "traceId": self.trace_id,
                "parentObservationId": self.parent_observation_id,
                "name": self.name,
                "metadata": self.metadata,
                "input": self.input,
                "level": self.level,
                "startTime": self.start_time.isoformat() + "Z",
            },
        })

    def end(self, output: Optional[Any] = None) -> None:
        if self._ended:
            return

        self._ended = True
        self.end_time = datetime.utcnow()
        if output is not None:
            self.output = output

        if self._client.enabled:
            self._client._batch_queue.add({
                "type": "span-update",
                "body": {
                    "id": self.id,
                    "output": self.output,
                    "endTime": self.end_time.isoformat() + "Z",
                },
            })


class Generation:
    """Represents an LLM generation within a trace."""

    def __init__(
        self,
        client: AgentTrace,
        trace_id: str,
        id: str,
        name: str,
        parent_observation_id: Optional[str] = None,
        model: Optional[str] = None,
        model_parameters: Optional[Dict[str, Any]] = None,
        input: Optional[Any] = None,
        metadata: Optional[Dict[str, Any]] = None,
        level: str = "DEFAULT",
    ):
        self._client = client
        self.trace_id = trace_id
        self.id = id
        self.name = name
        self.parent_observation_id = parent_observation_id
        self.model = model
        self.model_parameters = model_parameters or {}
        self.input = input
        self.metadata = metadata or {}
        self.level = level
        self.output: Optional[Any] = None
        self.start_time = datetime.utcnow()
        self.end_time: Optional[datetime] = None
        self.usage: Optional[Dict[str, int]] = None
        self._ended = False

        self._send_create()

    def _send_create(self) -> None:
        if not self._client.enabled:
            return

        self._client._batch_queue.add({
            "type": "generation-create",
            "body": {
                "id": self.id,
                "traceId": self.trace_id,
                "parentObservationId": self.parent_observation_id,
                "name": self.name,
                "model": self.model,
                "modelParameters": self.model_parameters,
                "input": self.input,
                "metadata": self.metadata,
                "level": self.level,
                "startTime": self.start_time.isoformat() + "Z",
            },
        })

    def end(
        self,
        output: Optional[Any] = None,
        usage: Optional[Dict[str, int]] = None,
        model: Optional[str] = None,
    ) -> None:
        if self._ended:
            return

        self._ended = True
        self.end_time = datetime.utcnow()
        if output is not None:
            self.output = output
        if usage is not None:
            self.usage = usage
        if model is not None:
            self.model = model

        if self._client.enabled:
            self._client._batch_queue.add({
                "type": "generation-update",
                "body": {
                    "id": self.id,
                    "output": self.output,
                    "usage": self.usage,
                    "model": self.model,
                    "endTime": self.end_time.isoformat() + "Z",
                },
            })

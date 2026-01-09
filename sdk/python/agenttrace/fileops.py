"""
File operations tracking for AgentTrace.

Track file reads, writes, edits, and other operations performed during
agent execution for full visibility into agent actions.
"""

from __future__ import annotations

import hashlib
import os
import uuid
from contextlib import contextmanager
from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from typing import Any, Dict, Generator, Optional, TYPE_CHECKING

if TYPE_CHECKING:
    from agenttrace.client import AgentTrace


class FileOperationType(Enum):
    """Types of file operations."""
    CREATE = "create"
    READ = "read"
    UPDATE = "update"
    DELETE = "delete"
    RENAME = "rename"
    COPY = "copy"
    MOVE = "move"
    CHMOD = "chmod"


@dataclass
class FileOperationInfo:
    """Information about a tracked file operation."""
    id: str
    trace_id: str
    observation_id: Optional[str]
    operation: FileOperationType
    file_path: str
    new_path: Optional[str]
    file_size: int
    content_hash: Optional[str]
    lines_added: int
    lines_removed: int
    success: bool
    duration_ms: int
    started_at: datetime
    completed_at: datetime


class FileOperationClient:
    """
    Client for tracking file operations.

    Example:
        client = AgentTrace(api_key="...")
        trace = client.trace(name="my-agent")

        # Track a file operation
        trace.file_op(
            operation=FileOperationType.UPDATE,
            file_path="src/main.py",
            lines_added=10,
            lines_removed=5
        )

        # Or use the context manager
        with trace.file_op_scope(FileOperationType.UPDATE, "src/main.py") as op:
            # Perform the operation
            with open("src/main.py", "w") as f:
                f.write(new_content)
    """

    def __init__(self, client: "AgentTrace"):
        self._client = client

    def track(
        self,
        trace_id: str,
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
        started_at: Optional[datetime] = None,
        completed_at: Optional[datetime] = None,
        success: bool = True,
        error_message: Optional[str] = None,
    ) -> FileOperationInfo:
        """
        Track a file operation.

        Args:
            trace_id: ID of the trace
            operation: Type of file operation
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
            started_at: When the operation started
            completed_at: When the operation completed
            success: Whether the operation succeeded
            error_message: Error message if failed

        Returns:
            FileOperationInfo with details about the tracked operation
        """
        op_id = str(uuid.uuid4())
        now = datetime.utcnow()

        if started_at is None:
            started_at = now
        if completed_at is None:
            completed_at = now

        duration_ms = int((completed_at - started_at).total_seconds() * 1000)

        # Calculate file info
        file_size = 0
        file_mode = None
        mime_type = None
        content_hash = None
        content_before_hash = None
        content_after_hash = None

        if os.path.exists(file_path):
            try:
                stat_info = os.stat(file_path)
                file_size = stat_info.st_size
                file_mode = oct(stat_info.st_mode)[-3:]
            except (IOError, OSError):
                pass

        if content_before is not None:
            content_before_hash = hashlib.sha256(content_before.encode()).hexdigest()

        if content_after is not None:
            content_after_hash = hashlib.sha256(content_after.encode()).hexdigest()
            content_hash = content_after_hash

        # Auto-calculate lines changed if content provided
        if lines_added is None and content_before is not None and content_after is not None:
            before_lines = content_before.splitlines()
            after_lines = content_after.splitlines()
            lines_added = len([l for l in after_lines if l not in before_lines])
            lines_removed = len([l for l in before_lines if l not in after_lines])

        lines_added = lines_added or 0
        lines_removed = lines_removed or 0

        # Send to API
        if self._client.enabled:
            self._client._batch_queue.add({
                "type": "file-operation-create",
                "body": {
                    "id": op_id,
                    "traceId": trace_id,
                    "observationId": observation_id,
                    "operation": operation.value,
                    "filePath": file_path,
                    "newPath": new_path,
                    "fileSize": file_size,
                    "fileMode": file_mode,
                    "contentHash": content_hash,
                    "mimeType": mime_type,
                    "linesAdded": lines_added,
                    "linesRemoved": lines_removed,
                    "diffPreview": diff_preview,
                    "contentBeforeHash": content_before_hash,
                    "contentAfterHash": content_after_hash,
                    "toolName": tool_name,
                    "reason": reason,
                    "startedAt": started_at.isoformat() + "Z",
                    "completedAt": completed_at.isoformat() + "Z",
                    "durationMs": duration_ms,
                    "success": success,
                    "errorMessage": error_message,
                },
            })

        return FileOperationInfo(
            id=op_id,
            trace_id=trace_id,
            observation_id=observation_id,
            operation=operation,
            file_path=file_path,
            new_path=new_path,
            file_size=file_size,
            content_hash=content_hash,
            lines_added=lines_added,
            lines_removed=lines_removed,
            success=success,
            duration_ms=duration_ms,
            started_at=started_at,
            completed_at=completed_at,
        )


@contextmanager
def file_op_scope(
    client: "AgentTrace",
    trace_id: str,
    operation: FileOperationType,
    file_path: str,
    observation_id: Optional[str] = None,
    new_path: Optional[str] = None,
    tool_name: Optional[str] = None,
    reason: Optional[str] = None,
) -> Generator[Dict[str, Any], None, None]:
    """
    Context manager for tracking file operations.

    Automatically captures timing and content changes.

    Example:
        with file_op_scope(client, trace_id, FileOperationType.UPDATE, "main.py") as op:
            # Store content before
            with open("main.py") as f:
                op["content_before"] = f.read()

            # Make changes
            with open("main.py", "w") as f:
                f.write(new_content)

            # Store content after
            op["content_after"] = new_content
        # Operation is automatically tracked
    """
    started_at = datetime.utcnow()
    context: Dict[str, Any] = {
        "content_before": None,
        "content_after": None,
        "success": True,
        "error_message": None,
    }

    try:
        yield context
    except Exception as e:
        context["success"] = False
        context["error_message"] = str(e)
        raise
    finally:
        completed_at = datetime.utcnow()
        op_client = FileOperationClient(client)
        op_client.track(
            trace_id=trace_id,
            operation=operation,
            file_path=file_path,
            observation_id=observation_id,
            new_path=new_path,
            content_before=context.get("content_before"),
            content_after=context.get("content_after"),
            tool_name=tool_name,
            reason=reason,
            started_at=started_at,
            completed_at=completed_at,
            success=context.get("success", True),
            error_message=context.get("error_message"),
        )

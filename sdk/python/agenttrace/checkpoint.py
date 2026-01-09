"""
Checkpoint functionality for AgentTrace.

Checkpoints allow you to create snapshots of code state during agent execution,
enabling restoration and debugging of agent sessions.
"""

from __future__ import annotations

import hashlib
import os
import subprocess
import uuid
from contextlib import contextmanager
from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from typing import Any, Dict, Generator, List, Optional, TYPE_CHECKING

if TYPE_CHECKING:
    from agenttrace.client import AgentTrace


class CheckpointType(Enum):
    """Types of checkpoints."""
    MANUAL = "manual"
    AUTO = "auto"
    TOOL_CALL = "tool_call"
    ERROR = "error"
    MILESTONE = "milestone"
    RESTORE = "restore"


@dataclass
class CheckpointInfo:
    """Information about a created checkpoint."""
    id: str
    name: str
    checkpoint_type: CheckpointType
    trace_id: str
    observation_id: Optional[str]
    git_commit_sha: Optional[str]
    git_branch: Optional[str]
    files_changed: List[str]
    total_files: int
    total_size_bytes: int
    created_at: datetime


class CheckpointClient:
    """
    Client for creating and managing checkpoints.

    Example:
        client = AgentTrace(api_key="...")
        trace = client.trace(name="my-agent")

        # Create a checkpoint
        cp = trace.checkpoint(
            name="before-edit",
            checkpoint_type=CheckpointType.MANUAL,
            files=["src/main.py", "src/utils.py"]
        )

        # Or use the context manager
        with trace.checkpoint_scope("edit-files") as cp:
            # Make changes
            pass
    """

    def __init__(self, client: "AgentTrace"):
        self._client = client

    def create(
        self,
        trace_id: str,
        name: str,
        checkpoint_type: CheckpointType = CheckpointType.MANUAL,
        observation_id: Optional[str] = None,
        description: Optional[str] = None,
        files: Optional[List[str]] = None,
        include_git_info: bool = True,
    ) -> CheckpointInfo:
        """
        Create a new checkpoint.

        Args:
            trace_id: ID of the trace this checkpoint belongs to
            name: Human-readable name for the checkpoint
            checkpoint_type: Type of checkpoint
            observation_id: Optional observation ID to link to
            description: Optional description
            files: List of files to include in the checkpoint
            include_git_info: Whether to automatically detect git info

        Returns:
            CheckpointInfo with details about the created checkpoint
        """
        checkpoint_id = str(uuid.uuid4())
        now = datetime.utcnow()

        # Gather git info if requested
        git_commit_sha = None
        git_branch = None
        git_repo_url = None

        if include_git_info:
            git_info = self._get_git_info()
            git_commit_sha = git_info.get("commit_sha")
            git_branch = git_info.get("branch")
            git_repo_url = git_info.get("repo_url")

        # Calculate file info
        files_changed = files or []
        total_files = len(files_changed)
        total_size_bytes = 0

        files_snapshot = {}
        for file_path in files_changed:
            if os.path.exists(file_path):
                try:
                    stat_info = os.stat(file_path)
                    total_size_bytes += stat_info.st_size

                    # Calculate content hash
                    with open(file_path, "rb") as f:
                        content = f.read()
                        content_hash = hashlib.sha256(content).hexdigest()
                        files_snapshot[file_path] = {
                            "size": stat_info.st_size,
                            "hash": content_hash,
                        }
                except (IOError, OSError):
                    pass

        # Send to API
        if self._client.enabled:
            self._client._batch_queue.add({
                "type": "checkpoint-create",
                "body": {
                    "id": checkpoint_id,
                    "traceId": trace_id,
                    "observationId": observation_id,
                    "name": name,
                    "description": description,
                    "type": checkpoint_type.value,
                    "gitCommitSha": git_commit_sha,
                    "gitBranch": git_branch,
                    "gitRepoUrl": git_repo_url,
                    "filesSnapshot": files_snapshot,
                    "filesChanged": files_changed,
                    "totalFiles": total_files,
                    "totalSizeBytes": total_size_bytes,
                    "timestamp": now.isoformat() + "Z",
                },
            })

        return CheckpointInfo(
            id=checkpoint_id,
            name=name,
            checkpoint_type=checkpoint_type,
            trace_id=trace_id,
            observation_id=observation_id,
            git_commit_sha=git_commit_sha,
            git_branch=git_branch,
            files_changed=files_changed,
            total_files=total_files,
            total_size_bytes=total_size_bytes,
            created_at=now,
        )

    def _get_git_info(self) -> Dict[str, Optional[str]]:
        """Get current git repository info."""
        result: Dict[str, Optional[str]] = {
            "commit_sha": None,
            "branch": None,
            "repo_url": None,
        }

        try:
            # Get current commit SHA
            commit_result = subprocess.run(
                ["git", "rev-parse", "HEAD"],
                capture_output=True,
                text=True,
                timeout=5,
            )
            if commit_result.returncode == 0:
                result["commit_sha"] = commit_result.stdout.strip()

            # Get current branch
            branch_result = subprocess.run(
                ["git", "rev-parse", "--abbrev-ref", "HEAD"],
                capture_output=True,
                text=True,
                timeout=5,
            )
            if branch_result.returncode == 0:
                result["branch"] = branch_result.stdout.strip()

            # Get remote URL
            remote_result = subprocess.run(
                ["git", "config", "--get", "remote.origin.url"],
                capture_output=True,
                text=True,
                timeout=5,
            )
            if remote_result.returncode == 0:
                result["repo_url"] = remote_result.stdout.strip()

        except (subprocess.TimeoutExpired, FileNotFoundError):
            pass

        return result


@contextmanager
def checkpoint_scope(
    client: "AgentTrace",
    trace_id: str,
    name: str,
    checkpoint_type: CheckpointType = CheckpointType.MANUAL,
    observation_id: Optional[str] = None,
    description: Optional[str] = None,
    files: Optional[List[str]] = None,
) -> Generator[CheckpointInfo, None, None]:
    """
    Context manager for checkpoint creation.

    Creates a checkpoint at the start and optionally updates it at the end.

    Example:
        with checkpoint_scope(client, trace_id, "edit-session") as cp:
            # Perform edits
            pass
        # Checkpoint is automatically finalized
    """
    cp_client = CheckpointClient(client)
    cp = cp_client.create(
        trace_id=trace_id,
        name=name,
        checkpoint_type=checkpoint_type,
        observation_id=observation_id,
        description=description,
        files=files,
    )
    try:
        yield cp
    finally:
        # Could send an update event here if needed
        pass

"""
Git link functionality for AgentTrace.

Git links allow you to associate traces and observations with git commits,
branches, and repositories for better traceability and debugging.
"""

from __future__ import annotations

import subprocess
import uuid
from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from typing import Dict, List, Optional, TYPE_CHECKING

if TYPE_CHECKING:
    from agenttrace.client import AgentTrace


class GitLinkType(Enum):
    """Types of git links."""
    START = "start"
    COMMIT = "commit"
    RESTORE = "restore"
    BRANCH = "branch"
    DIFF = "diff"


@dataclass
class GitLinkInfo:
    """Information about a created git link."""
    id: str
    trace_id: str
    observation_id: Optional[str]
    link_type: GitLinkType
    commit_sha: str
    branch: Optional[str]
    repo_url: Optional[str]
    commit_message: Optional[str]
    author_name: Optional[str]
    author_email: Optional[str]
    files_changed: List[str]
    created_at: datetime


class GitClient:
    """
    Client for creating and managing git links.

    Example:
        client = AgentTrace(api_key="...")
        trace = client.trace(name="my-agent")

        # Auto-link current git state
        trace.git_link()

        # Or with explicit info
        trace.git_link(
            commit_sha="abc123",
            branch="main",
            link_type=GitLinkType.COMMIT
        )
    """

    def __init__(self, client: "AgentTrace"):
        self._client = client

    def link(
        self,
        trace_id: str,
        observation_id: Optional[str] = None,
        link_type: GitLinkType = GitLinkType.COMMIT,
        commit_sha: Optional[str] = None,
        branch: Optional[str] = None,
        repo_url: Optional[str] = None,
        commit_message: Optional[str] = None,
        files_changed: Optional[List[str]] = None,
        auto_detect: bool = True,
    ) -> GitLinkInfo:
        """
        Create a git link for a trace.

        Args:
            trace_id: ID of the trace to link
            observation_id: Optional observation ID to link
            link_type: Type of git link
            commit_sha: Git commit SHA (auto-detected if not provided)
            branch: Git branch name (auto-detected if not provided)
            repo_url: Repository URL (auto-detected if not provided)
            commit_message: Commit message (auto-detected if not provided)
            files_changed: List of changed files
            auto_detect: Whether to auto-detect git info

        Returns:
            GitLinkInfo with details about the created link
        """
        link_id = str(uuid.uuid4())
        now = datetime.utcnow()

        # Auto-detect git info if requested
        if auto_detect:
            git_info = self._get_git_info()
            if commit_sha is None:
                commit_sha = git_info.get("commit_sha", "")
            if branch is None:
                branch = git_info.get("branch")
            if repo_url is None:
                repo_url = git_info.get("repo_url")
            if commit_message is None:
                commit_message = git_info.get("commit_message")

        author_name = None
        author_email = None
        if auto_detect and commit_sha:
            author_info = self._get_author_info()
            author_name = author_info.get("name")
            author_email = author_info.get("email")

        # Get changed files if not provided
        if files_changed is None and auto_detect:
            files_changed = self._get_changed_files()

        files_changed = files_changed or []

        # Ensure commit_sha is not None
        if commit_sha is None:
            commit_sha = ""

        # Calculate diff stats
        additions = 0
        deletions = 0
        if files_changed:
            diff_stats = self._get_diff_stats()
            additions = diff_stats.get("additions", 0)
            deletions = diff_stats.get("deletions", 0)

        # Send to API
        if self._client.enabled:
            self._client._batch_queue.add({
                "type": "git-link-create",
                "body": {
                    "id": link_id,
                    "traceId": trace_id,
                    "observationId": observation_id,
                    "linkType": link_type.value,
                    "commitSha": commit_sha,
                    "branch": branch,
                    "repoUrl": repo_url,
                    "commitMessage": commit_message,
                    "authorName": author_name,
                    "authorEmail": author_email,
                    "filesChanged": files_changed,
                    "additions": additions,
                    "deletions": deletions,
                    "timestamp": now.isoformat() + "Z",
                },
            })

        return GitLinkInfo(
            id=link_id,
            trace_id=trace_id,
            observation_id=observation_id,
            link_type=link_type,
            commit_sha=commit_sha,
            branch=branch,
            repo_url=repo_url,
            commit_message=commit_message,
            author_name=author_name,
            author_email=author_email,
            files_changed=files_changed,
            created_at=now,
        )

    def _get_git_info(self) -> Dict[str, Optional[str]]:
        """Get current git repository info."""
        result: Dict[str, Optional[str]] = {
            "commit_sha": None,
            "branch": None,
            "repo_url": None,
            "commit_message": None,
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

            # Get commit message
            message_result = subprocess.run(
                ["git", "log", "-1", "--format=%s"],
                capture_output=True,
                text=True,
                timeout=5,
            )
            if message_result.returncode == 0:
                result["commit_message"] = message_result.stdout.strip()

        except (subprocess.TimeoutExpired, FileNotFoundError):
            pass

        return result

    def _get_author_info(self) -> Dict[str, Optional[str]]:
        """Get author info for current commit."""
        result: Dict[str, Optional[str]] = {
            "name": None,
            "email": None,
        }

        try:
            name_result = subprocess.run(
                ["git", "log", "-1", "--format=%an"],
                capture_output=True,
                text=True,
                timeout=5,
            )
            if name_result.returncode == 0:
                result["name"] = name_result.stdout.strip()

            email_result = subprocess.run(
                ["git", "log", "-1", "--format=%ae"],
                capture_output=True,
                text=True,
                timeout=5,
            )
            if email_result.returncode == 0:
                result["email"] = email_result.stdout.strip()

        except (subprocess.TimeoutExpired, FileNotFoundError):
            pass

        return result

    def _get_changed_files(self) -> List[str]:
        """Get list of changed files (staged and unstaged)."""
        files: List[str] = []

        try:
            # Get staged files
            staged_result = subprocess.run(
                ["git", "diff", "--cached", "--name-only"],
                capture_output=True,
                text=True,
                timeout=5,
            )
            if staged_result.returncode == 0:
                files.extend(staged_result.stdout.strip().split("\n"))

            # Get unstaged modified files
            unstaged_result = subprocess.run(
                ["git", "diff", "--name-only"],
                capture_output=True,
                text=True,
                timeout=5,
            )
            if unstaged_result.returncode == 0:
                for f in unstaged_result.stdout.strip().split("\n"):
                    if f and f not in files:
                        files.append(f)

            # Get untracked files
            untracked_result = subprocess.run(
                ["git", "ls-files", "--others", "--exclude-standard"],
                capture_output=True,
                text=True,
                timeout=5,
            )
            if untracked_result.returncode == 0:
                for f in untracked_result.stdout.strip().split("\n"):
                    if f and f not in files:
                        files.append(f)

        except (subprocess.TimeoutExpired, FileNotFoundError):
            pass

        return [f for f in files if f]

    def _get_diff_stats(self) -> Dict[str, int]:
        """Get diff statistics."""
        result = {"additions": 0, "deletions": 0}

        try:
            stat_result = subprocess.run(
                ["git", "diff", "--shortstat"],
                capture_output=True,
                text=True,
                timeout=5,
            )
            if stat_result.returncode == 0:
                output = stat_result.stdout.strip()
                # Parse output like: "2 files changed, 10 insertions(+), 5 deletions(-)"
                if "insertion" in output:
                    parts = output.split(",")
                    for part in parts:
                        if "insertion" in part:
                            result["additions"] = int(part.strip().split()[0])
                        elif "deletion" in part:
                            result["deletions"] = int(part.strip().split()[0])

        except (subprocess.TimeoutExpired, FileNotFoundError, ValueError):
            pass

        return result

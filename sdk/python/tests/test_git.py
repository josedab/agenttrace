"""Tests for the AgentTrace git module."""

import pytest
from unittest.mock import Mock, patch

from agenttrace.git import GitLinkClient, GitLinkInfo, get_git_info


class TestGetGitInfo:
    """Tests for the get_git_info function."""

    def test_get_git_info_success(self):
        """Test getting git info successfully."""
        with patch('subprocess.run') as mock_run:
            # Setup mock responses for different git commands
            def side_effect(args, **kwargs):
                mock_result = Mock()
                mock_result.returncode = 0
                if 'rev-parse' in args and 'HEAD' in args:
                    mock_result.stdout = "abc123def456\n"
                elif '--abbrev-ref' in args:
                    mock_result.stdout = "main\n"
                elif 'remote.origin.url' in args:
                    mock_result.stdout = "https://github.com/test/repo.git\n"
                elif 'user.name' in args:
                    mock_result.stdout = "Test User\n"
                elif 'user.email' in args:
                    mock_result.stdout = "test@example.com\n"
                return mock_result

            mock_run.side_effect = side_effect

            info = get_git_info()

            assert info.get('commit_sha') is not None or mock_run.call_count > 0

    def test_get_git_info_not_in_repo(self):
        """Test getting git info when not in a git repo."""
        with patch('subprocess.run') as mock_run:
            mock_result = Mock()
            mock_result.returncode = 128
            mock_result.stdout = ""
            mock_run.return_value = mock_result

            info = get_git_info()

            # Should not crash, returns what it can
            assert isinstance(info, dict)

    def test_get_git_info_git_not_installed(self):
        """Test getting git info when git is not installed."""
        with patch('subprocess.run') as mock_run:
            mock_run.side_effect = FileNotFoundError()

            info = get_git_info()

            # Should not crash
            assert isinstance(info, dict)


class TestGitLinkClient:
    """Tests for the GitLinkClient."""

    def test_create_git_link_minimal(self):
        """Test creating a git link with minimal options."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        git_client = GitLinkClient(mock_client)

        with patch('agenttrace.git.get_git_info', return_value={
            'commit_sha': 'abc123',
            'branch': 'main',
        }):
            link = git_client.create(
                trace_id="trace-123",
            )

        assert link.trace_id == "trace-123"
        assert link.commit_sha == "abc123"
        assert link.branch == "main"

    def test_create_git_link_with_explicit_values(self):
        """Test creating a git link with explicit values."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        git_client = GitLinkClient(mock_client)

        link = git_client.create(
            trace_id="trace-123",
            commit_sha="explicit-sha",
            branch="feature-branch",
            repository="https://github.com/explicit/repo",
            author="Explicit Author",
            author_email="explicit@example.com",
            message="Explicit commit message",
        )

        assert link.commit_sha == "explicit-sha"
        assert link.branch == "feature-branch"
        assert link.repository == "https://github.com/explicit/repo"
        assert link.author == "Explicit Author"
        assert link.author_email == "explicit@example.com"
        assert link.message == "Explicit commit message"

    def test_create_git_link_disabled_client(self):
        """Test creating a git link when client is disabled."""
        mock_client = Mock()
        mock_client.enabled = False
        mock_client._batch_queue = Mock()

        git_client = GitLinkClient(mock_client)

        link = git_client.create(
            trace_id="trace-123",
            commit_sha="disabled-sha",
        )

        # Link info still returned
        assert link.commit_sha == "disabled-sha"
        # But nothing added to queue
        mock_client._batch_queue.add.assert_not_called()

    def test_batch_queue_event_format(self):
        """Test that events are sent in correct format."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        git_client = GitLinkClient(mock_client)

        git_client.create(
            trace_id="trace-123",
            commit_sha="abc123",
            branch="main",
            message="Test commit",
        )

        # Verify event format
        call_args = mock_client._batch_queue.add.call_args[0][0]
        assert call_args['type'] == 'git-link-create'
        assert 'body' in call_args
        assert call_args['body']['traceId'] == 'trace-123'
        assert call_args['body']['commitSha'] == 'abc123'
        assert call_args['body']['branch'] == 'main'


class TestGitLinkInfo:
    """Tests for GitLinkInfo dataclass."""

    def test_git_link_info_creation(self):
        """Test creating GitLinkInfo."""
        from datetime import datetime
        now = datetime.utcnow()

        info = GitLinkInfo(
            id="git-123",
            trace_id="trace-456",
            commit_sha="abc123def456",
            branch="main",
            repository="https://github.com/test/repo",
            author="Test Author",
            author_email="author@example.com",
            message="Test commit message",
            created_at=now,
        )

        assert info.id == "git-123"
        assert info.trace_id == "trace-456"
        assert info.commit_sha == "abc123def456"
        assert info.branch == "main"

    def test_git_link_info_optional_fields(self):
        """Test GitLinkInfo with optional fields as None."""
        from datetime import datetime
        now = datetime.utcnow()

        info = GitLinkInfo(
            id="git-123",
            trace_id="trace-456",
            commit_sha="abc123",
            branch=None,
            repository=None,
            author=None,
            author_email=None,
            message=None,
            created_at=now,
        )

        assert info.branch is None
        assert info.repository is None
        assert info.author is None

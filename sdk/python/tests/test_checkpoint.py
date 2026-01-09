"""Tests for the AgentTrace checkpoint module."""

import os
import tempfile
import pytest
from unittest.mock import Mock, patch, MagicMock
from datetime import datetime

from agenttrace.checkpoint import (
    CheckpointClient,
    CheckpointInfo,
    CheckpointType,
    checkpoint_scope,
)


class TestCheckpointType:
    """Tests for CheckpointType enum."""

    def test_checkpoint_types(self):
        """Test all checkpoint types exist."""
        assert CheckpointType.MANUAL.value == "manual"
        assert CheckpointType.AUTO.value == "auto"
        assert CheckpointType.TOOL_CALL.value == "tool_call"
        assert CheckpointType.ERROR.value == "error"
        assert CheckpointType.MILESTONE.value == "milestone"
        assert CheckpointType.RESTORE.value == "restore"


class TestCheckpointClient:
    """Tests for the CheckpointClient."""

    def test_create_checkpoint_minimal(self):
        """Test creating a checkpoint with minimal options."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        cp_client = CheckpointClient(mock_client)

        with patch.object(cp_client, '_get_git_info', return_value={}):
            cp = cp_client.create(
                trace_id="trace-123",
                name="test-checkpoint",
            )

        assert cp.id is not None
        assert cp.name == "test-checkpoint"
        assert cp.checkpoint_type == CheckpointType.MANUAL
        assert cp.trace_id == "trace-123"
        assert isinstance(cp.created_at, datetime)

    def test_create_checkpoint_with_all_options(self):
        """Test creating a checkpoint with all options."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        cp_client = CheckpointClient(mock_client)

        with patch.object(cp_client, '_get_git_info', return_value={
            'commit_sha': 'abc123',
            'branch': 'main',
            'repo_url': 'https://github.com/test/repo',
        }):
            cp = cp_client.create(
                trace_id="trace-123",
                name="full-checkpoint",
                checkpoint_type=CheckpointType.MILESTONE,
                observation_id="obs-456",
                description="Important milestone",
                files=[],
                include_git_info=True,
            )

        assert cp.checkpoint_type == CheckpointType.MILESTONE
        assert cp.observation_id == "obs-456"
        assert cp.git_commit_sha == "abc123"
        assert cp.git_branch == "main"

    def test_create_checkpoint_with_files(self):
        """Test creating a checkpoint with file tracking."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        cp_client = CheckpointClient(mock_client)

        # Create temp files
        with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.py') as f1:
            f1.write("print('hello')")
            file1_path = f1.name

        with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.py') as f2:
            f2.write("print('world')")
            file2_path = f2.name

        try:
            with patch.object(cp_client, '_get_git_info', return_value={}):
                cp = cp_client.create(
                    trace_id="trace-123",
                    name="file-checkpoint",
                    files=[file1_path, file2_path],
                    include_git_info=False,
                )

            assert cp.total_files == 2
            assert cp.total_size_bytes > 0
            assert len(cp.files_changed) == 2
        finally:
            os.unlink(file1_path)
            os.unlink(file2_path)

    def test_create_checkpoint_with_nonexistent_file(self):
        """Test creating a checkpoint with nonexistent file doesn't crash."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        cp_client = CheckpointClient(mock_client)

        with patch.object(cp_client, '_get_git_info', return_value={}):
            cp = cp_client.create(
                trace_id="trace-123",
                name="test-checkpoint",
                files=["/nonexistent/file.py"],
                include_git_info=False,
            )

        # Should not crash, but file won't be counted
        assert cp.total_files == 1  # File is in list but not counted in size
        assert cp.total_size_bytes == 0  # File doesn't exist

    def test_create_checkpoint_disabled_client(self):
        """Test creating a checkpoint when client is disabled."""
        mock_client = Mock()
        mock_client.enabled = False
        mock_client._batch_queue = Mock()

        cp_client = CheckpointClient(mock_client)

        with patch.object(cp_client, '_get_git_info', return_value={}):
            cp = cp_client.create(
                trace_id="trace-123",
                name="disabled-checkpoint",
            )

        # Checkpoint info still returned
        assert cp.name == "disabled-checkpoint"
        # But nothing added to queue
        mock_client._batch_queue.add.assert_not_called()

    def test_get_git_info_success(self):
        """Test getting git info successfully."""
        mock_client = Mock()
        cp_client = CheckpointClient(mock_client)

        with patch('subprocess.run') as mock_run:
            mock_run.return_value = Mock(
                returncode=0,
                stdout="abc123def\n"
            )

            git_info = cp_client._get_git_info()

            # Should have called git commands
            assert mock_run.call_count >= 1

    def test_get_git_info_no_git(self):
        """Test getting git info when not in a git repo."""
        mock_client = Mock()
        cp_client = CheckpointClient(mock_client)

        with patch('subprocess.run') as mock_run:
            mock_run.side_effect = FileNotFoundError()

            git_info = cp_client._get_git_info()

            # Should return empty dict without crashing
            assert git_info.get('commit_sha') is None
            assert git_info.get('branch') is None

    def test_batch_queue_event_format(self):
        """Test that events are sent in correct format."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        cp_client = CheckpointClient(mock_client)

        with patch.object(cp_client, '_get_git_info', return_value={
            'commit_sha': 'abc123',
            'branch': 'main',
        }):
            cp_client.create(
                trace_id="trace-123",
                name="test-checkpoint",
                checkpoint_type=CheckpointType.AUTO,
            )

        # Verify event format
        call_args = mock_client._batch_queue.add.call_args[0][0]
        assert call_args['type'] == 'checkpoint-create'
        assert 'body' in call_args
        assert call_args['body']['traceId'] == 'trace-123'
        assert call_args['body']['name'] == 'test-checkpoint'
        assert call_args['body']['type'] == 'auto'


class TestCheckpointScope:
    """Tests for checkpoint_scope context manager."""

    def test_checkpoint_scope_basic(self):
        """Test basic checkpoint scope usage."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        with patch('agenttrace.checkpoint.CheckpointClient._get_git_info', return_value={}):
            with checkpoint_scope(
                mock_client,
                trace_id="trace-123",
                name="scope-test"
            ) as cp:
                assert cp.name == "scope-test"
                assert cp.trace_id == "trace-123"

    def test_checkpoint_scope_with_exception(self):
        """Test checkpoint scope handles exceptions gracefully."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        with patch('agenttrace.checkpoint.CheckpointClient._get_git_info', return_value={}):
            with pytest.raises(ValueError):
                with checkpoint_scope(
                    mock_client,
                    trace_id="trace-123",
                    name="error-scope"
                ) as cp:
                    raise ValueError("Test error")

    def test_checkpoint_scope_with_all_options(self):
        """Test checkpoint scope with all options."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        with patch('agenttrace.checkpoint.CheckpointClient._get_git_info', return_value={}):
            with checkpoint_scope(
                mock_client,
                trace_id="trace-123",
                name="full-scope",
                checkpoint_type=CheckpointType.TOOL_CALL,
                observation_id="obs-789",
                description="Full scope test",
                files=[]
            ) as cp:
                assert cp.checkpoint_type == CheckpointType.TOOL_CALL
                assert cp.observation_id == "obs-789"


class TestCheckpointInfo:
    """Tests for CheckpointInfo dataclass."""

    def test_checkpoint_info_creation(self):
        """Test creating CheckpointInfo."""
        now = datetime.utcnow()
        cp_info = CheckpointInfo(
            id="cp-123",
            name="test",
            checkpoint_type=CheckpointType.MANUAL,
            trace_id="trace-456",
            observation_id="obs-789",
            git_commit_sha="abc123",
            git_branch="main",
            files_changed=["file1.py", "file2.py"],
            total_files=2,
            total_size_bytes=1024,
            created_at=now,
        )

        assert cp_info.id == "cp-123"
        assert cp_info.name == "test"
        assert cp_info.checkpoint_type == CheckpointType.MANUAL
        assert cp_info.trace_id == "trace-456"
        assert cp_info.total_files == 2
        assert cp_info.total_size_bytes == 1024

    def test_checkpoint_info_optional_fields(self):
        """Test CheckpointInfo with optional fields."""
        now = datetime.utcnow()
        cp_info = CheckpointInfo(
            id="cp-123",
            name="minimal",
            checkpoint_type=CheckpointType.AUTO,
            trace_id="trace-456",
            observation_id=None,
            git_commit_sha=None,
            git_branch=None,
            files_changed=[],
            total_files=0,
            total_size_bytes=0,
            created_at=now,
        )

        assert cp_info.observation_id is None
        assert cp_info.git_commit_sha is None
        assert cp_info.git_branch is None

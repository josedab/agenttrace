"""Tests for the AgentTrace file operations module."""

import os
import tempfile
import pytest
from unittest.mock import Mock, patch
from datetime import datetime

from agenttrace.fileops import (
    FileOperationClient,
    FileOperationInfo,
    FileOperationType,
    track_file_operation,
)


class TestFileOperationType:
    """Tests for FileOperationType enum."""

    def test_file_operation_types(self):
        """Test all file operation types exist."""
        assert FileOperationType.READ.value == "read"
        assert FileOperationType.WRITE.value == "write"
        assert FileOperationType.CREATE.value == "create"
        assert FileOperationType.DELETE.value == "delete"
        assert FileOperationType.RENAME.value == "rename"
        assert FileOperationType.MODIFY.value == "modify"


class TestFileOperationClient:
    """Tests for the FileOperationClient."""

    def test_track_read_operation(self):
        """Test tracking a read operation."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        fo_client = FileOperationClient(mock_client)

        with tempfile.NamedTemporaryFile(mode='w', delete=False) as f:
            f.write("test content")
            file_path = f.name

        try:
            op = fo_client.track(
                trace_id="trace-123",
                operation=FileOperationType.READ,
                path=file_path,
            )

            assert op.operation == FileOperationType.READ
            assert op.path == file_path
            assert op.trace_id == "trace-123"
        finally:
            os.unlink(file_path)

    def test_track_write_operation(self):
        """Test tracking a write operation."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        fo_client = FileOperationClient(mock_client)

        op = fo_client.track(
            trace_id="trace-123",
            operation=FileOperationType.WRITE,
            path="/path/to/file.py",
            lines_changed=50,
            content_preview="def hello(): pass",
        )

        assert op.operation == FileOperationType.WRITE
        assert op.lines_changed == 50
        assert op.content_preview == "def hello(): pass"

    def test_track_rename_operation(self):
        """Test tracking a rename operation."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        fo_client = FileOperationClient(mock_client)

        op = fo_client.track(
            trace_id="trace-123",
            operation=FileOperationType.RENAME,
            path="/new/path/file.py",
            old_path="/old/path/file.py",
        )

        assert op.operation == FileOperationType.RENAME
        assert op.path == "/new/path/file.py"
        assert op.old_path == "/old/path/file.py"

    def test_track_delete_operation(self):
        """Test tracking a delete operation."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        fo_client = FileOperationClient(mock_client)

        op = fo_client.track(
            trace_id="trace-123",
            operation=FileOperationType.DELETE,
            path="/deleted/file.py",
        )

        assert op.operation == FileOperationType.DELETE
        assert op.path == "/deleted/file.py"

    def test_track_with_observation_id(self):
        """Test tracking with observation ID."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        fo_client = FileOperationClient(mock_client)

        op = fo_client.track(
            trace_id="trace-123",
            observation_id="obs-456",
            operation=FileOperationType.MODIFY,
            path="/path/file.py",
        )

        assert op.observation_id == "obs-456"

    def test_track_disabled_client(self):
        """Test tracking when client is disabled."""
        mock_client = Mock()
        mock_client.enabled = False
        mock_client._batch_queue = Mock()

        fo_client = FileOperationClient(mock_client)

        op = fo_client.track(
            trace_id="trace-123",
            operation=FileOperationType.READ,
            path="/path/file.py",
        )

        # Operation info still returned
        assert op.operation == FileOperationType.READ
        # But nothing added to queue
        mock_client._batch_queue.add.assert_not_called()

    def test_batch_queue_event_format(self):
        """Test that events are sent in correct format."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        fo_client = FileOperationClient(mock_client)

        fo_client.track(
            trace_id="trace-123",
            observation_id="obs-456",
            operation=FileOperationType.WRITE,
            path="/path/file.py",
            lines_changed=100,
        )

        # Verify event format
        call_args = mock_client._batch_queue.add.call_args[0][0]
        assert call_args['type'] == 'file-operation-create'
        assert 'body' in call_args
        assert call_args['body']['traceId'] == 'trace-123'
        assert call_args['body']['observationId'] == 'obs-456'
        assert call_args['body']['operation'] == 'write'
        assert call_args['body']['path'] == '/path/file.py'
        assert call_args['body']['linesChanged'] == 100


class TestTrackFileOperation:
    """Tests for the track_file_operation convenience function."""

    def test_track_file_operation_function(self):
        """Test the convenience function."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        op = track_file_operation(
            client=mock_client,
            trace_id="trace-123",
            operation=FileOperationType.CREATE,
            path="/new/file.py",
        )

        assert op.operation == FileOperationType.CREATE
        assert op.path == "/new/file.py"


class TestFileOperationInfo:
    """Tests for FileOperationInfo dataclass."""

    def test_file_operation_info_creation(self):
        """Test creating FileOperationInfo."""
        now = datetime.utcnow()
        info = FileOperationInfo(
            id="op-123",
            trace_id="trace-456",
            observation_id="obs-789",
            operation=FileOperationType.WRITE,
            path="/path/to/file.py",
            old_path=None,
            size_bytes=1024,
            lines_changed=50,
            content_preview="def test(): pass",
            created_at=now,
        )

        assert info.id == "op-123"
        assert info.operation == FileOperationType.WRITE
        assert info.size_bytes == 1024
        assert info.lines_changed == 50

    def test_file_operation_info_optional_fields(self):
        """Test FileOperationInfo with optional fields."""
        now = datetime.utcnow()
        info = FileOperationInfo(
            id="op-123",
            trace_id="trace-456",
            observation_id=None,
            operation=FileOperationType.READ,
            path="/path/file.py",
            old_path=None,
            size_bytes=None,
            lines_changed=None,
            content_preview=None,
            created_at=now,
        )

        assert info.observation_id is None
        assert info.size_bytes is None
        assert info.lines_changed is None

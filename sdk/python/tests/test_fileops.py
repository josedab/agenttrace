"""Tests for the AgentTrace file operations module."""

import os
import tempfile
from datetime import datetime
from unittest.mock import Mock

from agenttrace.fileops import (
    FileOperationClient,
    FileOperationInfo,
    FileOperationType,
    file_op_scope,
)


class TestFileOperationType:
    """Tests for FileOperationType enum."""

    def test_file_operation_types(self):
        """Test all file operation types exist."""
        assert FileOperationType.READ.value == "read"
        assert FileOperationType.UPDATE.value == "update"
        assert FileOperationType.CREATE.value == "create"
        assert FileOperationType.DELETE.value == "delete"
        assert FileOperationType.RENAME.value == "rename"
        assert FileOperationType.COPY.value == "copy"
        assert FileOperationType.MOVE.value == "move"
        assert FileOperationType.CHMOD.value == "chmod"


class TestFileOperationClient:
    """Tests for the FileOperationClient."""

    def test_track_read_operation(self):
        """Test tracking a read operation."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        fo_client = FileOperationClient(mock_client)

        with tempfile.NamedTemporaryFile(mode="w", delete=False) as f:
            f.write("test content")
            file_path = f.name

        try:
            op = fo_client.track(
                trace_id="trace-123",
                operation=FileOperationType.READ,
                file_path=file_path,
            )

            assert op.operation == FileOperationType.READ
            assert op.file_path == file_path
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
            operation=FileOperationType.UPDATE,
            file_path="/path/to/file.py",
            lines_added=50,
            content_after="def hello(): pass",
        )

        assert op.operation == FileOperationType.UPDATE
        assert op.lines_added == 50
        assert op.content_hash is not None

    def test_track_rename_operation(self):
        """Test tracking a rename operation."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        fo_client = FileOperationClient(mock_client)

        op = fo_client.track(
            trace_id="trace-123",
            operation=FileOperationType.RENAME,
            file_path="/old/path/file.py",
            new_path="/new/path/file.py",
        )

        assert op.operation == FileOperationType.RENAME
        assert op.file_path == "/old/path/file.py"
        assert op.new_path == "/new/path/file.py"

    def test_track_delete_operation(self):
        """Test tracking a delete operation."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        fo_client = FileOperationClient(mock_client)

        op = fo_client.track(
            trace_id="trace-123",
            operation=FileOperationType.DELETE,
            file_path="/deleted/file.py",
        )

        assert op.operation == FileOperationType.DELETE
        assert op.file_path == "/deleted/file.py"

    def test_track_with_observation_id(self):
        """Test tracking with observation ID."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        fo_client = FileOperationClient(mock_client)

        op = fo_client.track(
            trace_id="trace-123",
            observation_id="obs-456",
            operation=FileOperationType.UPDATE,
            file_path="/path/file.py",
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
            file_path="/path/file.py",
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
            operation=FileOperationType.UPDATE,
            file_path="/path/file.py",
            lines_added=100,
        )

        # Verify event format
        call_args = mock_client._batch_queue.add.call_args[0][0]
        assert call_args["type"] == "file-operation-create"
        assert "body" in call_args
        assert call_args["body"]["traceId"] == "trace-123"
        assert call_args["body"]["observationId"] == "obs-456"
        assert call_args["body"]["operation"] == "update"
        assert call_args["body"]["filePath"] == "/path/file.py"
        assert call_args["body"]["linesAdded"] == 100


class TestFileOperationScope:
    """Tests for the file operation context manager."""

    def test_file_operation_scope(self):
        """Test the convenience context manager."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        with file_op_scope(
            client=mock_client,
            trace_id="trace-123",
            operation=FileOperationType.CREATE,
            file_path="/new/file.py",
        ) as context:
            context["content_after"] = "print('hello')"

        event = mock_client._batch_queue.add.call_args[0][0]
        assert event["body"]["operation"] == "create"
        assert event["body"]["filePath"] == "/new/file.py"


class TestFileOperationInfo:
    """Tests for FileOperationInfo dataclass."""

    def test_file_operation_info_creation(self):
        """Test creating FileOperationInfo."""
        now = datetime.utcnow()
        info = FileOperationInfo(
            id="op-123",
            trace_id="trace-456",
            observation_id="obs-789",
            operation=FileOperationType.UPDATE,
            file_path="/path/to/file.py",
            new_path=None,
            file_size=1024,
            content_hash="abc123",
            lines_added=50,
            lines_removed=2,
            success=True,
            duration_ms=10,
            started_at=now,
            completed_at=now,
        )

        assert info.id == "op-123"
        assert info.operation == FileOperationType.UPDATE
        assert info.file_size == 1024
        assert info.lines_added == 50

    def test_file_operation_info_optional_fields(self):
        """Test FileOperationInfo with optional fields."""
        now = datetime.utcnow()
        info = FileOperationInfo(
            id="op-123",
            trace_id="trace-456",
            observation_id=None,
            operation=FileOperationType.READ,
            file_path="/path/file.py",
            new_path=None,
            file_size=0,
            content_hash=None,
            lines_added=0,
            lines_removed=0,
            success=True,
            duration_ms=0,
            started_at=now,
            completed_at=now,
        )

        assert info.observation_id is None
        assert info.new_path is None
        assert info.content_hash is None

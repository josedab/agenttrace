"""Tests for the AgentTrace terminal module."""

import os
import subprocess
import pytest
from unittest.mock import Mock, patch, MagicMock
from datetime import datetime

from agenttrace.terminal import (
    TerminalClient,
    TerminalCommandInfo,
    run,
    terminal_scope,
)


class TestTerminalClient:
    """Tests for the TerminalClient."""

    def test_track_command_minimal(self):
        """Test tracking a command with minimal options."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        term_client = TerminalClient(mock_client)

        cmd = term_client.track(
            trace_id="trace-123",
            command="npm",
        )

        assert cmd.command == "npm"
        assert cmd.trace_id == "trace-123"
        assert cmd.id is not None
        assert cmd.args == []
        assert cmd.working_directory is not None
        assert cmd.exit_code == 0
        assert cmd.success is True

    def test_track_command_with_args(self):
        """Test tracking a command with arguments."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        term_client = TerminalClient(mock_client)

        cmd = term_client.track(
            trace_id="trace-123",
            command="npm",
            args=["install", "--save", "lodash"],
        )

        assert cmd.command == "npm"
        assert cmd.args == ["install", "--save", "lodash"]

    def test_track_command_with_output(self):
        """Test tracking a command with stdout/stderr."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        term_client = TerminalClient(mock_client)

        cmd = term_client.track(
            trace_id="trace-123",
            command="echo",
            args=["hello"],
            stdout="hello\n",
            stderr="",
            exit_code=0,
        )

        assert cmd.stdout == "hello\n"
        assert cmd.stderr == ""
        assert cmd.exit_code == 0
        assert cmd.success is True

    def test_track_failed_command(self):
        """Test tracking a failed command."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        term_client = TerminalClient(mock_client)

        cmd = term_client.track(
            trace_id="trace-123",
            command="false",
            exit_code=1,
            stderr="Error occurred",
        )

        assert cmd.exit_code == 1
        assert cmd.success is False
        assert cmd.stderr == "Error occurred"

    def test_track_with_custom_working_directory(self):
        """Test tracking with custom working directory."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        term_client = TerminalClient(mock_client)

        cmd = term_client.track(
            trace_id="trace-123",
            command="ls",
            working_directory="/custom/path",
        )

        assert cmd.working_directory == "/custom/path"

    def test_track_with_observation_id(self):
        """Test tracking with observation ID."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        term_client = TerminalClient(mock_client)

        cmd = term_client.track(
            trace_id="trace-123",
            command="test",
            observation_id="obs-456",
        )

        assert cmd.observation_id == "obs-456"

    def test_track_calculates_duration(self):
        """Test that duration is calculated correctly."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        term_client = TerminalClient(mock_client)

        started_at = datetime(2024, 1, 1, 10, 0, 0)
        completed_at = datetime(2024, 1, 1, 10, 0, 5)

        cmd = term_client.track(
            trace_id="trace-123",
            command="long-command",
            started_at=started_at,
            completed_at=completed_at,
        )

        assert cmd.duration_ms == 5000

    def test_track_sends_event_when_enabled(self):
        """Test that events are sent when client is enabled."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        term_client = TerminalClient(mock_client)

        term_client.track(
            trace_id="trace-123",
            command="npm",
            args=["test"],
            exit_code=0,
            stdout="All tests passed",
        )

        mock_client._batch_queue.add.assert_called_once()
        call_args = mock_client._batch_queue.add.call_args[0][0]
        assert call_args['type'] == 'terminal-command-create'
        assert call_args['body']['traceId'] == 'trace-123'
        assert call_args['body']['command'] == 'npm'
        assert call_args['body']['args'] == ['test']

    def test_track_does_not_send_event_when_disabled(self):
        """Test that events are not sent when client is disabled."""
        mock_client = Mock()
        mock_client.enabled = False
        mock_client._batch_queue = Mock()

        term_client = TerminalClient(mock_client)

        cmd = term_client.track(
            trace_id="trace-123",
            command="npm",
            args=["test"],
        )

        mock_client._batch_queue.add.assert_not_called()
        assert cmd.command == "npm"  # Still returns info

    def test_track_timed_out_command(self):
        """Test tracking a timed out command."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        term_client = TerminalClient(mock_client)

        cmd = term_client.track(
            trace_id="trace-123",
            command="slow-command",
            timed_out=True,
            killed=True,
            exit_code=-1,
        )

        call_args = mock_client._batch_queue.add.call_args[0][0]
        assert call_args['body']['timedOut'] is True
        assert call_args['body']['killed'] is True

    def test_track_truncated_output(self):
        """Test tracking with truncated output flags."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        term_client = TerminalClient(mock_client)

        cmd = term_client.track(
            trace_id="trace-123",
            command="verbose-command",
            stdout="truncated...",
            stdout_truncated=True,
            stderr="errors...",
            stderr_truncated=True,
        )

        call_args = mock_client._batch_queue.add.call_args[0][0]
        assert call_args['body']['stdoutTruncated'] is True
        assert call_args['body']['stderrTruncated'] is True


class TestTerminalClientRun:
    """Tests for TerminalClient.run method."""

    def test_run_echo_command(self):
        """Test running echo command."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        term_client = TerminalClient(mock_client)

        cmd, result = term_client.run(
            trace_id="trace-123",
            command="echo",
            args=["hello"],
        )

        assert cmd.command == "echo"
        assert cmd.exit_code == 0
        assert "hello" in cmd.stdout
        assert result.returncode == 0

    def test_run_handles_failed_command(self):
        """Test running a command that fails."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        term_client = TerminalClient(mock_client)

        cmd, result = term_client.run(
            trace_id="trace-123",
            command="false",
        )

        assert cmd.exit_code != 0
        assert cmd.success is False

    def test_run_handles_nonexistent_command(self):
        """Test running a nonexistent command."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        term_client = TerminalClient(mock_client)

        cmd, result = term_client.run(
            trace_id="trace-123",
            command="nonexistent-command-12345",
        )

        assert cmd.exit_code == -1
        assert cmd.success is False

    def test_run_with_custom_working_directory(self):
        """Test running in custom working directory."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        term_client = TerminalClient(mock_client)

        cmd, result = term_client.run(
            trace_id="trace-123",
            command="pwd",
            working_directory="/tmp",
        )

        assert cmd.working_directory == "/tmp"
        assert cmd.exit_code == 0

    def test_run_with_timeout(self):
        """Test running with timeout."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        term_client = TerminalClient(mock_client)

        cmd, result = term_client.run(
            trace_id="trace-123",
            command="sleep",
            args=["10"],
            timeout=0.1,  # 100ms timeout
        )

        assert cmd.exit_code == -1
        # Command should have timed out

    def test_run_with_env_variables(self):
        """Test running with environment variables."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        term_client = TerminalClient(mock_client)

        env = os.environ.copy()
        env["TEST_VAR"] = "test_value"

        cmd, result = term_client.run(
            trace_id="trace-123",
            command="sh",
            args=["-c", "echo $TEST_VAR"],
            env=env,
        )

        assert cmd.exit_code == 0
        assert "test_value" in cmd.stdout

    def test_run_truncates_long_output(self):
        """Test that long output is truncated."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        term_client = TerminalClient(mock_client)

        cmd, result = term_client.run(
            trace_id="trace-123",
            command="sh",
            args=["-c", "yes | head -1000"],
            max_output_bytes=100,
        )

        assert len(cmd.stdout) <= 100


class TestRunFunction:
    """Tests for the run convenience function."""

    def test_run_raises_without_client(self):
        """Test that run raises error without client."""
        with patch('agenttrace.terminal.get_client', return_value=None):
            with pytest.raises(RuntimeError, match="No AgentTrace client"):
                run(["echo", "hello"])

    def test_run_raises_without_trace(self):
        """Test that run raises error without active trace."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        with patch('agenttrace.terminal.get_client', return_value=mock_client):
            with patch('agenttrace.terminal.get_current_trace', return_value=None):
                with pytest.raises(RuntimeError, match="No active trace"):
                    run(["echo", "hello"])

    def test_run_with_list_command(self):
        """Test run with list command."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        mock_trace = Mock()
        mock_trace.id = "trace-123"

        with patch('agenttrace.terminal.get_client', return_value=mock_client):
            with patch('agenttrace.terminal.get_current_trace', return_value=mock_trace):
                cmd, result = run(["echo", "hello"])
                assert cmd.command == "echo"

    def test_run_with_string_command(self):
        """Test run with string command (shell mode)."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        mock_trace = Mock()
        mock_trace.id = "trace-123"

        with patch('agenttrace.terminal.get_client', return_value=mock_client):
            with patch('agenttrace.terminal.get_current_trace', return_value=mock_trace):
                cmd, result = run("echo hello")
                assert cmd.exit_code == 0


class TestTerminalScope:
    """Tests for terminal_scope context manager."""

    def test_terminal_scope_basic(self):
        """Test basic terminal scope usage."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        with terminal_scope(
            mock_client,
            trace_id="trace-123",
            command="npm",
            args=["test"],
        ) as ctx:
            ctx["exit_code"] = 0
            ctx["stdout"] = "All tests passed"
            ctx["success"] = True

        mock_client._batch_queue.add.assert_called_once()
        call_args = mock_client._batch_queue.add.call_args[0][0]
        assert call_args['body']['command'] == 'npm'
        assert call_args['body']['exitCode'] == 0

    def test_terminal_scope_with_exception(self):
        """Test terminal scope handles exceptions."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        with pytest.raises(ValueError):
            with terminal_scope(
                mock_client,
                trace_id="trace-123",
                command="error-command",
            ) as ctx:
                raise ValueError("Test error")

        mock_client._batch_queue.add.assert_called_once()
        call_args = mock_client._batch_queue.add.call_args[0][0]
        assert call_args['body']['success'] is False
        assert call_args['body']['exitCode'] == -1

    def test_terminal_scope_with_options(self):
        """Test terminal scope with all options."""
        mock_client = Mock()
        mock_client.enabled = True
        mock_client._batch_queue = Mock()

        with terminal_scope(
            mock_client,
            trace_id="trace-123",
            command="full-command",
            args=["arg1", "arg2"],
            observation_id="obs-456",
            tool_name="test-tool",
            reason="testing",
        ) as ctx:
            ctx["exit_code"] = 0

        call_args = mock_client._batch_queue.add.call_args[0][0]
        assert call_args['body']['observationId'] == 'obs-456'
        assert call_args['body']['toolName'] == 'test-tool'
        assert call_args['body']['reason'] == 'testing'


class TestTerminalCommandInfo:
    """Tests for TerminalCommandInfo dataclass."""

    def test_terminal_command_info_creation(self):
        """Test creating TerminalCommandInfo."""
        now = datetime.utcnow()
        info = TerminalCommandInfo(
            id="cmd-123",
            trace_id="trace-456",
            observation_id="obs-789",
            command="npm",
            args=["test"],
            working_directory="/path/to/project",
            exit_code=0,
            stdout="output",
            stderr="",
            success=True,
            duration_ms=1000,
            started_at=now,
            completed_at=now,
        )

        assert info.id == "cmd-123"
        assert info.trace_id == "trace-456"
        assert info.observation_id == "obs-789"
        assert info.command == "npm"
        assert info.args == ["test"]
        assert info.working_directory == "/path/to/project"
        assert info.exit_code == 0
        assert info.stdout == "output"
        assert info.stderr == ""
        assert info.success is True
        assert info.duration_ms == 1000

    def test_terminal_command_info_optional_fields(self):
        """Test TerminalCommandInfo with optional fields as None."""
        now = datetime.utcnow()
        info = TerminalCommandInfo(
            id="cmd-123",
            trace_id="trace-456",
            observation_id=None,
            command="test",
            args=[],
            working_directory="/path",
            exit_code=0,
            stdout="",
            stderr="",
            success=True,
            duration_ms=0,
            started_at=now,
            completed_at=now,
        )

        assert info.observation_id is None

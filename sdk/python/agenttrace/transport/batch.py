"""
Batch queue for efficient event submission.
"""

from __future__ import annotations

import logging
import queue
import threading
import time
from typing import Any, Dict, List, Optional

from agenttrace.transport.http import HttpTransport

logger = logging.getLogger("agenttrace")


class BatchQueue:
    """
    Queue for batching events before sending to the API.

    Events are batched and sent either when the batch size is reached
    or when the flush interval expires.
    """

    def __init__(
        self,
        transport: HttpTransport,
        flush_at: int = 20,
        flush_interval: float = 5.0,
        max_queue_size: int = 10000,
    ):
        """
        Initialize the batch queue.

        Args:
            transport: HTTP transport for sending events
            flush_at: Number of events before auto-flush
            flush_interval: Seconds between auto-flush
            max_queue_size: Maximum queue size before dropping events
        """
        self._transport = transport
        self._flush_at = flush_at
        self._flush_interval = flush_interval
        self._max_queue_size = max_queue_size

        self._queue: queue.Queue[Dict[str, Any]] = queue.Queue(maxsize=max_queue_size)
        self._batch: List[Dict[str, Any]] = []
        self._batch_lock = threading.Lock()

        self._running = True
        self._flush_thread = threading.Thread(target=self._flush_loop, daemon=True)
        self._flush_thread.start()

    def add(self, event: Dict[str, Any]) -> bool:
        """
        Add an event to the queue.

        Args:
            event: Event dictionary

        Returns:
            True if added, False if queue is full
        """
        try:
            self._queue.put_nowait(event)
            return True
        except queue.Full:
            logger.warning("Event queue is full, dropping event")
            return False

    def flush(self) -> None:
        """Flush all pending events immediately."""
        # Drain the queue into the batch
        events: List[Dict[str, Any]] = []
        while True:
            try:
                event = self._queue.get_nowait()
                events.append(event)
            except queue.Empty:
                break

        with self._batch_lock:
            events.extend(self._batch)
            self._batch = []

        if events:
            self._send_batch(events)

    def stop(self) -> None:
        """Stop the flush thread and flush remaining events."""
        self._running = False
        self.flush()
        # Wait for flush thread to finish
        if self._flush_thread.is_alive():
            self._flush_thread.join(timeout=5.0)

    def _flush_loop(self) -> None:
        """Background thread for periodic flushing."""
        last_flush = time.time()

        while self._running:
            try:
                # Wait for events or timeout
                try:
                    event = self._queue.get(timeout=0.1)
                    with self._batch_lock:
                        self._batch.append(event)
                except queue.Empty:
                    pass

                # Check if we should flush
                now = time.time()
                with self._batch_lock:
                    should_flush = (
                        len(self._batch) >= self._flush_at or
                        (self._batch and now - last_flush >= self._flush_interval)
                    )

                    if should_flush:
                        events = self._batch
                        self._batch = []
                        last_flush = now

                if should_flush and events:
                    self._send_batch(events)

            except Exception as e:
                logger.error(f"Error in flush loop: {e}")
                time.sleep(0.1)

    def _send_batch(self, events: List[Dict[str, Any]]) -> None:
        """Send a batch of events to the API."""
        if not events:
            return

        try:
            success = self._transport.send_batch(events)
            if not success:
                logger.warning(f"Failed to send batch of {len(events)} events")
        except Exception as e:
            logger.error(f"Error sending batch: {e}")


class AsyncBatchQueue:
    """
    Async queue for batching events.

    For use in async applications where threading may not be appropriate.
    """

    def __init__(
        self,
        transport: HttpTransport,
        flush_at: int = 20,
        flush_interval: float = 5.0,
        max_queue_size: int = 10000,
    ):
        """Initialize the async batch queue."""
        self._transport = transport
        self._flush_at = flush_at
        self._flush_interval = flush_interval
        self._max_queue_size = max_queue_size

        self._batch: List[Dict[str, Any]] = []
        self._last_flush = time.time()

    def add(self, event: Dict[str, Any]) -> bool:
        """
        Add an event to the queue.

        This is synchronous - events are batched immediately.
        Call flush() or check_flush() to send.
        """
        if len(self._batch) >= self._max_queue_size:
            logger.warning("Event queue is full, dropping event")
            return False

        self._batch.append(event)

        # Auto-flush if batch is full
        if len(self._batch) >= self._flush_at:
            self.flush()

        return True

    def check_flush(self) -> None:
        """Check if a flush is needed based on interval."""
        if self._batch and time.time() - self._last_flush >= self._flush_interval:
            self.flush()

    def flush(self) -> None:
        """Flush all pending events."""
        if not self._batch:
            return

        events = self._batch
        self._batch = []
        self._last_flush = time.time()

        try:
            self._transport.send_batch(events)
        except Exception as e:
            logger.error(f"Error sending batch: {e}")

    def stop(self) -> None:
        """Flush remaining events."""
        self.flush()

    @property
    def pending_count(self) -> int:
        """Get the number of pending events."""
        return len(self._batch)

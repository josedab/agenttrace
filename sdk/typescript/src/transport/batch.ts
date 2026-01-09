/**
 * Batch queue for efficient event submission.
 */

import { HttpTransport, BatchEvent } from "./http";

export interface BatchQueueConfig {
  transport: HttpTransport;
  flushAt?: number;
  flushInterval?: number;
  maxQueueSize?: number;
}

/**
 * Queue for batching events before sending to the API.
 */
export class BatchQueue {
  private transport: HttpTransport;
  private flushAt: number;
  private flushInterval: number;
  private maxQueueSize: number;

  private queue: BatchEvent[] = [];
  private flushTimer: NodeJS.Timeout | null = null;
  private flushing: boolean = false;
  private stopped: boolean = false;

  constructor(config: BatchQueueConfig) {
    this.transport = config.transport;
    this.flushAt = config.flushAt ?? 20;
    this.flushInterval = config.flushInterval ?? 5000;
    this.maxQueueSize = config.maxQueueSize ?? 10000;

    this.startFlushTimer();
  }

  /**
   * Add an event to the queue.
   */
  add(event: Record<string, unknown>): boolean {
    if (this.stopped) {
      return false;
    }

    if (this.queue.length >= this.maxQueueSize) {
      console.warn("Event queue is full, dropping event");
      return false;
    }

    this.queue.push(event as BatchEvent);

    // Auto-flush if batch is full
    if (this.queue.length >= this.flushAt) {
      this.flush();
    }

    return true;
  }

  /**
   * Flush all pending events to the server.
   */
  async flush(): Promise<void> {
    if (this.flushing || this.queue.length === 0) {
      return;
    }

    this.flushing = true;

    // Take current batch
    const batch = this.queue.splice(0, this.queue.length);

    try {
      await this.transport.sendBatch(batch);
    } catch (error) {
      console.error("Error sending batch:", error);
    } finally {
      this.flushing = false;
    }
  }

  /**
   * Stop the queue and flush remaining events.
   */
  stop(): void {
    this.stopped = true;
    this.stopFlushTimer();
  }

  /**
   * Get the number of pending events.
   */
  get pendingCount(): number {
    return this.queue.length;
  }

  private startFlushTimer(): void {
    if (this.flushTimer !== null) {
      return;
    }

    this.flushTimer = setInterval(() => {
      if (this.queue.length > 0 && !this.flushing) {
        this.flush();
      }
    }, this.flushInterval);

    // Don't keep the process alive just for the timer
    if (typeof this.flushTimer.unref === "function") {
      this.flushTimer.unref();
    }
  }

  private stopFlushTimer(): void {
    if (this.flushTimer !== null) {
      clearInterval(this.flushTimer);
      this.flushTimer = null;
    }
  }
}

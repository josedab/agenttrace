import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@/lib/api', () => ({
  API_URL: 'https://api.example.test',
  getApiAccessToken: vi.fn(),
  getApiProjectId: vi.fn(),
}));

import { getApiAccessToken, getApiProjectId } from '@/lib/api';
import { useWebSocketStream } from '../use-websocket-stream';

const mockedGetApiAccessToken = vi.mocked(getApiAccessToken);
const mockedGetApiProjectId = vi.mocked(getApiProjectId);

class MockWebSocket {
  static instances: MockWebSocket[] = [];

  readonly url: string;
  readonly protocols: string[];
  onopen: ((event: Event) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onclose: ((event: CloseEvent) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  send = vi.fn();

  constructor(url: string, protocols: string[]) {
    this.url = url;
    this.protocols = protocols;
    MockWebSocket.instances.push(this);
  }

  close() {
    this.onclose?.(new CloseEvent('close'));
  }
}

describe('useWebSocketStream', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    MockWebSocket.instances = [];
    mockedGetApiAccessToken.mockReturnValue('header.jwt.token');
    mockedGetApiProjectId.mockReturnValue('project-1');
    vi.stubGlobal('WebSocket', MockWebSocket);
  });

  afterEach(() => {
    vi.clearAllTimers();
    vi.useRealTimers();
    vi.clearAllMocks();
    vi.unstubAllGlobals();
  });

  it('sends the token as a subprotocol instead of in the URL', () => {
    const { result, unmount } = renderHook(() => useWebSocketStream('trace-1'));
    const socket = MockWebSocket.instances[0];

    expect(socket.url).toBe('wss://api.example.test/ws/streaming/trace-1');
    expect(socket.url).not.toContain('header.jwt.token');
    expect(socket.protocols).toEqual(['agenttrace', 'header.jwt.token', 'project.project-1']);

    act(() => socket.onopen?.(new Event('open')));
    expect(result.current.connected).toBe(true);
    expect(socket.send).toHaveBeenCalledWith(
      JSON.stringify({ action: 'subscribe', traceId: 'trace-1' })
    );

    unmount();
    act(() => vi.advanceTimersByTime(3000));
    expect(MockWebSocket.instances).toHaveLength(1);
  });

  it('reconnects after an unexpected close', () => {
    const { unmount } = renderHook(() => useWebSocketStream('trace-1'));
    const socket = MockWebSocket.instances[0];

    act(() => socket.onclose?.(new CloseEvent('close')));
    expect(MockWebSocket.instances).toHaveLength(1);

    act(() => vi.advanceTimersByTime(3000));
    expect(MockWebSocket.instances).toHaveLength(2);

    unmount();
  });
});

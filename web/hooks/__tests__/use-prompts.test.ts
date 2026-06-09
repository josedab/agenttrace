import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import * as React from 'react';

vi.mock('@/lib/api', () => ({
  api: {
    prompts: {
      list: vi.fn(),
      get: vi.fn(),
      listVersions: vi.fn(),
      getVersion: vi.fn(),
      create: vi.fn(),
      createVersion: vi.fn(),
      setLabel: vi.fn(),
      removeLabel: vi.fn(),
      delete: vi.fn(),
      compile: vi.fn(),
    },
  },
}));

import {
  usePrompts,
  usePrompt,
  usePromptVersions,
  useCreatePrompt,
  useUpdatePrompt,
  useDeletePrompt,
  useCompilePrompt,
} from '../use-prompts';
import { api, type Prompt, type PromptVersion } from '@/lib/api';

const mockedApi = vi.mocked(api, { deep: true, partial: true });
const timestamp = '2025-01-01T00:00:00.000Z';

function createPromptVersion(overrides: Partial<PromptVersion> = {}): PromptVersion {
  return {
    id: 'prompt-version-1',
    version: 1,
    prompt: '',
    messages: null,
    config: null,
    variables: [],
    createdAt: timestamp,
    ...overrides,
  };
}

function createPrompt(overrides: Partial<Prompt> = {}): Prompt {
  return {
    id: 'prompt-1',
    name: 'prompt',
    type: 'TEXT',
    tags: [],
    activeVersion: null,
    versions: [],
    labels: {},
    versionCount: 0,
    createdAt: timestamp,
    updatedAt: timestamp,
    ...overrides,
  };
}

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
    },
  });
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return React.createElement(QueryClientProvider, { client: queryClient }, children);
  };
}

describe('usePrompts', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches all prompts with default filters', async () => {
    const mockPrompts = [
      createPrompt({ name: 'greeting', versionCount: 3 }),
      createPrompt({ id: 'prompt-2', name: 'summary', versionCount: 1 }),
    ];
    mockedApi.prompts.list.mockResolvedValue(mockPrompts);

    const { result } = renderHook(() => usePrompts(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.prompts.list).toHaveBeenCalledWith({});
    expect(result.current.data).toEqual(mockPrompts);
  });

  it('passes filters to the API', async () => {
    mockedApi.prompts.list.mockResolvedValue([]);

    const filters = { search: 'greeting', label: 'production' };
    const { result } = renderHook(() => usePrompts(filters), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.prompts.list).toHaveBeenCalledWith(filters);
  });

  it('handles errors', async () => {
    mockedApi.prompts.list.mockRejectedValue(new Error('Network error'));

    const { result } = renderHook(() => usePrompts(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error?.message).toBe('Network error');
  });
});

describe('usePrompt', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches a single prompt by name', async () => {
    const activeVersion = createPromptVersion({
      prompt: 'Hello {{name}}',
      version: 2,
    });
    const mockPrompt = createPrompt({
      name: 'greeting',
      activeVersion,
      versions: [activeVersion],
      versionCount: 2,
    });
    mockedApi.prompts.get.mockResolvedValue(mockPrompt);

    const { result } = renderHook(() => usePrompt('greeting'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.prompts.get).toHaveBeenCalledWith('greeting');
    expect(result.current.data).toEqual(mockPrompt);
  });

  it('does not fetch when promptName is empty', async () => {
    const { result } = renderHook(() => usePrompt(''), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe('idle');
    expect(mockedApi.prompts.get).not.toHaveBeenCalled();
  });
});

describe('usePromptVersions', () => {
  beforeEach(() => vi.clearAllMocks());

  it('fetches versions for a prompt', async () => {
    const mockVersions = [
      createPromptVersion({ version: 1, prompt: 'v1' }),
      createPromptVersion({
        id: 'prompt-version-2',
        version: 2,
        prompt: 'v2',
      }),
    ];
    mockedApi.prompts.listVersions.mockResolvedValue(mockVersions);

    const { result } = renderHook(() => usePromptVersions('greeting'), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(mockedApi.prompts.listVersions).toHaveBeenCalledWith('greeting');
    expect(result.current.data).toHaveLength(2);
  });

  it('does not fetch when promptName is empty', async () => {
    const { result } = renderHook(() => usePromptVersions(''), {
      wrapper: createWrapper(),
    });

    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useCreatePrompt', () => {
  beforeEach(() => vi.clearAllMocks());

  it('creates a prompt', async () => {
    mockedApi.prompts.create.mockResolvedValue(createPrompt({ name: 'new-prompt' }));

    const { result } = renderHook(() => useCreatePrompt(), {
      wrapper: createWrapper(),
    });

    await result.current.mutateAsync({
      name: 'new-prompt',
      prompt: 'Hello {{name}}',
    });

    expect(mockedApi.prompts.create).toHaveBeenCalledWith({
      name: 'new-prompt',
      prompt: 'Hello {{name}}',
    });
  });
});

describe('useUpdatePrompt', () => {
  beforeEach(() => vi.clearAllMocks());

  it('creates a new version of a prompt', async () => {
    mockedApi.prompts.createVersion.mockResolvedValue(createPromptVersion({ version: 2 }));

    const { result } = renderHook(() => useUpdatePrompt(), {
      wrapper: createWrapper(),
    });

    await result.current.mutateAsync({
      promptName: 'greeting',
      data: { prompt: 'Hi {{name}}!' },
    });

    expect(mockedApi.prompts.createVersion).toHaveBeenCalledWith('greeting', {
      prompt: 'Hi {{name}}!',
    });
  });
});

describe('useDeletePrompt', () => {
  beforeEach(() => vi.clearAllMocks());

  it('deletes a prompt', async () => {
    mockedApi.prompts.delete.mockResolvedValue(undefined);

    const { result } = renderHook(() => useDeletePrompt(), {
      wrapper: createWrapper(),
    });

    await result.current.mutateAsync('greeting');

    expect(mockedApi.prompts.delete).toHaveBeenCalledWith('greeting');
  });
});

describe('useCompilePrompt', () => {
  beforeEach(() => vi.clearAllMocks());

  it('compiles a prompt with variables', async () => {
    mockedApi.prompts.compile.mockResolvedValue({
      compiled: 'Hello Alice!',
      prompt: createPrompt(),
      version: 1,
      variables: { name: 'Alice' },
    });

    const { result } = renderHook(() => useCompilePrompt(), {
      wrapper: createWrapper(),
    });

    await result.current.mutateAsync({
      promptName: 'greeting',
      version: 1,
      variables: { name: 'Alice' },
    });

    expect(mockedApi.prompts.compile).toHaveBeenCalledWith('greeting', 1, { name: 'Alice' });
  });
});

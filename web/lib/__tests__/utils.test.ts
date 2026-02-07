import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  formatNumber,
  formatCost,
  formatLatency,
  formatRelativeTime,
  truncate,
  slugify,
  debounce,
  sleep,
  safeParseJSON,
  getInitials,
  isLightColor,
  stringToColor,
  formatBytes,
  extractPromptVariables,
  compilePrompt,
} from '../utils';

// ============================================
// formatNumber
// ============================================
describe('formatNumber', () => {
  it('returns plain number below 1000', () => {
    expect(formatNumber(0)).toBe('0');
    expect(formatNumber(42)).toBe('42');
    expect(formatNumber(999)).toBe('999');
  });

  it('formats thousands with K', () => {
    expect(formatNumber(1000)).toBe('1.0K');
    expect(formatNumber(1500)).toBe('1.5K');
    expect(formatNumber(999_999)).toBe('1000.0K');
  });

  it('formats millions with M', () => {
    expect(formatNumber(1_000_000)).toBe('1.0M');
    expect(formatNumber(2_500_000)).toBe('2.5M');
  });

  it('formats billions with B', () => {
    expect(formatNumber(1_000_000_000)).toBe('1.0B');
    expect(formatNumber(7_800_000_000)).toBe('7.8B');
  });
});

// ============================================
// formatCost
// ============================================
describe('formatCost', () => {
  it('shows 6 decimal places for very small costs', () => {
    expect(formatCost(0.000001)).toBe('$0.000001');
    expect(formatCost(0.009999)).toBe('$0.009999');
  });

  it('shows 4 decimal places for sub-dollar costs', () => {
    expect(formatCost(0.01)).toBe('$0.0100');
    expect(formatCost(0.5)).toBe('$0.5000');
    expect(formatCost(0.99)).toBe('$0.9900');
  });

  it('shows 2 decimal places for dollar amounts', () => {
    expect(formatCost(1)).toBe('$1.00');
    expect(formatCost(42.5)).toBe('$42.50');
    expect(formatCost(1000)).toBe('$1000.00');
  });
});

// ============================================
// formatLatency
// ============================================
describe('formatLatency', () => {
  it('shows microseconds for sub-millisecond', () => {
    expect(formatLatency(0.5)).toBe('500μs');
    expect(formatLatency(0.001)).toBe('1μs');
  });

  it('shows milliseconds', () => {
    expect(formatLatency(1)).toBe('1ms');
    expect(formatLatency(500)).toBe('500ms');
    expect(formatLatency(999)).toBe('999ms');
  });

  it('shows seconds for 1000ms+', () => {
    expect(formatLatency(1000)).toBe('1.00s');
    expect(formatLatency(2500)).toBe('2.50s');
    expect(formatLatency(60000)).toBe('60.00s');
  });
});

// ============================================
// formatRelativeTime
// ============================================
describe('formatRelativeTime', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-02-20T12:00:00Z'));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('returns "just now" for less than 60 seconds', () => {
    const date = new Date('2026-02-20T11:59:30Z');
    expect(formatRelativeTime(date)).toBe('just now');
  });

  it('returns minutes for less than 60 minutes', () => {
    const date = new Date('2026-02-20T11:30:00Z');
    expect(formatRelativeTime(date)).toBe('30m ago');
  });

  it('returns hours for less than 24 hours', () => {
    const date = new Date('2026-02-20T06:00:00Z');
    expect(formatRelativeTime(date)).toBe('6h ago');
  });

  it('returns days for less than 7 days', () => {
    const date = new Date('2026-02-17T12:00:00Z');
    expect(formatRelativeTime(date)).toBe('3d ago');
  });

  it('returns formatted date for 7+ days', () => {
    const date = new Date('2026-02-01T12:00:00Z');
    const result = formatRelativeTime(date);
    expect(result).not.toContain('ago');
  });

  it('accepts string dates', () => {
    expect(formatRelativeTime('2026-02-20T11:55:00Z')).toBe('5m ago');
  });
});

// ============================================
// truncate
// ============================================
describe('truncate', () => {
  it('returns original string if shorter than max', () => {
    expect(truncate('hello', 10)).toBe('hello');
    expect(truncate('hello', 5)).toBe('hello');
  });

  it('truncates with ellipsis', () => {
    expect(truncate('hello world', 8)).toBe('hello...');
    expect(truncate('abcdefghij', 6)).toBe('abc...');
  });
});

// ============================================
// slugify
// ============================================
describe('slugify', () => {
  it('converts to lowercase and replaces spaces with hyphens', () => {
    expect(slugify('Hello World')).toBe('hello-world');
  });

  it('removes special characters', () => {
    expect(slugify('Hello, World!')).toBe('hello-world');
  });

  it('trims leading and trailing hyphens', () => {
    expect(slugify('  hello  ')).toBe('hello');
    expect(slugify('---hello---')).toBe('hello');
  });

  it('collapses multiple separators', () => {
    expect(slugify('hello   world')).toBe('hello-world');
  });
});

// ============================================
// debounce
// ============================================
describe('debounce', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('delays function execution', () => {
    const fn = vi.fn();
    const debounced = debounce(fn, 100);

    debounced();
    expect(fn).not.toHaveBeenCalled();

    vi.advanceTimersByTime(100);
    expect(fn).toHaveBeenCalledOnce();
  });

  it('resets timer on subsequent calls', () => {
    const fn = vi.fn();
    const debounced = debounce(fn, 100);

    debounced();
    vi.advanceTimersByTime(50);
    debounced();
    vi.advanceTimersByTime(50);
    expect(fn).not.toHaveBeenCalled();

    vi.advanceTimersByTime(50);
    expect(fn).toHaveBeenCalledOnce();
  });

  it('passes arguments to the debounced function', () => {
    const fn = vi.fn();
    const debounced = debounce(fn, 100);

    debounced('arg1', 'arg2');
    vi.advanceTimersByTime(100);
    expect(fn).toHaveBeenCalledWith('arg1', 'arg2');
  });
});

// ============================================
// sleep
// ============================================
describe('sleep', () => {
  it('resolves after the specified delay', async () => {
    vi.useFakeTimers();
    const promise = sleep(100);
    vi.advanceTimersByTime(100);
    await promise;
    vi.useRealTimers();
  });
});

// ============================================
// safeParseJSON
// ============================================
describe('safeParseJSON', () => {
  it('parses valid JSON', () => {
    expect(safeParseJSON('{"key":"value"}', {})).toEqual({ key: 'value' });
    expect(safeParseJSON('[1,2,3]', [])).toEqual([1, 2, 3]);
  });

  it('returns fallback for invalid JSON', () => {
    expect(safeParseJSON('not json', 'default')).toBe('default');
    expect(safeParseJSON('', null)).toBe(null);
    expect(safeParseJSON('{broken', {})).toEqual({});
  });
});

// ============================================
// getInitials
// ============================================
describe('getInitials', () => {
  it('returns first two initials', () => {
    expect(getInitials('John Doe')).toBe('JD');
    expect(getInitials('Alice Bob Charlie')).toBe('AB');
  });

  it('returns single initial for single name', () => {
    expect(getInitials('John')).toBe('J');
  });

  it('returns uppercase', () => {
    expect(getInitials('john doe')).toBe('JD');
  });
});

// ============================================
// isLightColor
// ============================================
describe('isLightColor', () => {
  it('identifies white as light', () => {
    expect(isLightColor('#ffffff')).toBe(true);
    expect(isLightColor('#FFFFFF')).toBe(true);
  });

  it('identifies black as dark', () => {
    expect(isLightColor('#000000')).toBe(false);
  });

  it('identifies yellow as light', () => {
    expect(isLightColor('#ffff00')).toBe(true);
  });

  it('identifies dark blue as dark', () => {
    expect(isLightColor('#000080')).toBe(false);
  });
});

// ============================================
// stringToColor
// ============================================
describe('stringToColor', () => {
  it('returns a hex color string', () => {
    const color = stringToColor('test');
    expect(color).toMatch(/^#[0-9a-f]{6}$/);
  });

  it('is deterministic for the same input', () => {
    expect(stringToColor('hello')).toBe(stringToColor('hello'));
  });

  it('produces different colors for different inputs', () => {
    expect(stringToColor('hello')).not.toBe(stringToColor('world'));
  });
});

// ============================================
// formatBytes
// ============================================
describe('formatBytes', () => {
  it('returns 0 B for zero', () => {
    expect(formatBytes(0)).toBe('0 B');
  });

  it('formats bytes', () => {
    expect(formatBytes(500)).toBe('500 B');
  });

  it('formats kilobytes', () => {
    expect(formatBytes(1024)).toBe('1 KB');
    expect(formatBytes(1536)).toBe('1.5 KB');
  });

  it('formats megabytes', () => {
    expect(formatBytes(1048576)).toBe('1 MB');
  });

  it('formats gigabytes', () => {
    expect(formatBytes(1073741824)).toBe('1 GB');
  });
});

// ============================================
// extractPromptVariables
// ============================================
describe('extractPromptVariables', () => {
  it('extracts variables from template', () => {
    expect(extractPromptVariables('Hello {{name}}')).toEqual(['name']);
  });

  it('extracts multiple unique variables', () => {
    expect(extractPromptVariables('{{name}} is {{age}} years old')).toEqual(['name', 'age']);
  });

  it('deduplicates variables', () => {
    expect(extractPromptVariables('{{name}} and {{name}}')).toEqual(['name']);
  });

  it('returns empty for no variables', () => {
    expect(extractPromptVariables('Hello world')).toEqual([]);
  });

  it('handles empty string', () => {
    expect(extractPromptVariables('')).toEqual([]);
  });
});

// ============================================
// compilePrompt
// ============================================
describe('compilePrompt', () => {
  it('replaces variables with values', () => {
    expect(compilePrompt('Hello {{name}}', { name: 'World' })).toBe('Hello World');
  });

  it('replaces multiple variables', () => {
    const result = compilePrompt('{{greeting}} {{name}}!', {
      greeting: 'Hi',
      name: 'Alice',
    });
    expect(result).toBe('Hi Alice!');
  });

  it('leaves unmatched variables as-is', () => {
    expect(compilePrompt('Hello {{name}}', {})).toBe('Hello {{name}}');
  });

  it('handles template with no variables', () => {
    expect(compilePrompt('Hello world', { name: 'unused' })).toBe('Hello world');
  });
});

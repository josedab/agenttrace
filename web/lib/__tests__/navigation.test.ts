import { describe, expect, it } from 'vitest';

import { safeCallbackUrl } from '../navigation';

describe('safeCallbackUrl', () => {
  it('accepts local paths', () => {
    expect(safeCallbackUrl('/traces?level=ERROR')).toBe('/traces?level=ERROR');
  });

  it('rejects executable and cross-origin URLs', () => {
    expect(safeCallbackUrl('javascript:alert(1)')).toBe('/dashboard');
    expect(safeCallbackUrl('https://attacker.example')).toBe('/dashboard');
    expect(safeCallbackUrl('//attacker.example')).toBe('/dashboard');
  });
});

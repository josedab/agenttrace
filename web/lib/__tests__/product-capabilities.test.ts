import { describe, expect, it } from 'vitest';

import {
  compatibilityRedirects,
  getCapabilityForPath,
  primaryNavigationCapabilities,
} from '@/lib/product-capabilities';

describe('product capabilities', () => {
  it('exposes only the five approved product workflows', () => {
    expect(primaryNavigationCapabilities.map((capability) => capability.id)).toEqual([
      'trace-replay',
      'eval-hub',
      'prompts',
      'cost-center',
      'collaboration',
    ]);
  });

  it('maps nested canonical routes to their capability', () => {
    expect(getCapabilityForPath('/analytics/cost/models')?.id).toBe('cost-center');
    expect(getCapabilityForPath('/traces/trace-id')?.id).toBe('trace-replay');
    expect(getCapabilityForPath('/replay')?.id).toBe('trace-replay');
  });

  it('keeps legacy routes as non-permanent compatibility redirects', () => {
    const legacyRoutes = new Set(compatibilityRedirects.map((redirect) => redirect.source));

    expect(legacyRoutes.has('/eval-marketplace')).toBe(true);
    expect(legacyRoutes.has('/cost-attribution')).toBe(true);
    expect(legacyRoutes.has('/team')).toBe(true);
    expect(compatibilityRedirects.every((redirect) => redirect.permanent === false)).toBe(true);
  });
});

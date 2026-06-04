/**
 * Idempotency keys let the API recognize a repeated request, so a double-click
 * or a retry after a failed response never starts a second unit of work.
 */
export function createIdempotencyKey(): string {
  const globalCrypto = typeof globalThis === 'undefined' ? undefined : globalThis.crypto;
  if (globalCrypto?.randomUUID) {
    return globalCrypto.randomUUID();
  }
  if (globalCrypto?.getRandomValues) {
    const bytes = globalCrypto.getRandomValues(new Uint8Array(16));
    return Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('');
  }
  // Non-secure contexts still need a distinct key per attempt; collisions only
  // risk reusing a prior result, never leaking data across projects.
  return `${Date.now().toString(16)}-${Math.random().toString(16).slice(2, 14)}`;
}

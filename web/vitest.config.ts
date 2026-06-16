import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./vitest.setup.ts'],
    include: ['**/*.test.{ts,tsx}'],
    exclude: ['node_modules', '.next', 'dist', 'e2e'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      include: ['lib/**', 'hooks/**', 'components/**'],
      exclude: ['**/*.test.*', '**/*.d.ts'],
      // Repository baseline; raise these floors as coverage expands.
      // Current measured global coverage (statements 12.11%, branches 10.86%,
      // functions 8.58%, lines 12.23%) retains ~0.5pp headroom above each floor.
      thresholds: {
        lines: 11,
        branches: 10,
        functions: 8,
        statements: 11,
      },
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, '.'),
    },
  },
});

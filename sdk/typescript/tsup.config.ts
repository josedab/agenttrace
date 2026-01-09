import { defineConfig } from "tsup";

export default defineConfig([
  // Main Node.js SDK
  {
    entry: ["src/index.ts"],
    format: ["cjs", "esm"],
    dts: true,
    clean: true,
    sourcemap: true,
    splitting: false,
    treeshake: true,
    minify: false,
    outDir: "dist",
  },
  // Browser SDK
  {
    entry: ["browser/src/index.ts"],
    format: ["cjs", "esm"],
    dts: true,
    clean: false,
    sourcemap: true,
    splitting: false,
    treeshake: true,
    minify: false,
    outDir: "dist/browser",
    platform: "browser",
  },
]);

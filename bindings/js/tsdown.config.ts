import { defineConfig } from 'tsdown'

export default defineConfig({
    entry: ['./src/index.ts'],
    platform: 'node',
    clean: true,
    dts: true,
    format: ['esm', 'cjs']
});

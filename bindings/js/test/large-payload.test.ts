import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import { describe, it, expect } from 'vitest';
import { minify } from '@tdewolff/minify';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const benchmarkPath = (name: string) => path.resolve(__dirname, '../../../_benchmarks', name);
const readBenchmark = (name: string) => fs.readFileSync(benchmarkPath(name), 'utf8');

const largeHtml = readBenchmark('sample_es6.html');
const largeCss = readBenchmark('sample_bootstrap.css');
const largeJs = readBenchmark('sample_typescript.js');

describe('minify large payloads', () => {
    it('minifies a large HTML document without crashing', async () => {
        const output = await minify({
            data: largeHtml,
            type: 'text/html',
            htmlKeepDocumentTags: true,
        });

        expect(typeof output).toBe('string');
        expect(output.length).toBeLessThan(largeHtml.length);
        expect(output.length).toBeGreaterThan(0);
    }, 30_000);

    it('minifies a large CSS bundle', async () => {
        const output = await minify({
            data: largeCss,
            type: 'text/css',
        });

        expect(typeof output).toBe('string');
        expect(output.length).toBeLessThan(largeCss.length);
        expect(output.length).toBeGreaterThan(0);
    }, 30_000);

    it('minifies a large JavaScript bundle', async () => {
        const output = await minify({
            data: largeJs,
            type: 'application/javascript',
        });

        expect(typeof output).toBe('string');
        expect(output.length).toBeLessThan(largeJs.length);
        expect(output.length).toBeGreaterThan(1000);
    }, 30_000);
});

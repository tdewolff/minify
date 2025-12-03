import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import { describe, it, expect } from 'vitest';
import { minify } from '@tdewolff/minify';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const data = fs.readFileSync(path.resolve(__dirname, '../../../_benchmarks/sample_amazon.html'), 'utf8');

function getProcessMemoryMb() {
    return process.memoryUsage().rss / 1024 / 1024;
}

describe('minify', () => {
    it('does not leak memory across repeated HTML minifications', async () => {
        const memoryStats = [];
        for (let i = 0; i <= 100; i++) {
            // const cTs = Date.now();
            await minify({ data, type: 'text/html' });
            // const eTs = Date.now();
            const currentMemoryMb = getProcessMemoryMb();
            memoryStats.push(currentMemoryMb);
            // if (i % 10 === 0) {
            //     console.log({
            //         iteration: i,
            //         timeMs: eTs - cTs,
            //         memoryMb: currentMemoryMb,
            //     });
            // }
        }

        const sortedMemoryStats = [...memoryStats].sort((a, b) => a - b);
        const point25PercentileMemoryMb = sortedMemoryStats[Math.floor(sortedMemoryStats.length * 0.25)];
        const point75PercentileMemoryMb = sortedMemoryStats[Math.floor(sortedMemoryStats.length * 0.75)];

        console.log({
            point25PercentileMemoryMb,
            point75PercentileMemoryMb,
        });

        expect(point25PercentileMemoryMb).toBeDefined();
        expect(point75PercentileMemoryMb).toBeDefined();
        expect(point75PercentileMemoryMb).toBeLessThanOrEqual(point25PercentileMemoryMb! * 1.5);
    }, 30000);
});

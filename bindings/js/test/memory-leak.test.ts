import fs from 'node:fs';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import { tmpdir } from 'node:os';
import { fileURLToPath } from 'node:url';
import { describe, it, expect } from 'vitest';

type MemoryRunResult = {
    samples: number[];
    sampleEvery: number;
};

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const runnerPath = path.resolve(__dirname, './helpers/memory-leak-runner.mjs');

const MB = 1024 * 1024;

function runMemoryStats(): MemoryRunResult {
    const outputPath = path.join(tmpdir(), `minify-memory-${Math.random().toString(36).slice(2)}.json`);

    const { status, error, stderr } = spawnSync(process.execPath, ['--expose-gc', runnerPath], {
        cwd: path.resolve(__dirname, '..'),
        stdio: ['ignore', 'inherit', 'inherit'],
        env: {
            ...process.env,
            MINIFY_MEMORY_OUTPUT: outputPath,
        },
    });

    const output = fs.existsSync(outputPath) ? fs.readFileSync(outputPath, 'utf8').trim() : '';

    fs.rmSync(outputPath, { force: true });

    if (status !== 0) {
        if (error) {
            throw error;
        }

        throw new Error(`memory runner failed (${status}): ${stderr}`);
    }

    if (!output) {
        throw new Error('memory runner did not return any data');
    }

    try {
        return JSON.parse(output) as MemoryRunResult;
    } catch {
        throw new Error(`unable to parse memory runner output: ${output}`);
    }
}

function calculateSlope(samples: number[], sampleEvery: number): number {
    if (samples.length < 2) {
        return 0;
    }

    const xs = samples.map((_, index) => index * sampleEvery);
    const meanX = xs.reduce((sum, value) => sum + value, 0) / xs.length;
    const meanY = samples.reduce((sum, value) => sum + value, 0) / samples.length;

    const numerator = xs.reduce((sum, x, index) => {
        const sampleValue = samples[index]!;
        return sum + (x - meanX) * (sampleValue - meanY);
    }, 0);
    const denominator = xs.reduce((sum, x) => sum + (x - meanX) ** 2, 0);

    if (denominator === 0) {
        return 0;
    }

    return numerator / denominator;
}

describe('native bindings memory', () => {
    it('releases native buffers between minify calls', () => {
        const { samples, sampleEvery } = runMemoryStats();

        if (!samples || samples.length === 0) {
            throw new Error('memory runner returned no samples');
        }

        if (!sampleEvery) {
            throw new Error('memory runner returned an invalid sampling interval');
        }

        expect(samples.length).toBeGreaterThan(3);

        const baseline = samples[0]!;
        const maxGrowth = Math.max(...samples) - baseline;
        const endGrowth = samples[samples.length - 1]! - baseline;

        const recentSamples = samples.slice(-Math.min(samples.length, 6));
        const trendBytesPerIteration = calculateSlope(recentSamples, sampleEvery);

        const allowedGrowth = 32 * MB;
        const allowedTrend = 0.1 * MB;

        expect(maxGrowth).toBeLessThan(allowedGrowth);
        expect(endGrowth).toBeLessThan(allowedGrowth);
        expect(trendBytesPerIteration).toBeLessThan(allowedTrend);
    }, 60000);
});

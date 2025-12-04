import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { setTimeout as delay } from 'node:timers/promises';
import { minify } from '@tdewolff/minify';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const data = fs.readFileSync(path.resolve(__dirname, '../../../../_benchmarks/sample_amazon.html'), 'utf8');

const ITERATIONS = Number(process.env.MINIFY_MEMORY_ITERATIONS ?? '200');
const SAMPLE_EVERY = Number(process.env.MINIFY_MEMORY_SAMPLE_EVERY ?? '10');
const WARMUP = Number(process.env.MINIFY_MEMORY_WARMUP ?? '12');

async function settleMemory() {
    if (typeof global.gc === 'function') {
        global.gc();
        await delay(10);
        global.gc();
    } else {
        await delay(10);
    }
}

async function sampleMemory() {
    await settleMemory();
    return process.memoryUsage().rss;
}

async function main() {
    await settleMemory();

    for (let i = 0; i < WARMUP; i++) {
        await minify({
            data,
            type: 'text/html',
        });
    }

    const samples = [await sampleMemory()];

    for (let i = 0; i < ITERATIONS; i++) {
        await minify({
            data,
            type: 'text/html',
        });

        if ((i + 1) % SAMPLE_EVERY === 0) {
            samples.push(await sampleMemory());
        }
    }

    const payload = JSON.stringify({ samples, sampleEvery: SAMPLE_EVERY });
    const outputPath = process.env.MINIFY_MEMORY_OUTPUT;

    if (outputPath) {
        fs.writeFileSync(outputPath, payload, 'utf8');
        return;
    }

    process.stdout.write(payload);
}

main().catch((error) => {
    console.error(error);
    process.exitCode = 1;
});

import { Worker } from 'node:worker_threads';
import { describe, it, expect } from 'vitest';

const input = "<html><span style=\"color:#ff0000;\">A  phrase</span>";
const expected = "<html><span style=color:red>A phrase</span>";
const workerModuleUrl = new URL('./helpers/minify-worker.js', import.meta.url);

describe('minify worker thread', () => {
    it('minifies HTML using worker threads', async () => {
        const worker = new Worker(workerModuleUrl, {
            workerData: input
        });

        try {
            const output = await new Promise((resolve, reject) => {
                worker.on('message', resolve);
                worker.on('error', reject);
                worker.on('exit', (code) => {
                    if (code !== 0) {
                        reject(new Error(`Worker stopped with exit code ${code}`));
                    }
                });
            });

            expect(output).toBe(expected);
        } finally {
            await worker.terminate();
            expect(worker.threadId).toBe(-1); // ensure main process was not killed
        }
    });
});

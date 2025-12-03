import { minify } from '@tdewolff/minify';
import fs from 'fs';
import path from 'path';

const data = fs.readFileSync(path.join(import.meta.dirname, '../../../', '/_benchmarks/sample_amazon.html'), 'utf8').toString();

function getProcessMemoryMb() {
    return process.memoryUsage().rss / 1024 / 1024
}

async function test() {
    let memoryStats = [];
    for (let i = 0; i <= 300; i++) {
        const cTs = Date.now();
        await minify(data, { type: 'text/html', });
        const eTs = Date.now();
        const currentMemoryMb = getProcessMemoryMb();
        memoryStats.push(currentMemoryMb);
        if (i % 10 === 0) {
            console.log({
                i,
                timeMs: eTs - cTs,
                memory_mb: currentMemoryMb,
            })
        }
    }

    const point25PercentileMemoryMb = memoryStats.sort((a, b) => a - b)[Math.floor(memoryStats.length * 0.25)];
    const point75PercentileMemoryMb = memoryStats.sort((a, b) => a - b)[Math.floor(memoryStats.length * 0.75)];

    console.log({
        point25PercentileMemoryMb,
        point75PercentileMemoryMb,
    });

    if (point75PercentileMemoryMb > point25PercentileMemoryMb * 1.25) {
        throw new Error(`Likely memory leak detected! Multiple: ${point75PercentileMemoryMb / point25PercentileMemoryMb}x`);
    } else {
        console.log("No memory leak detected");
    }
}

test();
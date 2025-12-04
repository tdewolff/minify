import { parentPort, workerData } from 'node:worker_threads';
import { minify } from '@tdewolff/minify';

const minified = await minify({
    data: workerData,
    type: 'text/html',
    htmlKeepDocumentTags: true,
});

if (!parentPort) {
    throw new Error('parentPort is not available in this worker context.');
}

parentPort.postMessage(minified);

import { Worker, isMainThread, workerData, parentPort } from 'node:worker_threads';

if (isMainThread) {
    let input = "<html><span style=\"color:#ff0000;\">A  phrase</span>";
    const worker = new Worker(new URL(import.meta.url), {
        workerData: input,
    });

    let expected = "<html><span style=color:red>A phrase</span>";
    let output = await new Promise((resolve, reject) => {
        worker.on('message', resolve);
        worker.on('error', reject);
        worker.on('exit', (code) => {
            if (code !== 0)
                reject(new Error(`Worker stopped with exit code ${code}`));
        });
    })
    if (output != expected) {
        throw "unexpected output using worker threads: '" + output + "' instead of '" + expected + "'";
    }
    await worker.terminate();
    console.log("success!"); // must reach here to ensure node process was not killed
} else {
    const { minify } = await import('@tdewolff/minify');
    // config({'html-keep-document-tags': true})
    const minified = await minify(workerData, {
        type: 'text/html',
        htmlKeepDocumentTags: true,
    });
    parentPort.postMessage(minified);
}

import { config, string } from '@tdewolff/minify';
import { Worker } from 'node:worker_threads';

config({'html-keep-document-tags': true})

let input = "<html><span style=\"color:#ff0000;\">A  phrase</span>";
let expected = "<html><span style=color:red>A phrase</span>";
let output = string("text/html", input);
if (output != expected) {
    throw "unexpected output: '"+output+"' instead of '"+expected+"'";
}

const worker = new Worker(new URL("worker.js", import.meta.url), {
    workerData: input,
});
output = await new Promise((resolve, reject) => {
    worker.on('message', resolve);
    worker.on('error', reject);
    worker.on('exit', (code) => {
        if (code !== 0)
        reject(new Error(`Worker stopped with exit code ${code}`));
    });
})
if (output != expected) {
    throw "unexpected output using worker threads: '"+output+"' instead of '"+expected+"'";
}
await worker.terminate();

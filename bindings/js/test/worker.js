import { config, string } from '@tdewolff/minify';
import { workerData, parentPort } from 'node:worker_threads';

config({'html-keep-document-tags': true})
parentPort.postMessage(string("text/html", workerData));

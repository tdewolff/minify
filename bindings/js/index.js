import { createRequire } from "node:module";
import { dirname } from 'path';
import { fileURLToPath } from 'url';

const require = createRequire(import.meta.url);
const __dirname = dirname(fileURLToPath(import.meta.url));

// https://nodejs.org/api/esm.html#no-native-module-loading
export const { string, config, file } = require('node-gyp-build')(__dirname);

// Starting from node 16, json imports have been unflagged
// import pkg from "./package.json" assert { type: "json" }
export const version = require('./package.json').version;

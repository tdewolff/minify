import { createRequire } from "node:module";

const require = createRequire(import.meta.url);

// https://nodejs.org/api/esm.html#no-native-module-loading
export const { string, config, file } = require('./build/Release/obj.target/minify');

// Starting from node 16, json imports have been unflagged
// import pkg from "./package.json" assert { type: "json" }
export const version = require('./package.json').version;

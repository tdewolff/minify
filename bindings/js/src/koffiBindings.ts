import koffi, { type KoffiFunc } from 'koffi'
import { resolveLibPath } from './libPath.js'
import { promisify } from 'util'

const libPath = resolveLibPath(String(koffi.extension))
const lib = koffi.load(libPath)

// name, returnType, [argTypes]
const freeCString = lib.func('FreeCString', koffi.types.void, ['void*'])

// this tells koffi to automatically free the returned C string (minify_result) after use
koffi.disposable('minify_result', koffi.types.str, freeCString)


// name, returnType, [argTypes]
const minifyFn: KoffiFunc<(data: string, optionsJson: string) => string> = lib.func('MinifyString', 'minify_result', ['char*', 'char*'])

export function minifyString(data: string, optionsJson: string): string {
  return minifyFn(data, optionsJson)
}

export const minifyStringAsync = promisify(minifyFn.async.bind(minifyFn))

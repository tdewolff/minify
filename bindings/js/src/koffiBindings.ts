import koffi, { type KoffiFunc } from 'koffi'
import { resolveLibPath } from './libPath.js'
import { promisify } from 'util'
import type { MinifyOptions } from './types.js'
import { toInt } from './helpers.js'

const libPath = resolveLibPath(String(koffi.extension))
const lib = koffi.load(libPath)

const freeCString = lib.func('FreeCString', koffi.types.void, ['void*'])
const MinifyCString = koffi.disposable('MinifyCString', koffi.types.str, freeCString)

// Register the struct layouts once so the prototype string can reference them by name.
const MinifyOptionsStruct = koffi.struct('MinifyOptions', {
  mediatype: koffi.types.str,
  data: koffi.types.str,
  cssPrecision: koffi.types.int32_t,
  cssVersion: koffi.types.int32_t,
  htmlKeepComments: koffi.types.bool,
  htmlKeepConditionalComments: koffi.types.bool,
  htmlKeepDefaultAttrvals: koffi.types.bool,
  htmlKeepDocumentTags: koffi.types.bool,
  htmlKeepEndTags: koffi.types.bool,
  htmlKeepQuotes: koffi.types.bool,
  htmlKeepSpecialComments: koffi.types.bool,
  htmlKeepWhitespace: koffi.types.bool,
  jsKeepVarNames: koffi.types.bool,
  jsPrecision: koffi.types.int32_t,
  jsVersion: koffi.types.int32_t,
  jsonKeepNumbers: koffi.types.bool,
  jsonPrecision: koffi.types.int32_t,
  svgKeepComments: koffi.types.bool,
  svgPrecision: koffi.types.int32_t,
  xmlKeepWhitespace: koffi.types.bool
})

const MinifyResultStruct = koffi.struct('MinifyResult', {
  error: MinifyCString,
  data: MinifyCString
})

type NativeMinifyOptions = {
  mediatype: string
  data: string
  cssPrecision: number
  cssVersion: number
  htmlKeepComments: boolean
  htmlKeepConditionalComments: boolean
  htmlKeepDefaultAttrvals: boolean
  htmlKeepDocumentTags: boolean
  htmlKeepEndTags: boolean
  htmlKeepQuotes: boolean
  htmlKeepSpecialComments: boolean
  htmlKeepWhitespace: boolean
  jsKeepVarNames: boolean
  jsPrecision: number
  jsVersion: number
  jsonKeepNumbers: boolean
  jsonPrecision: number
  svgKeepComments: boolean
  svgPrecision: number
  xmlKeepWhitespace: boolean
}

export type NativeMinifyResult = {
  error?: string | null
  data?: string | null
}

// lib.func(name, returnType, argTypes[])
const minifyFn: KoffiFunc<(options: NativeMinifyOptions, out: NativeMinifyResult) => void> = lib.func('Minify', koffi.types.void, [
  // Options are input-only; pass by pointer.
  koffi.pointer(MinifyOptionsStruct),
  // Result is an output struct filled by the native code.
  koffi.out(koffi.pointer(MinifyResultStruct))
])

const minifyFnAsync = promisify(minifyFn.async.bind(minifyFn));

function normalizeOptions(opts: MinifyOptions): NativeMinifyOptions {
  const normalized = opts ?? {}
  return {
    mediatype: String(normalized.type ?? ''),
    data: String(normalized.data ?? ''),
    cssPrecision: toInt(normalized.cssPrecision),
    cssVersion: toInt(normalized.cssVersion),
    htmlKeepComments: Boolean(normalized.htmlKeepComments),
    htmlKeepConditionalComments: Boolean(normalized.htmlKeepConditionalComments),
    htmlKeepDefaultAttrvals: Boolean(normalized.htmlKeepDefaultAttrvals),
    htmlKeepDocumentTags: Boolean(normalized.htmlKeepDocumentTags),
    htmlKeepEndTags: Boolean(normalized.htmlKeepEndTags),
    htmlKeepQuotes: Boolean(normalized.htmlKeepQuotes),
    htmlKeepSpecialComments: Boolean(normalized.htmlKeepSpecialComments),
    htmlKeepWhitespace: Boolean(normalized.htmlKeepWhitespace),
    jsKeepVarNames: Boolean(normalized.jsKeepVarNames),
    jsPrecision: toInt(normalized.jsPrecision),
    jsVersion: toInt(normalized.jsVersion),
    jsonKeepNumbers: Boolean(normalized.jsonKeepNumbers),
    jsonPrecision: toInt(normalized.jsonPrecision),
    svgKeepComments: Boolean(normalized.svgKeepComments),
    svgPrecision: toInt(normalized.svgPrecision),
    xmlKeepWhitespace: Boolean(normalized.xmlKeepWhitespace)
  }
}

export async function minifyNativeAsync(options: MinifyOptions): Promise<NativeMinifyResult> {
  const result: NativeMinifyResult = {}
  await minifyFnAsync(normalizeOptions(options), result)
  return result
}


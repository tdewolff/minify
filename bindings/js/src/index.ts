import { minifyNativeAsync } from './koffiBindings.js'
import type { MinifyConfig, MinifyMediaType, MinifyOptions } from './types'
export type * from './types'

function resolveOptions(optsOrType: MinifyOptions | MinifyMediaType, data?: string, config?: MinifyConfig | null): MinifyOptions {
  if (typeof optsOrType === 'string') {
    if (typeof data !== 'string') throw new TypeError('minify data must be a string when using the (type, data, config) signature')
    if (config != null && typeof config !== 'object') throw new TypeError('minify config must be an object when provided')
    return { ...config, type: optsOrType, data }
  }

  if (!optsOrType || typeof optsOrType !== 'object') throw new TypeError('minify options must be an object')
  return optsOrType
}

export async function minify(opts: MinifyOptions): Promise<string>

export async function minify(type: MinifyMediaType, data: string, config?: MinifyConfig | null): Promise<string>

export async function minify(optsOrType: MinifyOptions | MinifyMediaType, data?: string, config?: MinifyConfig | null): Promise<string> {
  const result = await minifyNativeAsync(resolveOptions(optsOrType, data, config))
  if (result.error) throw new Error(result.error)
  if (typeof result.data !== 'string') throw new Error('Native response missing data')
  return result.data
}

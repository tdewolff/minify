import { minifyNativeAsync } from './koffiBindings.js'
import type { MinifyOptions } from './types'
export type * from './types'

export async function minify(opts: MinifyOptions): Promise<string> {
  const result = await minifyNativeAsync(opts ?? {})
  if (result.error) throw new Error(result.error)
  if (typeof result.data !== 'string') throw new Error('Native response missing data')
  return result.data
}

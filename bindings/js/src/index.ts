import { minifyStringAsync } from './koffiBindings.js'
import type { MinifyOptions } from './types'
export type * from './types'

export async function minify(opts: MinifyOptions): Promise<string> {
  const raw = await minifyStringAsync(JSON.stringify(opts ?? {}))
  const result = JSON.parse(raw);
  if (result.error) throw new Error(result.error)
  if (typeof result.data !== 'string') throw new Error('Native response missing data')
  return result.data
}

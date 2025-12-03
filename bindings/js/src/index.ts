import { minifyStringAsync } from './koffiBindings.js'
import type { MinifyOptions } from './types'
export type * from './types'

export async function minify(data: string, opts: MinifyOptions): Promise<string> {
  const raw = await minifyStringAsync(data, JSON.stringify(opts ?? {}))
  const result = JSON.parse(raw);
  if (result.error) throw new Error(result.error)
  if (typeof result.data !== 'string') throw new Error('Native response missing data')
  return result.data
}

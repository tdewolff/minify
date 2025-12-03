import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

export function resolveLibPath(extension: string): string {
  const override = process.env.NODE_MINIFY_LIB_PATH
  if (override && fs.existsSync(override)) {
    return override
  }

  const goPlatformDir = `${toGoOS(process.platform)}-${toGoArch(process.arch)}`
  const candidates = [path.resolve(__dirname, '..', 'build', goPlatformDir, `minify${extension}`)]

  for (const candidate of candidates) {
    if (fs.existsSync(candidate)) return candidate
  }

  throw new Error(
    `Native library not found. Looked in:\n${candidates
      .map(p => ' - ' + p)
      .join('\n')}\nRun: npm run build:go (optionally with GOOS/GOARCH) or set NODE_MINIFY_LIB_PATH`
  )
}

function toGoOS(platform: NodeJS.Platform): string {
  return platform === 'win32' ? 'windows' : platform
}

function toGoArch(arch: string): string {
  switch (arch) {
    case 'x64':
      return 'amd64'
    case 'ia32':
      return '386'
    default:
      return arch
  }
}

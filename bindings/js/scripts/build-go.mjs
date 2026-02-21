import { spawnSync } from 'child_process'
import { existsSync, mkdirSync, readFileSync } from 'fs'
import { dirname, join } from 'path'
import { fileURLToPath } from 'url'

if (typeof Bun !== "undefined") console.log(`Running build script with Bun runtime`);

const __filename = fileURLToPath(import.meta.url)
const __dirname = dirname(__filename)

const cliArgs = process.argv.slice(2)

const hostOs = toGoOS(process.platform)
const hostArch = toGoArch(process.arch)

const rawTargetOs = getArgValue('goos') || process.env.GOOS
const rawTargetArch = getArgValue('goarch') || process.env.GOARCH

const targetOs = rawTargetOs ? toGoOS(rawTargetOs) : hostOs
const targetArch = rawTargetArch ? toGoArch(rawTargetArch) : hostArch

const ext = getExt(targetOs)

const repoRoot = join(__dirname, '..')
const goRoot = join(repoRoot, 'go')
const buildRoot = join(repoRoot, 'build')
const outputDir = join(buildRoot, `${targetOs}-${targetArch}`)
const cacheDir = process.env.GOCACHE || join(repoRoot, '.cache', `${targetOs}-${targetArch}`, 'go-build')
const pkgVersion = readPkgVersion(join(repoRoot, 'package.json'))

const outputLib = join(outputDir, `minify${ext}`)
const skipBuild = process.env.NODE_MINIFY_SKIP_BUILD === '1' || hasArg('skip-build')
const forceBuild = process.env.NODE_MINIFY_FORCE_BUILD === '1' || hasArg('force-build')
const isDebugBuild = process.env.NODE_MINIFY_DEBUG_BUILD === '1' || hasArg('debug-build')

console.log(`Building minify bindings for ${targetOs}/${targetArch} -> ${outputLib}`)

if (skipBuild) {
  console.log('Skipping Go build because NODE_MINIFY_SKIP_BUILD=1 or --skip-build was provided (ensure the library exists at the expected path)')
  process.exit(0)
}

if (!forceBuild && existsSync(outputLib)) {
  console.log('Prebuilt library already present; skipping Go build. Set NODE_MINIFY_FORCE_BUILD=1 or pass --force-build to rebuild.')
  process.exit(0)
}

ensureDir(buildRoot)
ensureDir(outputDir)
ensureDir(cacheDir)
const localRootGoMod = join(goRoot, '..', '..', '..', 'go.mod')
const useLocalModule = existsSync(localRootGoMod)

const env = {
  ...process.env,
  GOOS: targetOs,
  GOARCH: targetArch,
  GOCACHE: cacheDir
}

if (!useLocalModule) {
  syncModuleVersion(env)
}

const tidyResult = spawnSync('go', ['mod', 'tidy'], {
  cwd: goRoot,
  env,
  stdio: 'inherit'
})
if (tidyResult.status !== 0) {
  process.exit(tidyResult.status ?? 1)
}

const goArgs = ['build', '-buildmode=c-shared', '-o', outputLib]

if (!isDebugBuild) {
  goArgs.push('-trimpath', '-ldflags=-s -w', '-buildvcs=false')
} else {
  console.log('Debug build requested (--debug-build or NODE_MINIFY_DEBUG_BUILD=1); skipping production strip flags.')
}

goArgs.push('.')

const result = spawnSync('go', goArgs, {
  cwd: goRoot,
  env,
  stdio: 'inherit'
})

if (result.status !== 0) {
  process.exit(result.status ?? 1)
}

function ensureDir(pathname) {
  if (!existsSync(pathname)) {
    mkdirSync(pathname, { recursive: true })
  }
}

function syncModuleVersion(env) {
  // Remove local replace for published installs and pull the latest module version.
  spawnSync('go', ['mod', 'edit', '-dropreplace', 'github.com/tdewolff/minify/v2'], {
    cwd: goRoot,
    env,
    stdio: 'inherit'
  })

  const targetVersion = resolveModuleVersion()

  const getResult = spawnSync('go', ['get', `github.com/tdewolff/minify/v2@${targetVersion}`], {
    cwd: goRoot,
    env,
    stdio: 'inherit'
  })
  if (getResult.status !== 0) {
    process.exit(getResult.status ?? 1)
  }
}

function readPkgVersion(pathname) {
  try {
    const contents = readFileSync(pathname, 'utf8')
    const parsed = JSON.parse(contents)
    return parsed?.version
  } catch {
    return undefined
  }
}

function resolveModuleVersion() {
  const override = process.env.NODE_MINIFY_GO_VERSION
  const raw = override || pkgVersion

  if (!raw || raw === '{VERSION}') {
    return 'latest'
  }

  return raw.startsWith('v') ? raw : `v${raw}`
}

function hasArg(name) {
  return cliArgs.some(arg => arg === `--${name}`)
}

function getArgValue(name) {
  for (let i = 0; i < cliArgs.length; i++) {
    const arg = cliArgs[i]
    if (arg === `--${name}`) {
      const next = cliArgs[i + 1]
      if (next && !next.startsWith('-')) {
        return next
      }
    }
    if (arg.startsWith(`--${name}=`)) {
      return arg.slice(name.length + 3)
    }
  }
  return undefined
}

function toGoOS(platform) {
  return platform === 'win32' ? 'windows' : platform
}

function toGoArch(arch) {
  switch (arch) {
    case 'x64':
      return 'amd64'
    case 'ia32':
      return '386'
    default:
      return arch
  }
}

function getExt(os) {
  switch (os) {
    case 'windows':
      return '.dll'
    case 'darwin':
      return '.dylib'
    default:
      return '.so'
  }
}

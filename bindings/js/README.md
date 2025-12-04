JavaScript bindings around the Go minifiers in [tdewolff/minify](https://github.com/tdewolff/minify). The package ships a small native library built from Go and exposes a single async `minify` function. Requires Node.js 20.19+.

## Quickstart
```bash
npm install @tdewolff/minify
```

```js
import { readFile, writeFile } from 'node:fs/promises'
import { minify } from '@tdewolff/minify'

// Inline string
const html = await minify({
  data: `<html><span class="text" style="color:#ff0000;">A  phrase</span></html>`,
  type: 'text/html',
  htmlKeepDocumentTags: true
})
console.log(html) // <html><span class=text style=color:red>A phrase</span></html>

// File input/output
const source = await readFile('example.html', 'utf8')
const minified = await minify({ data: source, type: 'text/html', htmlKeepDocumentTags: true })
await writeFile('example.min.html', minified, 'utf8')
```

## API
```ts
import { minify, type MinifyOptions } from '@tdewolff/minify'

declare function minify(opts: MinifyOptions): Promise<string>
```

`MinifyOptions` fields (all optional except `data` and `type`):

- `data`: string content to minify.
- `type`: mediatype used to pick the minifier (see below).
- `cssPrecision`, `cssVersion`
- `htmlKeepComments`, `htmlKeepConditionalComments`, `htmlKeepDefaultAttrvals`, `htmlKeepDocumentTags`, `htmlKeepEndTags`, `htmlKeepQuotes`, `htmlKeepSpecialComments`, `htmlKeepWhitespace`
- `jsKeepVarNames`, `jsPrecision`, `jsVersion`
- `jsonKeepNumbers`, `jsonPrecision`
- `svgKeepComments`, `svgPrecision`
- `xmlKeepWhitespace`

Errors are thrown for missing data, invalid types, or native parse errors.

## Mediatypes
These types are accepted by the Go minifiers (regex-style JSON/XML/JS matches are supported):

- `text/css`
- `text/html`
- `image/svg+xml`
- JavaScript media types matching `(application|text)/(x-)?(java|ecma|j|live)script` and `module`
- Any type ending in `/json` or `+json`, and `importmap`/`speculationrules`
- Any type ending in `/xml` or `+xml`

## Native build
`npm install` runs `npm run build:go`, which builds `build/<goos>-<goarch>/minify.{so|dll|dylib}` using Go 1.24+ and a C compiler. The library is resolved automatically based on `process.platform`/`process.arch`.

Useful knobs:
- `NODE_MINIFY_LIB_PATH`: point to an existing compiled library to skip detection.
- `NODE_MINIFY_SKIP_BUILD=1`: skip building (ensure the library already exists).
- `NODE_MINIFY_FORCE_BUILD=1`: rebuild even if a library is present.
- `NODE_MINIFY_DEBUG_BUILD=1`: keep symbols/paths (no strip flags).
- `GOOS`/`GOARCH`: cross-build, e.g. `GOOS=windows GOARCH=amd64 npm run build:go`.

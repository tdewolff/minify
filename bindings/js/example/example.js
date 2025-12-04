import { readFile, writeFile } from 'node:fs/promises'
import { minify } from '@tdewolff/minify'

async function run() {
  const inline = await minify({
    data: `<html><span class="text" style="color:#ff0000;">A  phrase</span></html>`,
    type: 'text/html',
    htmlKeepDocumentTags: true
  })
  console.log(inline)

  const sourcePath = new URL('./example.html', import.meta.url)
  const outputPath = new URL('./example.min.html', import.meta.url)

  const minifiedFile = await minify({
    data: await readFile(sourcePath, 'utf8'),
    type: 'text/html',
    htmlKeepDocumentTags: true
  })

  await writeFile(outputPath, minifiedFile, 'utf8')
  console.log(`Minified file written to ${outputPath.pathname}`)
}

run().catch(err => {
  console.error(err)
})

// polaris.docs.fairwinds.com is an S3 website. It serves page/index.html and
// redirects /page to /page/. VitePress writes page.html. Move those files
// after the build. Leave index.html and 404.html (the bucket error document).
import assert from 'node:assert/strict'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'

const KEEP = new Set(['index.html', '404.html'])

export function toS3DirIndex(root) {
  const htmlFiles = []
  function walk(dir) {
    for (const name of fs.readdirSync(dir)) {
      const file = path.join(dir, name)
      if (fs.statSync(file).isDirectory()) walk(file)
      else if (name.endsWith('.html') && !KEEP.has(name)) htmlFiles.push(file)
    }
  }
  walk(root)
  for (const file of htmlFiles) {
    const dir = file.slice(0, -'.html'.length)
    fs.mkdirSync(dir)
    fs.renameSync(file, path.join(dir, 'index.html'))
  }
  return htmlFiles.length
}

function selfCheck() {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 's3-dir-'))
  fs.writeFileSync(path.join(root, 'index.html'), 'home')
  fs.writeFileSync(path.join(root, '404.html'), 'missing')
  fs.writeFileSync(path.join(root, 'dashboard.html'), 'dash')
  fs.mkdirSync(path.join(root, 'checks'))
  fs.writeFileSync(path.join(root, 'checks', 'security.html'), 'sec')
  assert.equal(toS3DirIndex(root), 2)
  assert.equal(fs.readFileSync(path.join(root, 'index.html'), 'utf8'), 'home')
  assert.equal(fs.readFileSync(path.join(root, '404.html'), 'utf8'), 'missing')
  assert.equal(fs.readFileSync(path.join(root, 'dashboard', 'index.html'), 'utf8'), 'dash')
  assert.equal(fs.readFileSync(path.join(root, 'checks', 'security', 'index.html'), 'utf8'), 'sec')
  assert.equal(fs.existsSync(path.join(root, 'dashboard.html')), false)
  assert.equal(fs.existsSync(path.join(root, 'checks', 'security.html')), false)
  fs.rmSync(root, { recursive: true })
  console.log('s3-dir-index ok')
}

if (process.argv[1] && path.resolve(process.argv[1]) === path.resolve(new URL(import.meta.url).pathname)) {
  if (process.argv.includes('--self-check')) selfCheck()
  else {
    const out = path.resolve(process.argv[2] || '.')
    const moved = toS3DirIndex(out)
    console.log(`s3-dir-index: moved ${moved} pages in ${out}`)
  }
}

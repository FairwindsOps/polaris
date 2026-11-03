// polaris.docs.fairwinds.com is an S3 website. It serves page/index.html and
// redirects /page to /page/. VitePress writes page.html. Move those files
// after the build. Leave index.html and 404.html (the bucket error document).
import assert from 'node:assert/strict'
import { spawnSync } from 'node:child_process'
import crypto from 'node:crypto'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const KEEP = new Set(['index.html', '404.html'])
const here = path.dirname(fileURLToPath(import.meta.url))
const docsDir = path.resolve(here, '..')

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
    const dest = path.join(dir, 'index.html')
    // A parent page (checks.html) lands in the directory that already holds
    // its children (checks/security.html). Refuse to clobber an index that
    // VitePress already wrote.
    if (fs.existsSync(dest)) {
      throw new Error(`s3-dir-index: ${dest} already exists`)
    }
    fs.mkdirSync(dir, { recursive: true })
    fs.renameSync(file, dest)
  }
  return htmlFiles.length
}

// ponytail: stamp ignores node_modules and .vitepress/cache. The lockfile is
// hashed with the rest of docs/, so a dependency change still rebuilds.
function sourceStamp(dir, readmePath) {
  const hash = crypto.createHash('sha256')
  const files = []
  function walk(current) {
    for (const name of fs.readdirSync(current)) {
      if (name === 'node_modules') continue
      const file = path.join(current, name)
      if (file === path.join(dir, 'index.md')) continue
      if (path.basename(current) === '.vitepress' && (name === 'cache' || name === 'dist')) continue
      if (fs.statSync(file).isDirectory()) walk(file)
      else files.push(file)
    }
  }
  walk(dir)
  if (readmePath && fs.existsSync(readmePath)) files.push(readmePath)
  for (const file of files.sort()) {
    hash.update(path.relative(path.dirname(dir), file))
    hash.update('\0')
    hash.update(fs.readFileSync(file))
    hash.update('\0')
  }
  return hash.digest('hex')
}

function distReady(dist) {
  return fs.existsSync(path.join(dist, 'index.html'))
    && fs.existsSync(path.join(dist, '404.html'))
    && fs.existsSync(path.join(dist, 'dashboard', 'index.html'))
    && !fs.existsSync(path.join(dist, 'dashboard.html'))
}

function run(cmd, args, cwd) {
  const result = spawnSync(cmd, args, { cwd, stdio: 'inherit' })
  if (result.status !== 0) process.exit(result.status ?? 1)
}

function buildSite() {
  const dist = path.resolve(docsDir, '../dist')
  const stampPath = path.join(docsDir, '.vitepress', 'cache', 'site-stamp')
  const readme = path.resolve(docsDir, '../README.md')
  const stamp = sourceStamp(docsDir, readme)
  if (fs.existsSync(stampPath) && fs.readFileSync(stampPath, 'utf8') === stamp && distReady(dist)) {
    console.log('docs build: up to date')
    return
  }
  const npm = process.platform === 'win32' ? 'npm.cmd' : 'npm'
  const vitepress = path.join(docsDir, 'node_modules', '.bin', process.platform === 'win32' ? 'vitepress.cmd' : 'vitepress')
  run(npm, ['run', 'build:readme'], docsDir)
  run(vitepress, ['build'], docsDir)
  const moved = toS3DirIndex(dist)
  console.log(`s3-dir-index: moved ${moved} pages in ${dist}`)
  fs.mkdirSync(path.dirname(stampPath), { recursive: true })
  fs.writeFileSync(stampPath, stamp)
}

function selfCheck() {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 's3-dir-'))
  const clash = fs.mkdtempSync(path.join(os.tmpdir(), 's3-clash-'))
  const stampDir = fs.mkdtempSync(path.join(os.tmpdir(), 's3-stamp-'))
  try {
    fs.writeFileSync(path.join(root, 'index.html'), 'home')
    fs.writeFileSync(path.join(root, '404.html'), 'missing')
    fs.writeFileSync(path.join(root, 'dashboard.html'), 'dash')
    fs.writeFileSync(path.join(root, 'checks.html'), 'parent')
    fs.mkdirSync(path.join(root, 'checks'))
    fs.writeFileSync(path.join(root, 'checks', 'security.html'), 'sec')
    assert.equal(toS3DirIndex(root), 3)
    assert.equal(fs.readFileSync(path.join(root, 'index.html'), 'utf8'), 'home')
    assert.equal(fs.readFileSync(path.join(root, '404.html'), 'utf8'), 'missing')
    assert.equal(fs.readFileSync(path.join(root, 'dashboard', 'index.html'), 'utf8'), 'dash')
    assert.equal(fs.readFileSync(path.join(root, 'checks', 'index.html'), 'utf8'), 'parent')
    assert.equal(fs.readFileSync(path.join(root, 'checks', 'security', 'index.html'), 'utf8'), 'sec')
    assert.equal(fs.existsSync(path.join(root, 'dashboard.html')), false)
    assert.equal(fs.existsSync(path.join(root, 'checks.html')), false)
    assert.equal(fs.existsSync(path.join(root, 'checks', 'security.html')), false)

    fs.mkdirSync(path.join(clash, 'checks'))
    fs.writeFileSync(path.join(clash, 'checks', 'index.html'), 'keep')
    fs.writeFileSync(path.join(clash, 'checks.html'), 'page')
    assert.throws(() => toS3DirIndex(clash), /already exists/)
    assert.equal(fs.readFileSync(path.join(clash, 'checks', 'index.html'), 'utf8'), 'keep')

    fs.writeFileSync(path.join(stampDir, 'a.md'), 'a')
    fs.mkdirSync(path.join(stampDir, 'node_modules'))
    fs.writeFileSync(path.join(stampDir, 'node_modules', 'x.js'), 'x')
    fs.mkdirSync(path.join(stampDir, '.vitepress', 'cache'), { recursive: true })
    fs.writeFileSync(path.join(stampDir, '.vitepress', 'cache', 't'), 't')
    const first = sourceStamp(stampDir)
    fs.writeFileSync(path.join(stampDir, 'node_modules', 'x.js'), 'y')
    fs.writeFileSync(path.join(stampDir, '.vitepress', 'cache', 't'), 'u')
    assert.equal(sourceStamp(stampDir), first)
    fs.writeFileSync(path.join(stampDir, 'a.md'), 'b')
    assert.notEqual(sourceStamp(stampDir), first)
    console.log('s3-dir-index ok')
  } finally {
    fs.rmSync(root, { recursive: true, force: true })
    fs.rmSync(clash, { recursive: true, force: true })
    fs.rmSync(stampDir, { recursive: true, force: true })
  }
}

if (process.argv[1] && path.resolve(process.argv[1]) === path.resolve(fileURLToPath(import.meta.url))) {
  if (process.argv.includes('--self-check')) selfCheck()
  else if (process.argv.includes('--build-site')) buildSite()
  else {
    const out = path.resolve(process.argv[2] || '.')
    const moved = toS3DirIndex(out)
    console.log(`s3-dir-index: moved ${moved} pages in ${out}`)
  }
}

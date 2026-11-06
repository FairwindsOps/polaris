// This file is generated from FairwindsOps/documentation-template
// DO NOT EDIT MANUALLY
//
// Project docs hosts are S3 websites. They serve page/index.html and
// redirect /page to /page/. The VitePress config rewrites foo.md to
// foo/index.md so the build already writes foo/index.html. This mover is
// the fallback for a leftover page.html. Leave index.html and 404.html
// (the bucket error document).
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

export function distReady(dist) {
  if (!fs.existsSync(path.join(dist, 'index.html'))) return false
  if (!fs.existsSync(path.join(dist, '404.html'))) return false
  function walk(dir) {
    for (const name of fs.readdirSync(dir)) {
      const file = path.join(dir, name)
      if (fs.statSync(file).isDirectory()) {
        if (!walk(file)) return false
      } else if (name.endsWith('.html') && !KEEP.has(name)) {
        return false
      }
    }
    return true
  }
  return walk(dist)
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

function run(cmd, args, cwd) {
  const result = spawnSync(cmd, args, { cwd, stdio: 'inherit' })
  if (result.status !== 0) process.exit(result.status ?? 1)
}

// README images are relative to the repo root (`img/logo.svg`). The homepage
// is written under docs/, and those files are served from docs/public.
function publishRootImg(line) {
  return line
    .replace(/(\bsrc\s*=\s*["'])img\//g, '$1/img/')
    .replace(/(\]\()img\//g, '$1/img/')
}

// Drop a Documentation heading and the blurb under it, through the next
// heading of any level. Project READMEs use `# Documentation` and then h2
// sections; those sections are the homepage. Lines that merely mention
// documentation stay, including a sample inside a fence. Docs-host URLs
// are removed so the homepage does not link at itself.
export function homepageFromReadme(readme) {
  const lines = String(readme).split(/\r?\n/)
  const out = []
  let skipping = false
  let fence = null
  for (const line of lines) {
    const fenceMatch = /^(\s*)(`{3,}|~{3,})/.exec(line)
    if (fenceMatch) {
      const marker = fenceMatch[2][0]
      const size = fenceMatch[2].length
      if (!fence) fence = { marker, size }
      else if (fence.marker === marker && size >= fence.size) fence = null
      if (!skipping) out.push(line)
      continue
    }
    if (!fence) {
      const heading = /^(#{1,6})[ \t]+(.+?)[ \t]*$/.exec(line)
      if (heading) {
        const text = heading[2].replace(/\s+#+\s*$/, '').trim()
        if (/^documentation$/i.test(text)) {
          skipping = true
          continue
        }
        if (skipping) skipping = false
      }
    }
    if (!skipping) {
      const published = line.replace(/https:\/\/[A-Za-z0-9.-]+\.docs\.fairwinds\.com/g, '')
      out.push(fence ? line : publishRootImg(published))
    }
  }
  return out.join('\n')
}

function writeHomepage() {
  const indexPath = path.join(docsDir, 'index.md')
  const metaPath = path.join(docsDir, 'main-metadata.md')
  const readmePath = path.resolve(docsDir, '../README.md')
  const meta = fs.existsSync(metaPath) ? fs.readFileSync(metaPath, 'utf8') : ''
  const readme = fs.existsSync(readmePath) ? fs.readFileSync(readmePath, 'utf8') : ''
  let body = homepageFromReadme(readme)
  if (!body.endsWith('\n')) body += '\n'
  const front = meta.trim() ? `${meta.replace(/\s*$/, '')}\n` : ''
  fs.writeFileSync(indexPath, front + body)
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
  const vitepress = path.join(docsDir, 'node_modules', '.bin', process.platform === 'win32' ? 'vitepress.cmd' : 'vitepress')
  writeHomepage()
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
    assert.equal(distReady(root), true)
    fs.writeFileSync(path.join(root, 'leftover.html'), 'x')
    assert.equal(distReady(root), false)

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

    const home = homepageFromReadme([
      '# Project',
      'Fairwinds hosts project documentation at `docs.fairwinds.com`.',
      '',
      '```md',
      '# Documentation',
      'Check out the [documentation](https://project.docs.fairwinds.com)',
      '```',
      '',
      'title: "PROJECT-NAME Documentation"',
      'Mirror of https://nova.docs.fairwinds.com stays local.',
      '',
      '## Documentation',
      'See https://goldilocks.docs.fairwinds.com',
      '',
      '## Install',
      'helm install',
    ].join('\n'))
    assert.match(home, /hosts project documentation/)
    assert.match(home, /# Documentation/)
    assert.match(home, /PROJECT-NAME Documentation/)
    assert.match(home, /## Install/)
    assert.doesNotMatch(home, /## Documentation/)
    assert.doesNotMatch(home, /goldilocks\.docs\.fairwinds\.com/)
    assert.doesNotMatch(home, /nova\.docs\.fairwinds\.com/)
    assert.match(home, /Mirror of  stays local/)
    assert.match(home, /project\.docs\.fairwinds\.com/)
    const h1 = homepageFromReadme([
      '# Goldilocks',
      'A utility.',
      '',
      '# Documentation',
      'Check out the [documentation](https://goldilocks.docs.fairwinds.com/)',
      '',
      '## How can this help',
      'VPA recommendations.',
      '',
      '## Quickstart',
      'helm install',
    ].join('\n'))
    assert.match(h1, /A utility/)
    assert.match(h1, /## How can this help/)
    assert.match(h1, /VPA recommendations/)
    assert.match(h1, /## Quickstart/)
    assert.match(h1, /helm install/)
    assert.doesNotMatch(h1, /^# Documentation$/m)
    assert.doesNotMatch(h1, /goldilocks\.docs\.fairwinds\.com/)
    const img = homepageFromReadme([
      '<img src="img/reckoner-logo.svg" alt="Reckoner" />',
      '![icon](img/reckoner-icon.svg)',
      '',
      '```html',
      '<img src="img/keep.svg" />',
      '```',
    ].join('\n'))
    assert.match(img, /src="\/img\/reckoner-logo.svg"/)
    assert.match(img, /\]\(\/img\/reckoner-icon.svg\)/)
    assert.match(img, /src="img\/keep.svg"/)
    console.log('s3-dir-index ok')
  } finally {
    fs.rmSync(root, { recursive: true, force: true })
    fs.rmSync(clash, { recursive: true, force: true })
    fs.rmSync(stampDir, { recursive: true, force: true })
  }
}

if (process.argv[1] && path.resolve(process.argv[1]) === path.resolve(fileURLToPath(import.meta.url))) {
  if (process.argv.includes('--self-check')) selfCheck()
  else if (process.argv.includes('--build-readme')) writeHomepage()
  else if (process.argv.includes('--build-site')) buildSite()
  else {
    const out = path.resolve(process.argv[2] || '.')
    const moved = toS3DirIndex(out)
    console.log(`s3-dir-index: moved ${moved} pages in ${out}`)
  }
}

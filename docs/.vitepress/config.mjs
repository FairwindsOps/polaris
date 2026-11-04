// This file is generated from FairwindsOps/documentation-template
// DO NOT EDIT MANUALLY
import { createRequire } from 'node:module'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import assert from 'node:assert/strict'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vitepress'

const require = createRequire(import.meta.url)
const here = path.dirname(fileURLToPath(import.meta.url))
const docsDir = path.resolve(here, '..')

// package.json overrides vite to 6.4.5. VitePress 1.6.4 declares vite ^5, and Vite 5
// does not get the 2026 dev-server fixes that shipped in 6.4.5 (GHSA-8g5g-mxqh-grgj).
// The published site is static; only `vitepress dev` runs that server.

const VUEPRESS_THEME_KEYS = [
  'docsRepo',
  'docsDir',
  'docsBranch',
  'editLinks',
  'editLinkText',
  'heroText',
  'sidebarDepth',
  'sidebar',
]

function linkOf(value) {
  if (value == null || value === '') return undefined
  let link = String(value).trim()
  if (/^https?:\/\//.test(link)) return link
  if (!link.startsWith('/')) link = `/${link}`
  link = link.replace(/\.md$/, '')
  if (link.length > 1 && link.endsWith('/')) link = link.slice(0, -1)
  return link
}

function titleFromSlug(link) {
  const base = String(link).split('/').filter(Boolean).pop()
  if (!base) return 'Home'
  return base.split(/[-_]/).filter(Boolean).map((word) => word.charAt(0).toUpperCase() + word.slice(1)).join(' ')
}

// String sidebar entries have no title of their own. VuePress used the page
// heading; fall back to the slug when the file is not there yet.
function pageTitle(root, link) {
  const rel = String(link || '').replace(/^\//, '').replace(/\.md$/, '')
  if (!rel) return 'Home'
  const candidates = [
    path.join(root, `${rel}.md`),
    path.join(root, rel, 'index.md'),
    path.join(root, rel, 'README.md'),
  ]
  for (const file of candidates) {
    if (!fs.existsSync(file)) continue
    const text = fs.readFileSync(file, 'utf8')
    const frontmatter = text.match(/^---\r?\n([\s\S]*?)\r?\n---/)
    if (frontmatter) {
      const titleLine = frontmatter[1].match(/^title:\s*(.+)\s*$/m)
      if (titleLine) return titleLine[1].replace(/^['"]|['"]$/g, '')
    }
    const body = frontmatter ? text.slice(frontmatter[0].length) : text
    const heading = body.match(/^#\s+(.+)$/m)
    if (heading) return heading[1].trim()
  }
  return titleFromSlug(link)
}

function translateItem(item, root) {
  if (typeof item === 'string') {
    const link = linkOf(item)
    return { text: pageTitle(root, link), link }
  }
  if (Array.isArray(item)) {
    const link = linkOf(item[0])
    return { text: item[1] || pageTitle(root, link), link }
  }
  const link = linkOf(item.link || item.path)
  const text = item.text || item.title || (link ? pageTitle(root, link) : undefined)
  const children = item.items || item.children
  const out = {}
  if (text) out.text = text
  if (link) out.link = link
  if (Array.isArray(children)) out.items = children.map((child) => translateItem(child, root))
  if (Object.prototype.hasOwnProperty.call(item, 'collapsed')) out.collapsed = item.collapsed
  else if (item.collapsable === false) out.collapsed = false
  return out
}

export function translateSidebar(sidebar, root) {
  if (!Array.isArray(sidebar)) return sidebar
  return sidebar.map((item) => translateItem(item, root))
}

function autoSidebar(root) {
  const skip = new Set(['README.md', 'index.md', 'main-metadata.md', 'package.json', 'package-lock.json', 'node_modules', 'public'])
  const entries = fs.readdirSync(root)
    .filter((name) => !name.startsWith('.') && !skip.has(name))
    .sort((a, b) => a.localeCompare(b))
  const items = [{ text: 'Home', link: '/' }]
  for (const name of entries) {
    const full = path.join(root, name)
    if (fs.statSync(full).isDirectory()) {
      const children = fs.readdirSync(full)
        .filter((sub) => sub.endsWith('.md'))
        .sort((a, b) => a.localeCompare(b))
        .map((sub) => {
          const link = `/${name}/${sub.replace(/\.md$/, '')}`
          return { text: pageTitle(root, link), link }
        })
      if (children.length) items.push({ text: name, items: children, collapsed: false })
    } else if (name.endsWith('.md')) {
      const link = `/${name.replace(/\.md$/, '')}`
      items.push({ text: pageTitle(root, link), link })
    }
  }
  return items
}

export function buildSiteConfig(extras, root) {
  const themeExtras = extras.themeConfig || {}
  const repo = themeExtras.docsRepo
  const branch = themeExtras.docsBranch || 'master'
  const docsDirName = themeExtras.docsDir || 'docs'
  const themeConfig = {
    siteTitle: false,
    logo: '/img/fairwinds-logo.svg',
    logoLink: {
      link: 'https://fairwinds.com',
      target: '_blank',
      rel: 'noopener noreferrer',
    },
    search: { provider: 'local' },
    editLink: {
      pattern: `https://github.com/${repo}/edit/${branch}/${docsDirName}/:path`,
      text: themeExtras.editLinkText || 'Help us improve this page',
    },
    nav: [
      { text: 'View on GitHub', link: `https://github.com/${repo}` },
    ],
    sidebar: themeExtras.sidebar && themeExtras.sidebar !== 'auto'
      ? translateSidebar(themeExtras.sidebar, root)
      : autoSidebar(root),
  }
  const restTheme = { ...themeExtras }
  for (const key of VUEPRESS_THEME_KEYS) delete restTheme[key]
  Object.assign(themeConfig, restTheme)

  const head = [
    ['link', { rel: 'icon', href: '/favicon.png' }],
    ['script', { src: '/scripts/marketing.js' }],
  ]
  if (Array.isArray(extras.head)) head.push(...extras.head)

  const skip = new Set(['title', 'description', 'head', 'themeConfig', 'plugins', 'chainWebpack', 'configureWebpack', 'evergreen', 'dest', 'port', 'theme', 'patterns'])
  const rest = {}
  for (const key of Object.keys(extras)) {
    if (!skip.has(key)) rest[key] = extras[key]
  }

  return {
    title: extras.title,
    description: extras.description,
    appearance: false,
    cleanUrls: true,
    ignoreDeadLinks: [/^https?:\/\/localhost/],
    outDir: '../dist',
    srcExclude: ['README.md', 'main-metadata.md'],
    head,
    // inferDescription() already ran. It reads frontmatter.description / head,
    // not the VuePress `meta` list, and it ignores later frontmatter edits.
    transformPageData(pageData) {
      const meta = pageData.frontmatter.meta
      if (!Array.isArray(meta) || pageData.description) return
      const desc = meta.find((item) => item && item.name === 'description')
      if (desc && desc.content) return { description: desc.content }
    },
    themeConfig,
    ...rest,
  }
}

function extrasPath() {
  const next = path.join(here, 'config-extras.js')
  if (fs.existsSync(next)) return next
  return path.join(docsDir, '.vuepress', 'config-extras.js')
}

function loadExtras(configPath) {
  if (!fs.existsSync(configPath)) {
    throw new Error('Please add config-extras.js to specify your project details')
  }
  const extras = require(configPath)
  if (!extras.title || !extras.description || !extras.themeConfig || !extras.themeConfig.docsRepo) {
    throw new Error("Please specify 'title', 'description', and 'themeConfig.docsRepo' in config-extras.js")
  }
  return extras
}

function selfCheck() {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), 'docs-sidebar-'))
  try {
    fs.mkdirSync(path.join(root, 'features'))
    fs.writeFileSync(path.join(root, 'features', 'policies.md'), '# Admission\n')
    fs.writeFileSync(path.join(root, 'cli.md'), '---\ntitle: CLI\n---\n# ignored\n')
    const sidebar = translateSidebar([
      ['/', 'Home'],
      {
        title: 'Polaris',
        path: '/',
        sidebarDepth: 0,
        collapsable: false,
        children: [{ title: 'Changelog', path: '/changelog' }],
      },
      { title: 'Guide', path: 'contributing/guide' },
      { text: 'CLI', link: '/cli' },
      { title: 'Features', children: ['features/policies'] },
      '/missing-page',
    ], root)
    assert.deepEqual(sidebar[0], { text: 'Home', link: '/' })
    assert.equal(sidebar[1].collapsed, false)
    assert.equal(sidebar[1].sidebarDepth, undefined)
    assert.deepEqual(sidebar[1].items[0], { text: 'Changelog', link: '/changelog' })
    assert.deepEqual(sidebar[2], { text: 'Guide', link: '/contributing/guide' })
    assert.deepEqual(sidebar[3], { text: 'CLI', link: '/cli' })
    assert.deepEqual(sidebar[4].items[0], { text: 'Admission', link: '/features/policies' })
    assert.deepEqual(sidebar[5], { text: 'Missing Page', link: '/missing-page' })

    const generated = autoSidebar(root)
    assert.deepEqual(generated.find((item) => item.link === '/cli'), { text: 'CLI', link: '/cli' })

    const cfg = buildSiteConfig({
      title: 'T',
      description: 'D',
      base: '/',
      head: [['link', { rel: 'icon', href: '/favicon.png' }]],
      themeConfig: {
        docsRepo: 'FairwindsOps/insights-docs',
        docsBranch: 'main',
        docsDir: 'docs',
        editLinks: true,
        heroText: '',
        logo: '/img/logo.svg',
        sidebar: [{ title: 'Home', path: '/' }],
      },
    }, root)
    assert.equal(cfg.themeConfig.editLink.pattern, 'https://github.com/FairwindsOps/insights-docs/edit/main/docs/:path')
    assert.equal(cfg.themeConfig.docsRepo, undefined)
    assert.equal(cfg.themeConfig.logo, '/img/logo.svg')
    assert.equal(cfg.base, '/')
    assert.deepEqual(cfg.themeConfig.sidebar, [{ text: 'Home', link: '/' }])
    assert.deepEqual(cfg.transformPageData({
      frontmatter: { meta: [{ name: 'description', content: 'Hello' }] },
      description: '',
    }), { description: 'Hello' })
    assert.equal(cfg.transformPageData({
      frontmatter: { meta: [{ name: 'description', content: 'Nope' }] },
      description: 'Already',
    }), undefined)
    console.log('vitepress config ok')
  } finally {
    fs.rmSync(root, { recursive: true, force: true })
  }
}

const isMain = process.argv[1] && path.resolve(process.argv[1]) === path.resolve(fileURLToPath(import.meta.url))
if (isMain && process.argv.includes('--self-check')) {
  selfCheck()
  process.exit(0)
}

export default defineConfig(buildSiteConfig(loadExtras(extrasPath()), docsDir))

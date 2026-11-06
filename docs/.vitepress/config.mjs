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

// package.json overrides vite to 6.4.5. VitePress 1.6.4 declares vite ^5, and
// that line is still inside the dev-server path-traversal range (vite <=6.4.3)
// with no 5.x release. 6.4.5 is the patched build. The published site is
// static; only `vitepress dev` runs that server.

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

// S3 serves page/index.html and redirects /page to /page/. Rewriting the
// source to page/index.md makes VitePress canonicalize the same URL.
function linkOf(value) {
  if (value == null || value === '') return undefined
  let link = String(value).trim()
  if (/^[a-z]+:/i.test(link) || link.startsWith('//')) return link
  const hashAt = link.indexOf('#')
  const hash = hashAt >= 0 ? link.slice(hashAt) : ''
  if (hashAt >= 0) link = link.slice(0, hashAt)
  const queryAt = link.indexOf('?')
  const query = queryAt >= 0 ? link.slice(queryAt) : ''
  if (queryAt >= 0) link = link.slice(0, queryAt)
  if (!link.startsWith('/')) link = `/${link}`
  link = link.replace(/\.md$/, '')
  if (link.length > 1 && link.endsWith('/')) link = link.slice(0, -1)
  if (link !== '/') link += '/'
  return link + query + hash
}

// `from` is the source page (`customization/configuration.md`). VitePress
// rewrites that file to `customization/configuration/index.md` and then
// resolves relative links from the rewritten directory, so `checks.md`
// would look for `customization/configuration/checks.md`. Absolutize from
// the source file first. The published URL is a directory, so page links
// keep the trailing slash S3 redirects to.
export function slashPageLink(url, from) {
  if (!url || url.startsWith('#') || /^[a-z]+:/i.test(url) || url.startsWith('//')) return url
  const hashAt = url.indexOf('#')
  const hash = hashAt >= 0 ? url.slice(hashAt) : ''
  let rest = hashAt >= 0 ? url.slice(0, hashAt) : url
  const queryAt = rest.indexOf('?')
  const query = queryAt >= 0 ? rest.slice(queryAt) : ''
  if (queryAt >= 0) rest = rest.slice(0, queryAt)
  if (from && rest && !rest.startsWith('/')) {
    const dir = path.posix.dirname(String(from).replace(/\\/g, '/'))
    const base = dir === '.' ? '/' : `/${dir.replace(/^\//, '')}`
    rest = path.posix.normalize(path.posix.join(base, rest))
    if (!rest.startsWith('/')) rest = `/${rest}`
  }
  if (rest === '/' || rest === '') return url
  rest = rest.replace(/\.(md|html)$/i, '')
  if (rest === '/' || rest === '') return `/${query}${hash}`
  const last = rest.split('/').pop()
  if (!last || last.includes('.')) return `${rest}${query}${hash}`
  if (rest.endsWith('/')) return `${rest}${query}${hash}`
  if (!rest.startsWith('/') && !rest.startsWith('.')) return url
  return `${rest}/${query}${hash}`
}

function sourceRel(env, root) {
  const real = env?.realPath || env?.path
  if (!real) return undefined
  const rel = path.isAbsolute(real) ? path.relative(root, real) : real
  return rel.split(path.sep).join('/')
}

function slashInternalLinks(md, root) {
  const render = md.renderer.rules.link_open
  if (!render) return
  md.renderer.rules.link_open = (tokens, idx, options, env, self) => {
    const token = tokens[idx]
    const hrefAt = token.attrIndex('href')
    if (hrefAt >= 0) {
      token.attrs[hrefAt][1] = slashPageLink(token.attrs[hrefAt][1], sourceRel(env, root))
    }
    return render(tokens, idx, options, env, self)
  }
}

function titleFromSlug(link) {
  const base = String(link).split('/').filter(Boolean).pop()
  if (!base) return 'Home'
  return base.split(/[-_]/).filter(Boolean).map((word) => word.charAt(0).toUpperCase() + word.slice(1)).join(' ')
}

function splitMarkdown(text) {
  const frontmatter = text.match(/^---\r?\n([\s\S]*?)\r?\n---/)
  const body = frontmatter ? text.slice(frontmatter[0].length) : text
  return {
    frontmatter: frontmatter ? frontmatter[1] : '',
    // Fenced blocks contain shell comments (`# flag`). Those are not headings.
    body: body.replace(/```[\s\S]*?```/g, ''),
  }
}

function frontmatterTitle(yaml) {
  const titleLine = yaml.match(/^title:\s*(.+)\s*$/m)
  if (!titleLine) return undefined
  return titleLine[1].replace(/^['"]|['"]$/g, '')
}

// VuePress used the first heading of any level. VitePress only promotes h1.
function firstHeading(body) {
  const heading = body.match(/^#{1,6}[ \t]+(.+?)[ \t]*$/m)
  return heading ? heading[1].trim() : undefined
}

// Document title only when the page opens with the heading. A later heading
// (the homepage's registry notice, for example) is not the page title.
function leadingHeading(body) {
  const heading = body.match(/^\s*#{1,6}[ \t]+([^\n]+?)\s*(?:\n|$)/)
  return heading ? heading[1].trim() : undefined
}

function readMarkdown(root, rel) {
  const candidates = [
    path.join(root, `${rel}.md`),
    path.join(root, rel, 'index.md'),
    path.join(root, rel, 'README.md'),
    path.join(root, rel),
  ]
  for (const file of candidates) {
    if (fs.existsSync(file) && fs.statSync(file).isFile()) return splitMarkdown(fs.readFileSync(file, 'utf8'))
  }
  return undefined
}

// String sidebar entries have no title of their own. VuePress used the page
// heading; fall back to the slug when the file is not there yet.
function pageTitle(root, link) {
  const rel = String(link || '').replace(/^\//, '').replace(/\.md$/, '').replace(/\/$/, '')
  if (!rel) return 'Home'
  const page = readMarkdown(root, rel)
  if (!page) return titleFromSlug(link)
  return frontmatterTitle(page.frontmatter) || firstHeading(page.body) || titleFromSlug(link)
}

// VuePress sidebarDepth 0 hid headings. 1 showed h2, 2 showed h2 and h3.
function outlineForDepth(depth) {
  if (depth <= 0) return false
  if (depth === 1) return 2
  return [2, depth + 1]
}

// Group sidebarDepth applies to that group's own page and its direct page
// children (VuePress passes the group's depth into SidebarLinks). A nested
// group uses its own depth, then the theme depth. Theme depth 1 is the
// VitePress outline default, so it is not recorded.
function outlineMap(sidebar, themeDepth) {
  const map = new Map()
  const themeFallback = typeof themeDepth === 'number' && themeDepth !== 1 ? themeDepth : undefined
  function remember(link, depth) {
    if (!link || typeof depth !== 'number') return
    const key = linkOf(link)
    if (key) map.set(key, outlineForDepth(depth))
  }
  function walk(items, inherited) {
    if (Array.isArray(items)) {
      for (const item of items) visit(item, inherited)
      return
    }
    if (items && typeof items === 'object') {
      for (const value of Object.values(items)) walk(value, inherited)
    }
  }
  function visit(item, inherited) {
    if (typeof item === 'string') {
      remember(item, inherited)
      return
    }
    if (Array.isArray(item)) {
      remember(item[0], inherited)
      return
    }
    if (!item || typeof item !== 'object') return
    const own = typeof item.sidebarDepth === 'number' ? item.sidebarDepth : undefined
    const depth = own !== undefined ? own : inherited
    const link = item.link || item.path
    if (link) remember(link, depth)
    const children = item.items || item.children
    if (Array.isArray(children)) walk(children, own !== undefined ? own : themeFallback)
  }
  walk(sidebar, themeFallback)
  return map
}

function pageLink(rel) {
  if (!rel) return undefined
  let link = String(rel).split(path.sep).join('/')
  link = link.replace(/\.md$/i, '')
  if (link.endsWith('/index')) link = link.slice(0, -'/index'.length)
  if (link === 'index' || link === 'README' || link === '') return '/'
  return linkOf(link)
}

function withLogoAlt(logo) {
  if (typeof logo === 'string') return { src: logo, alt: 'Fairwinds' }
  if (logo && typeof logo === 'object' && logo.src && !logo.alt) return { ...logo, alt: 'Fairwinds' }
  return logo
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
  // VuePress groups are collapsible and start open unless collapsable is false.
  // VitePress omits the chevron when collapsed is unset, and collapsed: false
  // means collapsible but expanded. A VitePress group that already uses `items`
  // keeps that default unless it sets collapsed itself.
  if (Object.prototype.hasOwnProperty.call(item, 'collapsed')) out.collapsed = item.collapsed
  else if (Array.isArray(children) && vuePressGroup(item) && item.collapsable !== false) out.collapsed = false
  return out
}

function vuePressGroup(item) {
  return Array.isArray(item.children) || Object.prototype.hasOwnProperty.call(item, 'collapsable')
}

export function translateSidebar(sidebar, root) {
  if (Array.isArray(sidebar)) return sidebar.map((item) => translateItem(item, root))
  if (sidebar && typeof sidebar === 'object') {
    const out = {}
    for (const [key, value] of Object.entries(sidebar)) out[key] = translateSidebar(value, root)
    return out
  }
  return sidebar
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
          const link = linkOf(`/${name}/${sub.replace(/\.md$/, '')}`)
          return { text: pageTitle(root, link), link }
        })
      if (children.length) items.push({ text: name, items: children, collapsed: false })
    } else if (name.endsWith('.md')) {
      const link = linkOf(`/${name.replace(/\.md$/, '')}`)
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
      // VitePress serializes this function and runs it later, so the repo
      // path has to be written into the function body. index.md is generated
      // from the repository README and is not committed.
      pattern: new Function('page', [
        'const file = page.filePath === "index.md" ? "README.md" : ' + JSON.stringify(`${docsDirName}/`) + ' + page.filePath',
        'return ' + JSON.stringify(`https://github.com/${repo}/edit/${branch}/`) + ' + file',
      ].join('\n')),
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
  themeConfig.logo = withLogoAlt(themeConfig.logo)
  const outlines = outlineMap(themeExtras.sidebar, themeExtras.sidebarDepth)

  const head = [
    ['link', { rel: 'icon', href: '/favicon.png' }],
    ['script', { src: '/scripts/marketing.js' }],
  ]
  if (Array.isArray(extras.head)) head.push(...extras.head)

  assertNoRewriteClash(root)
  const skip = new Set(['title', 'description', 'head', 'themeConfig', 'plugins', 'chainWebpack', 'configureWebpack', 'evergreen', 'dest', 'port', 'theme', 'patterns'])
  const rest = {}
  for (const key of Object.keys(extras)) {
    if (!skip.has(key)) rest[key] = extras[key]
  }

  const userMarkdown = rest.markdown
  delete rest.markdown

  return {
    title: extras.title,
    description: extras.description,
    appearance: false,
    cleanUrls: true,
    ignoreDeadLinks: [/^https?:\/\/localhost/],
    outDir: '../dist',
    srcExclude: ['README.md', 'main-metadata.md'],
    head,
    themeConfig,
    ...rest,
    rewrites: s3Rewrite,
    // VitePress 1.6 runs @mdit-vue/plugin-headers only when markdown.headers
    // is set. The outline (VuePress sidebarDepth) stays empty without it.
    markdown: {
      ...userMarkdown,
      headers: userMarkdown?.headers ?? true,
      config(md) {
        userMarkdown?.config?.(md)
        slashInternalLinks(md, root)
      },
    },
    // inferDescription() and inferTitle() already ran. inferDescription reads
    // frontmatter.description / head, not the VuePress `meta` list. inferTitle
    // reads frontmatter.title or the h1, and ignores a lower heading.
    transformPageData(pageData) {
      const frontmatter = pageData.frontmatter || {}
      const patch = {}
      const meta = frontmatter.meta
      if (Array.isArray(meta)) {
        if (!pageData.description) {
          const desc = meta.find((item) => item && item.name === 'description' && item.content)
          if (desc) patch.description = desc.content
        }
        // metadata yaml titles are the document title ("Goldilocks Documentation | Fairwinds").
        const titleMeta = meta.find((item) => item && item.name === 'title' && item.content)
        if (titleMeta && titleMeta.content !== pageData.title) {
          patch.title = titleMeta.content
          // The metadata title is already the full document title.
          patch.titleTemplate = false
        }
      }
      if (!pageData.title && !patch.title) {
        // filePath is the source file. relativePath is the rewritten
        // `page/index.md`, which is not on disk.
        const rel = pageData.filePath || pageData.relativePath
        const page = rel && readMarkdown(root, rel)
        const title = page && leadingHeading(page.body)
        if (title) patch.title = title
      }
      if (frontmatter.outline === undefined) {
        const outline = typeof frontmatter.sidebarDepth === 'number'
          ? outlineForDepth(frontmatter.sidebarDepth)
          : outlines.get(pageLink(pageData.filePath || pageData.relativePath))
        if (outline !== undefined) patch.frontmatter = { ...frontmatter, outline }
      }
      return Object.keys(patch).length ? patch : undefined
    },
  }
}

const SKIP_PAGE_DIRS = new Set(['node_modules', 'public'])
const SKIP_PAGES = new Set(['README.md', 'main-metadata.md'])

export function s3Rewrite(id) {
  if (!id.endsWith('.md') || id === 'index.md' || id === '404.md' || id.endsWith('/index.md')) return id
  return `${id.slice(0, -'.md'.length)}/index.md`
}

function assertNoRewriteClash(root) {
  if (!fs.existsSync(root)) return
  const dests = new Map()
  function walk(dir, rel) {
    for (const name of fs.readdirSync(dir)) {
      if (name.startsWith('.') || SKIP_PAGE_DIRS.has(name)) continue
      const full = path.join(dir, name)
      const id = rel ? `${rel}/${name}` : name
      if (fs.statSync(full).isDirectory()) walk(full, id)
      else if (name.endsWith('.md') && !SKIP_PAGES.has(name)) {
        const dest = s3Rewrite(id)
        const prev = dests.get(dest)
        if (prev) throw new Error(`s3 rewrite: ${prev} and ${id} both publish to ${dest}`)
        dests.set(dest, id)
      }
    }
  }
  walk(root, '')
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
    fs.writeFileSync(path.join(root, 'flags.md'), [
      '#### CLI Options',
      '',
      '```',
      '# top-level commands',
      '```',
      '',
    ].join('\n'))
    fs.writeFileSync(path.join(root, 'home.md'), '<div>logo</div>\n\n## Notice\n')
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
      '/flags',
      { text: 'Native', items: [{ text: 'Kept', link: '/cli' }] },
    ], root)
    assert.deepEqual(sidebar[0], { text: 'Home', link: '/' })
    assert.equal(sidebar[1].collapsed, undefined)
    assert.equal(sidebar[1].sidebarDepth, undefined)
    assert.deepEqual(sidebar[1].items[0], { text: 'Changelog', link: '/changelog/' })
    assert.deepEqual(sidebar[2], { text: 'Guide', link: '/contributing/guide/' })
    assert.deepEqual(sidebar[3], { text: 'CLI', link: '/cli/' })
    assert.equal(sidebar[4].collapsed, false)
    assert.deepEqual(sidebar[4].items[0], { text: 'Admission', link: '/features/policies/' })
    assert.deepEqual(sidebar[5], { text: 'Missing Page', link: '/missing-page/' })
    assert.deepEqual(sidebar[6], { text: 'CLI Options', link: '/flags/' })
    assert.equal(sidebar[7].collapsed, undefined)
    assert.deepEqual(sidebar[7].items[0], { text: 'Kept', link: '/cli/' })
    assert.deepEqual(translateSidebar({
      '/': [{ title: 'Contributing', children: [{ title: 'Guide', path: 'contributing/guide' }] }],
    }, root)['/'][0], {
      text: 'Contributing',
      collapsed: false,
      items: [{ text: 'Guide', link: '/contributing/guide/' }],
    })

    const generated = autoSidebar(root)
    assert.deepEqual(generated.find((item) => item.link === '/cli/'), { text: 'CLI', link: '/cli/' })
    assert.equal(s3Rewrite('Foo/feature.md'), 'Foo/feature/index.md')
    assert.equal(s3Rewrite('index.md'), 'index.md')
    assert.equal(s3Rewrite('checks/index.md'), 'checks/index.md')
    assert.equal(slashPageLink('/Foo/feature'), '/Foo/feature/')
    assert.equal(slashPageLink('/Foo/feature/'), '/Foo/feature/')
    assert.equal(slashPageLink('/Foo/feature#hash'), '/Foo/feature/#hash')
    assert.equal(slashPageLink('/img/logo.svg'), '/img/logo.svg')
    assert.equal(slashPageLink('https://fairwinds.com'), 'https://fairwinds.com')
    assert.equal(slashPageLink('checks.md', 'customization/configuration.md'), '/customization/checks/')
    assert.equal(slashPageLink('./feature', 'Foo/linktest.md'), '/Foo/feature/')
    assert.equal(slashPageLink('./feature.md#feature-1', 'Foo/linktest.md'), '/Foo/feature/#feature-1')
    assert.equal(slashPageLink('../cli', 'customization/configuration.md'), '/cli/')
    assert.equal(slashPageLink('code-of-conduct.md', 'contributing.md'), '/code-of-conduct/')
    assert.equal(slashPageLink('/Quux/thing', 'Foo/linktest.md'), '/Quux/thing/')
    assert.equal(slashPageLink('./diagram.png', 'Foo/linktest.md'), '/Foo/diagram.png')
    assert.equal(slashPageLink('checks.md?plain=1', 'customization/configuration.md'), '/customization/checks/?plain=1')
    const clash = fs.mkdtempSync(path.join(os.tmpdir(), 'docs-clash-'))
    fs.mkdirSync(path.join(clash, 'checks'))
    fs.writeFileSync(path.join(clash, 'checks.md'), '# parent\n')
    fs.writeFileSync(path.join(clash, 'checks', 'index.md'), '# child\n')
    assert.throws(() => assertNoRewriteClash(clash), /both publish/)
    fs.rmSync(clash, { recursive: true, force: true })

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
    assert.equal(cfg.markdown.headers, true)
    assert.equal(cfg.themeConfig.editLink.pattern({ filePath: 'cli.md' }), 'https://github.com/FairwindsOps/insights-docs/edit/main/docs/cli.md')
    assert.equal(cfg.themeConfig.editLink.pattern({ filePath: 'index.md' }), 'https://github.com/FairwindsOps/insights-docs/edit/main/README.md')
    // VitePress round-trips the function through toString().
    const revived = new Function(`return ${cfg.themeConfig.editLink.pattern.toString()}`)()
    assert.equal(revived({ filePath: 'index.md' }), 'https://github.com/FairwindsOps/insights-docs/edit/main/README.md')
    assert.equal(cfg.themeConfig.docsRepo, undefined)
    assert.deepEqual(cfg.themeConfig.logo, { src: '/img/logo.svg', alt: 'Fairwinds' })
    assert.equal(cfg.base, '/')
    assert.deepEqual(cfg.themeConfig.sidebar, [{ text: 'Home', link: '/' }])
    assert.deepEqual(cfg.transformPageData({
      frontmatter: { meta: [{ name: 'description', content: 'Hello' }] },
      description: '',
    }), { description: 'Hello' })
    assert.equal(cfg.transformPageData({
      frontmatter: { meta: [{ name: 'description', content: 'Nope' }] },
      description: 'Already',
      title: 'Kept',
    }), undefined)
    assert.deepEqual(cfg.transformPageData({
      frontmatter: { meta: [
        { name: 'title', content: 'Goldilocks Documentation | Fairwinds' },
        { name: 'description', content: 'Hello' },
      ] },
      description: '',
      title: 'Goldilocks',
    }), {
      description: 'Hello',
      title: 'Goldilocks Documentation | Fairwinds',
      titleTemplate: false,
    })
    assert.deepEqual(cfg.transformPageData({
      frontmatter: { sidebarDepth: 0, meta: [{ name: 'description', content: 'Hello' }] },
      description: '',
      title: '',
      relativePath: 'flags/index.md',
      filePath: 'flags.md',
    }), {
      description: 'Hello',
      title: 'CLI Options',
      frontmatter: {
        sidebarDepth: 0,
        meta: [{ name: 'description', content: 'Hello' }],
        outline: false,
      },
    })
    assert.deepEqual(cfg.transformPageData({
      frontmatter: { sidebarDepth: 2 },
      description: 'Already',
      title: 'Kept',
    }).frontmatter.outline, [2, 3])
    assert.equal(cfg.transformPageData({
      frontmatter: {},
      description: 'Already',
      title: '',
      relativePath: 'home.md',
    }), undefined)
    const depthCfg = buildSiteConfig({
      title: 'T',
      description: 'D',
      themeConfig: {
        docsRepo: 'FairwindsOps/polaris',
        sidebar: [
          {
            title: 'Polaris',
            path: '/',
            sidebarDepth: 0,
            collapsable: false,
            children: [
              { title: 'Changelog', path: '/changelog' },
              { title: 'Nested', children: [{ title: 'Guide', path: '/contributing/guide' }] },
            ],
          },
          { title: 'Checks', sidebarDepth: 0, children: ['/checks/security'] },
          { title: 'Usage', children: ['/cli'] },
        ],
      },
    }, root)
    assert.deepEqual(depthCfg.transformPageData({
      frontmatter: {},
      title: 'Home',
      description: 'd',
      filePath: 'index.md',
    }), { frontmatter: { outline: false } })
    assert.deepEqual(depthCfg.transformPageData({
      frontmatter: {},
      title: 'Changelog',
      description: 'd',
      relativePath: 'changelog/index.md',
      filePath: 'changelog.md',
    }), { frontmatter: { outline: false } })
    assert.deepEqual(depthCfg.transformPageData({
      frontmatter: {},
      title: 'Security',
      description: 'd',
      relativePath: 'checks/security/index.md',
      filePath: 'checks/security.md',
    }), { frontmatter: { outline: false } })
    assert.equal(depthCfg.transformPageData({
      frontmatter: {},
      title: 'Guide',
      description: 'd',
      filePath: 'contributing/guide.md',
    }), undefined)
    assert.equal(depthCfg.transformPageData({
      frontmatter: {},
      title: 'CLI',
      description: 'd',
      filePath: 'cli.md',
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

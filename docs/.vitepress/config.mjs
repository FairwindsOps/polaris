import { defineConfig } from 'vitepress'

// Sidebar matches the old VuePress config-extras. Page titles come from the markdown.
// package.json overrides vite to 6.4.5. VitePress 1.6.4 declares vite ^5, and Vite 5
// does not get the 2026 dev-server fixes that shipped in 6.4.5 (GHSA-8g5g-mxqh-grgj).
// The published site is static; only `vitepress dev` runs that server.
export default defineConfig({
  title: 'Fairwinds Polaris Documentation',
  description: 'Documentation for Fairwinds Polaris - audit and enforce Kubernetes best practices for your workloads',
  appearance: false,
  cleanUrls: true,
  ignoreDeadLinks: [/^https?:\/\/localhost/],
  outDir: '../dist',
  srcExclude: ['README.md', 'main-metadata.md'],
  head: [
    ['link', { rel: 'icon', href: '/favicon.png' }],
    ['script', { src: '/scripts/marketing.js' }],
  ],
  transformPageData(pageData) {
    // inferDescription() already ran. It reads frontmatter.description / head,
    // not the VuePress `meta` list, and it ignores later frontmatter edits.
    const meta = pageData.frontmatter.meta
    if (!Array.isArray(meta) || pageData.description) return
    const desc = meta.find((item) => item && item.name === 'description')
    if (desc && desc.content) return { description: desc.content }
  },
  themeConfig: {
    siteTitle: false,
    logo: '/img/fairwinds-logo.svg',
    logoLink: {
      link: 'https://fairwinds.com',
      target: '_blank',
      rel: 'noopener noreferrer',
    },
    editLink: {
      pattern: 'https://github.com/FairwindsOps/polaris/edit/master/docs/:path',
      text: 'Help us improve this page',
    },
    search: { provider: 'local' },
    nav: [
      { text: 'View on GitHub', link: 'https://github.com/FairwindsOps/polaris' },
    ],
    sidebar: [
      {
        text: 'Polaris',
        link: '/',
        collapsed: false,
        items: [
          { text: 'Changelog', link: '/changelog' },
          { text: 'Code of Conduct', link: '/code-of-conduct' },
          { text: 'Contributing', link: '/contributing' },
        ],
      },
      {
        text: 'Ways to Run Polaris',
        collapsed: false,
        items: [
          { text: 'Dashboard', link: '/dashboard' },
          { text: 'Admission Controller', link: '/admission-controller' },
          { text: 'Infrastructure as Code', link: '/infrastructure-as-code' },
        ],
      },
      {
        text: 'Usage',
        collapsed: false,
        items: [
          { text: 'CLI', link: '/cli' },
        ],
      },
      {
        text: 'Customization',
        collapsed: false,
        items: [
          { text: 'Configuration', link: '/customization/configuration' },
          { text: 'Check Settings', link: '/customization/checks' },
          { text: 'Custom Checks', link: '/customization/custom-checks' },
          { text: 'Exemptions', link: '/customization/exemptions' },
        ],
      },
      {
        text: 'Checks',
        collapsed: false,
        items: [
          { text: 'Security', link: '/checks/security' },
          { text: 'Efficiency', link: '/checks/efficiency' },
          { text: 'Reliability', link: '/checks/reliability' },
        ],
      },
    ],
  },
})

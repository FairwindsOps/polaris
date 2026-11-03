import { defineConfig } from 'vitepress'

// Sidebar matches the old VuePress config-extras. Page titles come from the markdown.
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
    ['script', { src: '/scripts/modify.js' }],
    ['script', { src: '/scripts/marketing.js' }],
  ],
  transformPageData(pageData) {
    const meta = pageData.frontmatter.meta
    if (!Array.isArray(meta) || pageData.frontmatter.description) return
    const desc = meta.find((item) => item && item.name === 'description')
    if (desc && desc.content) pageData.frontmatter.description = desc.content
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

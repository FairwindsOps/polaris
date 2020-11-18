module.exports = {
  title: "Fairwinds Polaris Documentation",
  description: "Documentation for Fairwinds Polaris - audit and enforce Kubernetes best practices for your workloads",
  themeConfig: {
    docsRepo: "FairwindsOps/polaris",
    sidebar: [
      {
        title: "Polaris",
        path: "/",
        sidebarDepth: 0,
        collapsable: false,
        children: [
          {
            title: "Changelog",
            path: "/changelog",
          },
          {
            title: "Code of Conduct",
            path: "/code-of-conduct",
          },
          {
            title: "Contributing",
            path: "/contributing",
          },
        ],
      },
      {
        title: "Components",
        collapsable: false,
        children: [
          "/dashboard",
          "/admission-controller",
          "/infrastructure-as-code",
        ],
      },
      {
        title: "Configuration",
        collapsable: false,
        children: [
          "/configuration/check-configuration",
          "/configuration/custom-checks",
          "/configuration/exemptions",
        ]
      }
    ]
  }
}
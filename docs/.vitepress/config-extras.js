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
        title: "Ways to Run Polaris",
        collapsable: false,
        children: [
          "/dashboard",
          "/admission-controller",
          "/infrastructure-as-code",
        ],
      },
      {
        title: "Usage",
        collapsable: false,
        children: [
          "/cli",
        ],
      },
      {
        title: "Customization",
        collapsable: false,
        children: [
          "/customization/configuration",
          "/customization/checks",
          "/customization/custom-checks",
          "/customization/exemptions",
        ]
      },
      {
        title: "Checks",
        collapsable: false,
        sidebarDepth: 0,
        children: [
          "/checks/security",
          "/checks/efficiency",
          "/checks/reliability",
        ],
      },
    ]
  }
}

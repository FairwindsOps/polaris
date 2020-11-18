module.exports = {
  title: "Fairwinds Polaris Documentation",
  description: "Documentation for Fairwinds Polaris - audit and enforce Kubernetes best practices for your workloads",
  themeConfig: {
    docsRepo: "FairwindsOps/polaris",
    sidebar: [
      ["/", "Polaris"],
      {
        title: "Changelog",
        sidebarDepth: 0,
        path: "/changelog",
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

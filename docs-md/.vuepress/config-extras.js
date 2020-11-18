module.exports = {
  title: "Fairwinds Polaris Documentation",
  description: "Documentation for Fairwinds Polaris - audit and enforce Kubernetes best practices for your workloads",
  themeConfig: {
    sidebar: [
      ["/", "Home"],
      {
        title: "Usage",
        collapsable: false,
        children: [
          "/dashboard/dashboard",
          "/admission-controller/admission-controller",
          "/iac/iac",
        ],
      }
    ]
  }
}

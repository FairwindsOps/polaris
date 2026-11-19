// DO NOT EDIT - Managed by Terraform
// This file is generated from FairwindsOps/documentation-template
import DefaultTheme from 'vitepress/theme'
import { withBase } from 'vitepress'
import Layout from './Layout.vue'
import './custom.css'

export default {
  extends: DefaultTheme,
  Layout,
  enhanceApp({ app }) {
    // Markdown still uses the VuePress image helper. Without it the page
    // render throws and the document body stays empty.
    app.config.globalProperties.$withBase = withBase
  },
}

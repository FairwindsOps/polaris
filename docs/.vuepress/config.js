// This file is generated from FairwindsOps/documentation-template
// DO NOT EDIT MANUALLY

const fs = require('fs');
const npath = require('path');

const CONFIG_FILE = npath.join(__dirname, 'config-extras.js');
const BASE_DIR = npath.join(__dirname, '..');

const extras = require(CONFIG_FILE);
if (!extras.title || !extras.description || !extras.themeConfig.docsRepo) {
  throw new Error("Please specify 'title', 'description', and 'themeConfig.docsRepo' in config-extras.js");
}

const docFiles = fs.readdirSync(BASE_DIR)
  .filter(f => f !== "README.md")
  .filter(f => f !== ".vuepress")
  .filter(f => f !== "node_modules")
  .filter(f => npath.extname(f) === '.md' || npath.extname(f) === '');

const sidebar = [['/', 'Home']].concat(docFiles.map(f => {
  const ext = npath.extname(f);
  if (ext === '') {
    // this is a directory
    const title = f;
    const children = fs.readdirSync(npath.join(BASE_DIR, f)).map(subf => {
      return '/' + f + '/' + npath.basename(subf);
    });
    return {title, children};
  }
  const path = npath.basename(f);
  return path;
}));

const baseConfig = {
  title: "",
  description: "",
  head: [
    ['link', { rel: 'icon', href: '/favicon.png' }],
    ['script', { src: '/scripts/modify.js' }],
    ['script', { src: '/scripts/marketing.js' }],
  ],
  themeConfig: {
    docsRepo: "",
    docsDir: 'docs-md',
    editLinks: true,
    editLinkText: "Help us improve this page",
    logo: '/img/fairwinds-logo.svg',
    heroText: "",
    sidebar,
    nav: [
      {text: 'View on GitHub', link: 'https://github.com/' + extras.themeConfig.docsRepo},
    ],
  },
  plugins: {
    'vuepress-plugin-clean-urls': {
      normalSuffix: '/',
      notFoundPath: '/404.html',
    },
    'check-md': {},
  },
}

let config = JSON.parse(JSON.stringify(baseConfig))
if (!fs.existsSync(CONFIG_FILE)) {
  throw new Error("Please add config-extras.js to specify your project details");
}
for (let key in extras) {
  if (!config[key]) config[key] = extras[key];
  else if (key === 'head') config[key] = config[key].concat(extras[key]);
  else Object.assign(config[key], extras[key]);
}
module.exports = config;

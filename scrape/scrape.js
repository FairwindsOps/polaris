const puppeteer = require('puppeteer');
const OUT_FILE = process.argv[2];
const HOST = process.argv[3] || 'http://localhost:8080';
const URL = `${HOST}?expand=true`;
(async () => {
  const browser = await puppeteer.launch({ headless: true });
  const page = await browser.newPage();
  await page.goto(URL, {waitUntil: 'networkidle0'});
  await page.waitFor(1000);
  const height = await page.evaluate(() => document.documentElement.offsetHeight);
  const pdf = await page.pdf({ path: OUT_FILE, width: '8.27in', height: height + 'px' });
  await browser.close();
  return pdf
})();

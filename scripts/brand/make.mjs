#!/usr/bin/env node
// Draws the favicon set and the 1200×630 share image from internal/sprites/sprites.json,
// and (with --shots) takes the README screenshots from a running copy of the game.
//
//   npm i --no-save puppeteer-core@23.11.1
//   node scripts/brand/make.mjs                                   # icons + share image
//   node scripts/brand/make.mjs --shots http://localhost:8080/    # README screenshots
//
// Needs Google Chrome at CHROME_PATH (default: the macOS app path).
// Outputs (committed): web/icon.svg, web/favicon.ico, web/favicon-32.png,
// web/apple-touch-icon.png, web/og.png, docs/desktop.png, docs/phone.png

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import puppeteer from 'puppeteer-core';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
const web = path.join(root, 'web');
const chrome = process.env.CHROME_PATH || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome';
const SITE = 'incident-commander-game.vercel.app';
const S = JSON.parse(fs.readFileSync(path.join(root, 'internal/sprites/sprites.json'), 'utf8'));
const heroPal = (beacon) => ({ ...S.palette.hero, ...S.palette.beacon[beacon] });

// A sprite as crisp SVG rects (one per pixel), sized to `size` and placed at (x, y).
function spriteSvg(rows, pal, x = 0, y = 0, size = rows.length) {
  const px = size / rows.length;
  let out = '';
  rows.forEach((row, j) => [...row].forEach((ch, i) => {
    if (pal[ch]) out += `<rect x="${x + i * px}" y="${y + j * px}" width="${px}" height="${px}" fill="${pal[ch]}"/>`;
  }));
  return out;
}

function iconSvg(bg) {
  const n = S.hero.length;
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${n} ${n}" shape-rendering="crispEdges">` +
    (bg ? `<rect width="${n}" height="${n}" fill="${bg}"/>` : '') + spriteSvg(S.hero, heroPal('alert')) + '</svg>';
}

function ico(pngs) {
  const header = Buffer.alloc(6 + 16 * pngs.length);
  header.writeUInt16LE(0, 0); header.writeUInt16LE(1, 2); header.writeUInt16LE(pngs.length, 4);
  let offset = header.length;
  pngs.forEach(({ size, buf }, i) => {
    const o = 6 + 16 * i;
    header.writeUInt8(size, o); header.writeUInt8(size, o + 1);
    header.writeUInt16LE(1, o + 4); header.writeUInt16LE(32, o + 6);
    header.writeUInt32LE(buf.length, o + 8); header.writeUInt32LE(offset, o + 12);
    offset += buf.length;
  });
  return Buffer.concat([header, ...pngs.map((p) => p.buf)]);
}

function ogHtml() {
  const font = fs.readFileSync(path.join(web, 'fonts/press-start-2p.woff2')).toString('base64');
  // A strip of board: resolved checks leading to the hero, alerts ahead.
  const cell = 56, tiles = [];
  for (let i = 0; i < 5; i++) {
    tiles.push(`<rect x="${i * cell + 5}" y="5" width="${cell - 10}" height="${cell - 10}" fill="#0f6b4e"/><rect x="${i * cell + 5}" y="5" width="${cell - 10}" height="${cell - 15}" fill="#1aa77a"/>` +
      spriteSvg(S.check, S.palette.check, i * cell + 10, 9, cell * 0.64));
  }
  const alerts = [0, 2].map((k) => spriteSvg(S.alert, S.palette.alert, (7 + k) * cell + 3, 2, cell * 0.94)).join('');
  return `<!doctype html><html><head><meta charset="utf-8"><style>
  @font-face{font-family:P;src:url(data:font/woff2;base64,${font})}
  *{margin:0;box-sizing:border-box}
  body{width:1200px;height:630px;overflow:hidden;font-family:P;color:#fff;
    background:radial-gradient(ellipse 70% 60% at 30% 0,#2a1840,#0e0a16 70%)}
  .hero{position:absolute;left:70px;top:70px;width:360px;height:360px;filter:drop-shadow(0 0 30px #ff4d4d55)}
  h1{position:absolute;left:480px;top:120px;font-weight:400;font-size:52px;line-height:1.35;color:#ffe45c;text-shadow:5px 5px 0 #b5179e}
  p{position:absolute;left:484px;top:300px;font-size:20px;line-height:1.9;color:#cbbfe0}
  .strip{position:absolute;left:70px;top:480px}
  .url{position:absolute;right:64px;bottom:44px;font-size:16px;color:#4ff0ff}
  .go{position:absolute;left:72px;bottom:44px;font-size:16px;color:#9b8fb3}
  </style></head><body>
  <svg class="hero" viewBox="0 0 16 16" shape-rendering="crispEdges">${spriteSvg(S.hero, heroPal('alert'))}</svg>
  <h1>INCIDENT<br>COMMANDER</h1>
  <p>RESOLVE EVERY ALERT.<br>DON'T HIT YOUR TRAIL.</p>
  <svg class="strip" width="${10 * cell}" height="${cell}" shape-rendering="crispEdges">${tiles.join('')}
    ${spriteSvg(S.hero, heroPal('alert'), 5 * cell - 6, -14, cell * 1.3)}${alerts}</svg>
  <div class="go">GO + WEBASSEMBLY</div>
  <div class="url">${SITE}</div>
  </body></html>`;
}

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const browser = await puppeteer.launch({ executablePath: chrome, headless: true });
try {
  const shotsAt = process.argv.indexOf('--shots');
  if (shotsAt !== -1) {
    const url = process.argv[shotsAt + 1] || 'http://localhost:8080/';
    fs.mkdirSync(path.join(root, 'docs'), { recursive: true });
    for (const [name, w, h, mobile] of [['desktop', 1100, 820, false], ['phone', 390, 844, true]]) {
      const page = await browser.newPage();
      await page.setViewport({ width: w, height: h, deviceScaleFactor: 2, isMobile: mobile, hasTouch: mobile });
      await page.goto(url, { waitUntil: 'networkidle0' });
      await sleep(600);
      // Make a short run so the board shows a trail.
      for (const k of ['ArrowUp', 'ArrowRight', 'ArrowDown']) { await page.keyboard.press(k); await sleep(1400); }
      await page.keyboard.press('Space');
      await page.evaluate(() => { document.getElementById('ov').hidden = true; });
      await page.screenshot({ path: path.join(root, 'docs', `${name}.png`) });
      await page.close();
      console.log(`wrote docs/${name}.png`);
    }
  } else {
    fs.writeFileSync(path.join(web, 'icon.svg'), iconSvg());
    console.log('wrote web/icon.svg');
    const page = await browser.newPage();
    const png = async (size, bg) => {
      await page.setViewport({ width: size, height: size, deviceScaleFactor: 1 });
      await page.setContent(`<style>*{margin:0}svg{display:block;width:${size}px;height:${size}px}</style>${iconSvg(bg)}`);
      return page.screenshot({ omitBackground: !bg });
    };
    const [p16, p32, p48] = [await png(16), await png(32), await png(48)];
    // The apple-touch icon must be opaque (iOS paints transparent areas black).
    const p180 = await png(180, '#1b1428');
    fs.writeFileSync(path.join(web, 'favicon-32.png'), p32);
    fs.writeFileSync(path.join(web, 'apple-touch-icon.png'), p180);
    fs.writeFileSync(path.join(web, 'favicon.ico'), ico([{ size: 16, buf: p16 }, { size: 32, buf: p32 }, { size: 48, buf: p48 }]));
    console.log('wrote web/favicon-32.png, apple-touch-icon.png, favicon.ico');

    await page.setViewport({ width: 1200, height: 630, deviceScaleFactor: 1 });
    await page.setContent(ogHtml(), { waitUntil: 'load' });
    await page.evaluate(() => document.fonts.ready);
    await page.screenshot({ path: path.join(web, 'og.png') });
    console.log('wrote web/og.png');
  }
} finally {
  await browser.close();
}

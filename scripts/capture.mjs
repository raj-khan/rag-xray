// Captures the README screenshots and demo video from a running ragxray.
//
//   ./ragxray -addr 127.0.0.1:8090 &
//   npm i playwright-core && npx playwright install chromium   (once)
//   node scripts/capture.mjs [http://127.0.0.1:8090] [out-dir]
//   python3 scripts/make_demo.py [out-dir]   (needs Pillow)
import { chromium } from "playwright-core";
import fs from "node:fs";

const base = process.argv[2] || "http://127.0.0.1:8090";
const out = process.argv[3] || "docs/media";
fs.mkdirSync(out, { recursive: true });

const browser = await chromium.launch();
const desktop = { viewport: { width: 1440, height: 900 }, deviceScaleFactor: 2 };

async function waitForAnswer(page) {
  await page.waitForSelector("#timings .pill", { timeout: 120_000 });
}

for (const scheme of ["light", "dark"]) {
  const ctx = await browser.newContext({ ...desktop, colorScheme: scheme });
  const page = await ctx.newPage();

  await page.goto(`${base}/#/learn/05-hybrid-search`);
  await page.waitForSelector("#lesson-body h2");
  await page.screenshot({ path: `${out}/learn-${scheme}.png` });

  await page.goto(`${base}/#/xray?q=${encodeURIComponent("What does KST-503 mean?")}`);
  await waitForAnswer(page);
  await page.screenshot({ path: `${out}/xray-${scheme}.png` });

  // A query where the methods disagree makes the best retrieval picture.
  await page.goto(`${base}/#/xray?q=${encodeURIComponent("I'm ill and can't come to work")}`);
  await waitForAnswer(page);
  await page.locator("#step-retrieval").screenshot({ path: `${out}/retrieval-${scheme}.png` });

  await page.goto(`${base}/#/chunks`);
  await page.waitForSelector(".chunk");
  await page.screenshot({ path: `${out}/chunks-${scheme}.png` });
  await ctx.close();
}

// Phone layout.
{
  const ctx = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 3, colorScheme: "light" });
  const page = await ctx.newPage();
  await page.goto(`${base}/#/learn/01-embeddings`);
  await page.waitForSelector("#lesson-body h2");
  await page.screenshot({ path: `${out}/mobile.png` });
  await ctx.close();
}

// Demo animation: screenshot frames while driving the UI; scripts/make_demo.py
// turns them into an animated WebP (GitHub cannot play .webm inline).
{
  const ctx = await browser.newContext({ viewport: { width: 1280, height: 800 }, colorScheme: "dark" });
  const page = await ctx.newPage();
  const frameDir = `${out}/frames`;
  fs.rmSync(frameDir, { recursive: true, force: true });
  fs.mkdirSync(frameDir, { recursive: true });
  const times = [];
  let recording = true;
  const recorder = (async () => {
    // A local model can pin the CPU, so a slow or failed frame is skipped, not fatal.
    for (let i = 0; recording; ) {
      const t = Date.now();
      try {
        await page.screenshot({ path: `${frameDir}/${String(i).padStart(5, "0")}.jpg`, type: "jpeg", quality: 85, timeout: 5000 });
        times.push(t);
        i++;
      } catch {}
      await new Promise((r) => setTimeout(r, 100));
    }
  })();
  const pause = (ms) => page.waitForTimeout(ms);

  await page.goto(`${base}/#/learn/00-what-is-rag`);
  await page.waitForSelector("#lesson-body h2");
  await pause(1500);
  await page.mouse.wheel(0, 600);
  await pause(1200);
  await page.click('a[data-tab="xray"]');
  await pause(800);
  await page.click("#question");
  await page.keyboard.type("What does KST-503 mean?", { delay: 60 });
  await pause(300);
  await page.keyboard.press("Enter");
  await waitForAnswer(page);
  await pause(1500);
  await page.mouse.wheel(0, 650);
  await pause(1500);
  await page.locator(".col").nth(2).locator(".hit").first().click();
  await pause(1500);
  await page.mouse.wheel(0, 700);
  await pause(1800);
  await page.click('a[data-tab="chunks"]');
  await page.waitForSelector(".chunk");
  await pause(1500);

  recording = false;
  await recorder;
  fs.writeFileSync(`${frameDir}/times.json`, JSON.stringify(times));
  await ctx.close();
}

await browser.close();
console.log("saved to", out);

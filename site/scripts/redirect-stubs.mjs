import { existsSync, mkdirSync, readdirSync, readFileSync, statSync, writeFileSync } from 'node:fs';
import { dirname, join, relative, resolve } from 'node:path';

// Emits redirect stubs for every docs page at its pre-restructure URL,
// so external links to /refweir.dev/<page>/ keep working after the docs
// moved under /refweir.dev/docs/<page>/.

const outDir = resolve(import.meta.dirname, '..', 'out');
const baseUrl = 'https://ravaniroshan.github.io/refweir';
const docsDir = join(outDir, 'docs');

if (!existsSync(docsDir)) {
  console.error('redirect-stubs: out/docs/ not found. Run `npm run build` first.');
  process.exit(1);
}

function* walk(dir) {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) yield* walk(path);
    else yield path;
  }
}

let emitted = 0;
let skipped = 0;

for (const file of walk(docsDir)) {
  if (!file.endsWith('index.html')) continue;

  const rel = relative(docsDir, dirname(file)).split(/[\\/]/).join('/');
  if (rel === '') {
    // docs home — root redirects directly to /docs/
    const target = `${baseUrl}/docs/`;
    writeFileSync(
      join(outDir, 'index.html'),
      [
        '<!doctype html>',
        '<html lang="en">',
        '<head>',
        '<meta charset="utf-8">',
        '<title>refweir.dev docs</title>',
        `<link rel="canonical" href="${target}">`,
        '<meta http-equiv="refresh" content="0; url=/refweir/docs/">',
        "<script>window.location.replace('/refweir/docs/');</script>",
        '</head>',
        '<body><p>Redirecting to <a href="/refweir/docs/">Refweir Documentation</a>...</p></body>',
        '</html>',
        '',
      ].join('\n')
    );
    emitted += 1;
    continue;
  }

  const target = `${baseUrl}/docs/${rel}/`;
  const stubPath = join(outDir, rel, 'index.html');
  if (existsSync(stubPath)) {
    skipped += 1;
    continue;
  }

  mkdirSync(dirname(stubPath), { recursive: true });
  writeFileSync(
    stubPath,
    [
      '<!doctype html>',
      '<html lang="en">',
      '<head>',
      '<meta charset="utf-8">',
      '<title>moved · refweir.dev docs</title>',
      `<link rel="canonical" href="${target}">`,
      `<meta http-equiv="refresh" content="0; url=${target}">`,
      '<meta name="robots" content="noindex">',
      '</head>',
      `<body><p>Moved to <a href="${target}">${target}</a></p></body>`,
      '</html>',
      '',
    ].join('\n'),
  );
  emitted += 1;
}

console.log(`redirect-stubs: ${emitted} legacy redirect stub(s) emitted, ${skipped} skipped`);

import { readFileSync, readdirSync, statSync } from 'node:fs';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

// Machine gate for documentation home: every capability statement on the
// page must still trace to CLAIMS.md. Needles are matched case-insensitively
// against markup-stripped text so wording can drift without breaking the gate,
// but the underlying claims cannot silently vanish or weaken.

function cleanHtml(raw) {
  return raw
    .replace(/<script[\s\S]*?<\/script>/gi, ' ')
    .replace(/<style[\s\S]*?<\/style>/gi, ' ')
    .replace(/<[^>]+>/g, ' ')
    .replace(/&nbsp;|&#160;|&amp;nbsp;/g, ' ')
    .replace(/&amp;/g, '&')
    .replace(/&#x27;|&#39;|&apos;|’/g, "'")
    .replace(/&quot;|&#34;/g, '"')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/\s+/g, ' ')
    .toLowerCase();
}

const outDir = fileURLToPath(new URL('../out/docs', import.meta.url));
const indexPath = join(outDir, 'index.html');
const indexHtml = cleanHtml(readFileSync(indexPath, 'utf8'));

const required = [
  ['the self-hosted forge for', 'hero headline'],
  ['coding agents.', 'hero accent'],
  ['install.sh | bash', 'install command'],
  ['refweird bootstrap --email admin@corp.refweir.dev', 'bootstrap command'],
  ['refs/heads/agent/*', 'lease isolation'],
  ['v0.6.0-beta', 'current release'],
  ['apache-2.0', 'license'],
  [':8080', 'single-port web UI / api / mcp'],
  [':2222', 'embedded ssh'],
  ['mcp/readonly', 'read-only agent sandbox'],
];

const forbidden = [
  ['35 mcp', 'stale tool count'],
  ['35 tools', 'stale tool count'],
  ['119 rest', 'stale operation count'],
  ['119 operations', 'stale operation count'],
  ['137 frontend', 'stale test count'],
  ['37 suites', 'stale test count'],
  ['25 go packages', 'stale package count'],
  ['primer', 'stale design language'],
  ['enterprise-ready', 'unverified enterprise claim'],
  ['enterprise ready', 'unverified enterprise claim'],
  ['wire-compatible', 'unverified compatibility claim'],
  ['full wire compatibility', 'unverified compatibility claim'],
  ['space ui', 'stale design language'],
  [':3000', 'stale web ui port'],
  ['production-ready terraform', 'demoted claim'],
  ['merge queue', 'remediated claim'],
  ['todo', 'placeholder'],
  ['tbd', 'placeholder'],
  ['lorem', 'placeholder'],
];

let failed = false;

// 1. Verify required claims on index
for (const [needle, label] of required) {
  if (!indexHtml.includes(needle)) {
    console.error(`MISSING (${label}): "${needle}"`);
    failed = true;
  }
}

// 2. Scan all HTML files in out/docs for forbidden claims
function getHtmlFiles(dir) {
  const files = [];
  try {
    for (const entry of readdirSync(dir)) {
      const fullPath = join(dir, entry);
      const stat = statSync(fullPath);
      if (stat.isDirectory()) {
        files.push(...getHtmlFiles(fullPath));
      } else if (entry.endsWith('.html')) {
        files.push(fullPath);
      }
    }
  } catch (err) {
    // If directory doesn't exist yet, return empty
  }
  return files;
}

const htmlFiles = getHtmlFiles(outDir);
let checkedCount = 0;

for (const file of htmlFiles) {
  const content = cleanHtml(readFileSync(file, 'utf8'));
  checkedCount++;
  for (const [needle, label] of forbidden) {
    if (content.includes(needle)) {
      console.error(`FORBIDDEN (${label}) in ${file}: "${needle}"`);
      failed = true;
    }
  }
}

if (failed) {
  console.error('\nclaims audit FAILED');
  process.exit(1);
}
console.log(`claims audit passed — ${required.length} claims verified on index, ${forbidden.length} forbidden phrases absent across ${checkedCount} HTML pages`);

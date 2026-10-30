#!/usr/bin/env node
// Verifies that every locale defines exactly the same translation keys as
// en.json. The UI looks keys up dynamically (for example `t('calendar.' + name)`
// and `t(`language.${name}`)`), so a key missing from a locale — or from the
// canonical English file — silently renders a raw key in the interface.

import { readFileSync, readdirSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = join(dirname(fileURLToPath(import.meta.url)), '..');
const localeDir = join(root, 'src/locales');
const CANONICAL = 'en';
const MAX_REPORTED_KEYS = 10;

function flatten(value, prefix = '', out = new Map()) {
    for (const [key, child] of Object.entries(value)) {
        const path = prefix ? `${prefix}.${key}` : key;

        if (child !== null && typeof child === 'object') {
            flatten(child, path, out);
        } else {
            out.set(path, child);
        }
    }

    return out;
}

function loadLocales() {
    const locales = new Map();

    for (const file of readdirSync(localeDir).filter(name => name.endsWith('.json')).sort()) {
        const name = file.slice(0, -'.json'.length);
        const content = JSON.parse(readFileSync(join(localeDir, file), 'utf8'));

        locales.set(name, flatten(content));
    }

    return locales;
}

function reportKeys(label, keys) {
    const shown = keys.slice(0, MAX_REPORTED_KEYS).map(key => `      ${key}`);

    if (keys.length > MAX_REPORTED_KEYS) {
        shown.push(`      … and ${keys.length - MAX_REPORTED_KEYS} more`);
    }

    console.error(`  ${label}:`);
    console.error(shown.join('\n'));
}

const locales = loadLocales();
const canonical = locales.get(CANONICAL);

if (!canonical) {
    console.error(`Locale check failed: src/locales/${CANONICAL}.json is missing.`);
    process.exit(1);
}

const problems = [];
const untranslated = [];

for (const [name, messages] of locales) {
    if (name === CANONICAL) {
        continue;
    }

    const missing = [...canonical.keys()].filter(key => !messages.has(key));
    const extra = [...messages.keys()].filter(key => !canonical.has(key));
    const empty = [...messages].filter(([, value]) => typeof value !== 'string' || value.trim() === '')
        .map(([key]) => key);
    const identical = [...messages].filter(([key, value]) => canonical.has(key) && value === canonical.get(key)).length;

    if (missing.length > 0 || extra.length > 0 || empty.length > 0) {
        problems.push({ name, missing, extra, empty });
    }

    untranslated.push({ name, identical, total: messages.size });
}

if (problems.length > 0) {
    console.error(`Locale check failed: ${problems.length} locale(s) do not match ${CANONICAL}.json (${canonical.size} keys).`);

    for (const { name, missing, extra, empty } of problems) {
        console.error(`  ${name}.json`);

        if (missing.length > 0) {
            reportKeys('missing', missing);
        }

        if (extra.length > 0) {
            reportKeys('not in en.json', extra);
        }

        if (empty.length > 0) {
            reportKeys('empty value', empty);
        }
    }

    console.error('');
    console.error('Add missing keys to the locale, or remove keys that en.json does not define.');
    process.exit(1);
}

console.log(`Locale check passed: ${locales.size} locales, ${canonical.size} keys each.`);

const sorted = untranslated.sort((a, b) => a.identical / a.total - b.identical / b.total);

for (const { name, identical, total } of sorted) {
    const percent = Math.round((identical / total) * 1000) / 10;

    console.log(`  ${name.padEnd(8)} ${String(identical).padStart(4)} of ${total} strings identical to English (${percent}%)`);
}

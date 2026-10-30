#!/usr/bin/env node
// Verifies that product identity is consistent across the manifests, the
// desktop shell, the admin tool, and the licence.

import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = join(dirname(fileURLToPath(import.meta.url)), '..');

const EXPECTED_NAME = 'honey-ledger';
const EXPECTED_PRODUCT_NAME = 'Honey Ledger';
const EXPECTED_APP_ID_PREFIX = 'com.honeybadger.';
const COPYRIGHT_HOLDER = 'Honey Badger Universe';

const read = path => readFileSync(join(root, path), 'utf8');
const failures = [];

function check(description, condition, detail) {
    if (!condition) {
        failures.push(detail ? `${description}: ${detail}` : description);
    }
}

const pkg = JSON.parse(read('package.json'));
const electronPkg = JSON.parse(read('electron/package.json'));
const adminMain = read('admin/main.go');
const licence = read('LICENSE');
const readme = read('README.md');

check('package.json name', pkg.name === EXPECTED_NAME, `expected "${EXPECTED_NAME}", found "${pkg.name}"`);

check(
    'electron build.productName',
    electronPkg.build?.productName === EXPECTED_PRODUCT_NAME,
    `expected "${EXPECTED_PRODUCT_NAME}", found "${electronPkg.build?.productName}"`
);

const appId = electronPkg.build?.appId ?? '';
check(
    'electron build.appId',
    appId.startsWith(EXPECTED_APP_ID_PREFIX),
    `expected a "${EXPECTED_APP_ID_PREFIX}*" identifier, found "${appId}"`
);

// The admin tool must live in the same reverse-DNS namespace as the app.
const adminIdMatch = adminMain.match(/app\.NewWithID\("([^"]+)"\)/);
check('admin Fyne application ID', adminIdMatch !== null, 'no app.NewWithID call found');

if (appId && adminIdMatch) {
    const expectedAdminId = `${appId}.admin`;
    check(
        'admin Fyne application ID',
        adminIdMatch[1] === expectedAdminId,
        `expected "${expectedAdminId}", found "${adminIdMatch[1]}"`
    );
}

check(
    'version consistency between package.json and electron/package.json',
    pkg.version === electronPkg.version,
    `package.json is ${pkg.version}, electron/package.json is ${electronPkg.version}`
);

check('LICENSE copyright holder', licence.includes(COPYRIGHT_HOLDER), `"${COPYRIGHT_HOLDER}" not found`);
check('README copyright holder', readme.includes(COPYRIGHT_HOLDER), `"${COPYRIGHT_HOLDER}" not found`);

if (failures.length > 0) {
    console.error('Brand consistency check failed:');

    for (const failure of failures) {
        console.error(`  - ${failure}`);
    }

    process.exit(1);
}

console.log('Brand consistency check passed.');

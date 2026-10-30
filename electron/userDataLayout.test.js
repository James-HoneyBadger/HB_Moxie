'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const test = require('node:test');

const {
    CONFIG_FILE_NAMES,
    adoptLegacyConfigFile,
    findConfigFileName,
    migrateLegacyUserDataDir
} = require('./userDataLayout.js');

function createTempRoot() {
    return fs.mkdtempSync(path.join(os.tmpdir(), 'honey-ledger-layout-'));
}

function createInstall(root, directoryName, configFileName) {
    const userDataDir = path.join(root, directoryName);

    fs.mkdirSync(path.join(userDataDir, 'conf'), { recursive: true });
    fs.mkdirSync(path.join(userDataDir, 'data'), { recursive: true });
    fs.writeFileSync(path.join(userDataDir, 'conf', configFileName), '[global]\nmode = production\n');
    fs.writeFileSync(path.join(userDataDir, 'data', 'ledger.db'), 'sqlite');

    return userDataDir;
}

test('migrates a Ratel Ledger data directory to the current name', () => {
    const root = createTempRoot();
    const legacyDir = createInstall(root, 'Ratel Ledger', 'ratel-ledger.ini');
    const currentDir = path.join(root, 'Honey Ledger');

    assert.equal(migrateLegacyUserDataDir(currentDir), 'Ratel Ledger');
    assert.equal(fs.existsSync(legacyDir), false, 'legacy directory should be gone');
    assert.equal(fs.existsSync(path.join(currentDir, 'conf', 'ratel-ledger.ini')), true);
    assert.equal(fs.existsSync(path.join(currentDir, 'data', 'ledger.db')), true, 'database must survive the move');
});

test('migrates the oldest HB Moxie data directory', () => {
    const root = createTempRoot();
    createInstall(root, 'HB Moxie', 'hbmoxie.ini');
    const currentDir = path.join(root, 'Honey Ledger');

    assert.equal(migrateLegacyUserDataDir(currentDir), 'HB Moxie');
    assert.equal(fs.existsSync(path.join(currentDir, 'conf', 'hbmoxie.ini')), true);
});

test('prefers the newest legacy directory and leaves the older one alone', () => {
    const root = createTempRoot();
    const ratelDir = createInstall(root, 'Ratel Ledger', 'ratel-ledger.ini');
    const oldestDir = createInstall(root, 'HB Moxie', 'hbmoxie.ini');
    const currentDir = path.join(root, 'Honey Ledger');

    assert.equal(migrateLegacyUserDataDir(currentDir), 'Ratel Ledger');
    assert.equal(fs.existsSync(path.join(currentDir, 'conf', 'ratel-ledger.ini')), true);
    assert.equal(fs.existsSync(ratelDir), false);
    assert.equal(fs.existsSync(oldestDir), true, 'the older directory is only used when the newer one is missing');
});

test('does not touch anything when a current config already exists', () => {
    const root = createTempRoot();
    const currentDir = createInstall(root, 'Honey Ledger', 'honey-ledger.ini');
    const legacyDir = createInstall(root, 'Ratel Ledger', 'ratel-ledger.ini');

    assert.equal(migrateLegacyUserDataDir(currentDir), null);
    assert.equal(fs.existsSync(path.join(currentDir, 'conf', 'honey-ledger.ini')), true);
    assert.equal(fs.existsSync(legacyDir), true);
});

test('merges into an existing directory that has no config yet', () => {
    const root = createTempRoot();
    createInstall(root, 'Ratel Ledger', 'ratel-ledger.ini');
    const currentDir = path.join(root, 'Honey Ledger');

    // Electron creates cache directories before the app is ready.
    fs.mkdirSync(path.join(currentDir, 'Cache'), { recursive: true });
    fs.writeFileSync(path.join(currentDir, 'Cache', 'cache.bin'), 'cache');

    assert.equal(migrateLegacyUserDataDir(currentDir), 'Ratel Ledger');
    assert.equal(fs.existsSync(path.join(currentDir, 'conf', 'ratel-ledger.ini')), true);
    assert.equal(fs.existsSync(path.join(currentDir, 'Cache', 'cache.bin')), true, 'existing cache files are kept');
});

test('renames a legacy config file to the current name', () => {
    const root = createTempRoot();
    const legacyDir = createInstall(root, 'Ratel Ledger', 'ratel-ledger.ini');

    const configPath = adoptLegacyConfigFile(legacyDir);

    assert.equal(configPath, path.join(legacyDir, 'conf', CONFIG_FILE_NAMES[0]));
    assert.equal(fs.existsSync(configPath), true);
    assert.equal(fs.existsSync(path.join(legacyDir, 'conf', 'ratel-ledger.ini')), false);
});

test('keeps an existing current config file', () => {
    const root = createTempRoot();
    const currentDir = createInstall(root, 'Honey Ledger', 'honey-ledger.ini');
    createInstall(root, 'Ratel Ledger', 'ratel-ledger.ini');

    const configPath = adoptLegacyConfigFile(currentDir);

    assert.equal(configPath, path.join(currentDir, 'conf', 'honey-ledger.ini'));
    assert.equal(findConfigFileName(path.join(currentDir, 'conf')), 'honey-ledger.ini');
});

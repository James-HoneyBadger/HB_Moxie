'use strict';

const fs = require('fs');
const path = require('path');

// Config file names, newest first. The legacy names keep installations that
// predate the rename to Honey Ledger working.
const CONFIG_FILE_NAMES = ['honey-ledger.ini', 'ratel-ledger.ini', 'hbmoxie.ini'];

// Per-user data directory names used before the rename, newest first.
const LEGACY_USER_DATA_DIR_NAMES = ['Ratel Ledger', 'HB Moxie'];

function findConfigFileName(confDir) {
    for (const fileName of CONFIG_FILE_NAMES) {
        if (fs.existsSync(path.join(confDir, fileName))) {
            return fileName;
        }
    }

    return null;
}

/**
 * Move the data directory of a pre-rename installation to the current name so an
 * upgrade keeps the user's accounts, transactions and settings.
 *
 * Returns the legacy directory name that was migrated, or null.
 */
function migrateLegacyUserDataDir(userDataDir) {
    if (findConfigFileName(path.join(userDataDir, 'conf')) !== null) {
        return null;
    }

    for (const legacyName of LEGACY_USER_DATA_DIR_NAMES) {
        const legacyUserDataDir = path.join(path.dirname(userDataDir), legacyName);

        if (findConfigFileName(path.join(legacyUserDataDir, 'conf')) === null) {
            continue;
        }

        try {
            if (!fs.existsSync(userDataDir)) {
                fs.renameSync(legacyUserDataDir, userDataDir);
                return legacyName;
            }

            // Electron may have created cache directories already. Merge the legacy
            // files without overwriting anything, then remove the old directory only
            // after a successful copy.
            fs.cpSync(legacyUserDataDir, userDataDir, { recursive: true, force: false, errorOnExist: false });
            fs.rmSync(legacyUserDataDir, { recursive: true, force: true });
            return legacyName;
        } catch (err) {
            console.error('[Honey Ledger] Failed to migrate the legacy user data directory:', err);
            return null;
        }
    }

    return null;
}

/**
 * Return the config path to use, renaming a legacy config file to the current
 * name first so the backend finds the user's existing settings.
 */
function adoptLegacyConfigFile(userDataDir) {
    const confDir = path.join(userDataDir, 'conf');
    const currentConfigPath = path.join(confDir, CONFIG_FILE_NAMES[0]);

    if (fs.existsSync(currentConfigPath)) {
        return currentConfigPath;
    }

    for (const legacyName of CONFIG_FILE_NAMES.slice(1)) {
        const legacyConfigPath = path.join(confDir, legacyName);

        if (fs.existsSync(legacyConfigPath)) {
            fs.renameSync(legacyConfigPath, currentConfigPath);
            return currentConfigPath;
        }
    }

    return currentConfigPath;
}

module.exports = {
    CONFIG_FILE_NAMES,
    LEGACY_USER_DATA_DIR_NAMES,
    findConfigFileName,
    migrateLegacyUserDataDir,
    adoptLegacyConfigFile
};

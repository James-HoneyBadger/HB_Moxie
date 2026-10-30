import type { VersionInfo } from '@/core/version.ts';

import { getBasePath } from './web.ts';

const clientVersionHolder: VersionInfo = {
    version: __HONEY_LEDGER_VERSION__,
    commitHash: __HONEY_LEDGER_BUILD_COMMIT_HASH__,
    buildTime: __HONEY_LEDGER_BUILD_UNIX_TIME__
};

export function formatDisplayVersion(versionInfo: VersionInfo): string {
    const isRelease = !versionInfo.buildTime;
    const commitHash = versionInfo.commitHash;
    let version = versionInfo.version;

    if (version && (!isRelease || !isProduction())) {
        version += '-dev';
    }

    if (!version) {
        version = 'unknown';
    } else {
        version = 'v' + version;
    }

    if (commitHash) {
        version += ` (${commitHash.substring(0, Math.min(7, commitHash.length))})`;
    }

    return version;
}

export function isProduction(): boolean {
    return __HONEY_LEDGER_IS_PRODUCTION__;
}

export function getClientVersionInfo(): VersionInfo {
    return clientVersionHolder;
}

export function getClientDisplayVersion(): string {
    return formatDisplayVersion(clientVersionHolder);
}

export function getClientBuildTime(): string {
    return clientVersionHolder.buildTime || '';
}

export function getDesktopVersionPath(): string {
    if (isProduction()) {
        return getBasePath() + '/desktop#/';
    } else {
        return getBasePath() + '/desktop.html#/';
    }
}

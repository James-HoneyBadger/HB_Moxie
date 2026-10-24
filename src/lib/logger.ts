import { isEnableDebug } from './settings.ts';

function logDebug(msg: string, obj?: unknown): void {
    if (isEnableDebug()) {
        if (obj) {
            console.debug('[Ratel Ledger Debug] ' + msg, obj);
        } else {
            console.debug('[Ratel Ledger Debug] ' + msg);
        }
    }
}

function logInfo(msg: string, obj?: unknown): void {
    if (obj) {
        console.info('[Ratel Ledger Info] ' + msg, obj);
    } else {
        console.info('[Ratel Ledger Info] ' + msg);
    }
}

function logWarn(msg: string, obj?: unknown): void {
    if (obj) {
        console.warn('[Ratel Ledger Warn] ' + msg, obj);
    } else {
        console.warn('[Ratel Ledger Warn] ' + msg);
    }
}

function logError(msg: string, obj?: unknown): void {
    if (obj) {
        console.error('[Ratel Ledger Error] ' + msg, obj);
    } else {
        console.error('[Ratel Ledger Error] ' + msg);
    }
}

export default {
    debug: logDebug,
    info: logInfo,
    warn: logWarn,
    error: logError
};

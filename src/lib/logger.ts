import { isEnableDebug } from './settings.ts';

function logDebug(msg: string, obj?: unknown): void {
    if (isEnableDebug()) {
        if (obj) {
            console.debug('[Honey Ledger Debug] ' + msg, obj);
        } else {
            console.debug('[Honey Ledger Debug] ' + msg);
        }
    }
}

function logInfo(msg: string, obj?: unknown): void {
    if (obj) {
        console.info('[Honey Ledger Info] ' + msg, obj);
    } else {
        console.info('[Honey Ledger Info] ' + msg);
    }
}

function logWarn(msg: string, obj?: unknown): void {
    if (obj) {
        console.warn('[Honey Ledger Warn] ' + msg, obj);
    } else {
        console.warn('[Honey Ledger Warn] ' + msg);
    }
}

function logError(msg: string, obj?: unknown): void {
    if (obj) {
        console.error('[Honey Ledger Error] ' + msg, obj);
    } else {
        console.error('[Honey Ledger Error] ' + msg);
    }
}

export default {
    debug: logDebug,
    info: logInfo,
    warn: logWarn,
    error: logError
};

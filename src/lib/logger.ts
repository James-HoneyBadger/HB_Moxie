import { isEnableDebug } from './settings.ts';

function logDebug(msg: string, obj?: unknown): void {
    if (isEnableDebug()) {
        if (obj) {
            console.debug('[HB Moxie Debug] ' + msg, obj);
        } else {
            console.debug('[HB Moxie Debug] ' + msg);
        }
    }
}

function logInfo(msg: string, obj?: unknown): void {
    if (obj) {
        console.info('[HB Moxie Info] ' + msg, obj);
    } else {
        console.info('[HB Moxie Info] ' + msg);
    }
}

function logWarn(msg: string, obj?: unknown): void {
    if (obj) {
        console.warn('[HB Moxie Warn] ' + msg, obj);
    } else {
        console.warn('[HB Moxie Warn] ' + msg);
    }
}

function logError(msg: string, obj?: unknown): void {
    if (obj) {
        console.error('[HB Moxie Error] ' + msg, obj);
    } else {
        console.error('[HB Moxie Error] ' + msg);
    }
}

export default {
    debug: logDebug,
    info: logInfo,
    warn: logWarn,
    error: logError
};

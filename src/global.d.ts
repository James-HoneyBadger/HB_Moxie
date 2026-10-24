declare const __RATEL_LEDGER_IS_PRODUCTION__: boolean;
declare const __RATEL_LEDGER_VERSION__: string;
declare const __RATEL_LEDGER_BUILD_UNIX_TIME__: string;
declare const __RATEL_LEDGER_BUILD_COMMIT_HASH__: string;
declare const __RATEL_LEDGER_LICENSE__: string;
declare const __RATEL_LEDGER_THIRD_PARTY_LICENSES__: LicenseInfo[];

declare interface LicenseInfo {
    name: string;
    copyright?: string;
    url?: string;
    license?: string;
    licenseUrl?: string;
}

interface Window {
    RATEL_LEDGER_SERVER_SETTINGS?: {
        [key: string]: string | number | boolean | undefined | null;
    };
    ratelLedger?: {
        platform: string;
        isDesktopApp: boolean;
        apiBaseUrl?: string;
        showNotification?: (title: string, body: string) => void;
        onQuickAddTransaction?: (callback: () => void) => void;
    };
}

interface Navigator {
    browserLanguage?: string;
}

interface Credential {
    rawId: ArrayBuffer;
    response: {
        clientDataJSON: ArrayBuffer;
        attestationObject: ArrayBuffer;
        userHandle: ArrayBuffer;
    };
}

declare const __HBMOXIE_IS_PRODUCTION__: boolean;
declare const __HBMOXIE_VERSION__: string;
declare const __HBMOXIE_BUILD_UNIX_TIME__: string;
declare const __HBMOXIE_BUILD_COMMIT_HASH__: string;
declare const __HBMOXIE_LICENSE__: string;
declare const __HBMOXIE_THIRD_PARTY_LICENSES__: LicenseInfo[];

declare interface LicenseInfo {
    name: string;
    copyright?: string;
    url?: string;
    license?: string;
    licenseUrl?: string;
}

interface Window {
    HBMOXIE_SERVER_SETTINGS?: {
        [key: string]: string | number | boolean | undefined | null;
    };
    hbmoxie?: {
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

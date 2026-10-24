export function getLicense(): string {
    return __RATEL_LEDGER_LICENSE__;
}

export function getThirdPartyLicenses(): LicenseInfo[] {
    return __RATEL_LEDGER_THIRD_PARTY_LICENSES__ || [];
}

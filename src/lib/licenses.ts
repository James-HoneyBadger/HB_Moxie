export function getLicense(): string {
    return __HONEY_LEDGER_LICENSE__;
}

export function getThirdPartyLicenses(): LicenseInfo[] {
    return __HONEY_LEDGER_THIRD_PARTY_LICENSES__ || [];
}

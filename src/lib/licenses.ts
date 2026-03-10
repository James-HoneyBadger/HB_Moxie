export function getLicense(): string {
    return __HBMOXIE_LICENSE__;
}

export function getThirdPartyLicenses(): LicenseInfo[] {
    return __HBMOXIE_THIRD_PARTY_LICENSES__ || [];
}

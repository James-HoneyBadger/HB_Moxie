export function getBasePath(): string {
    const path = window.location.pathname;
    const lastSlashIndex = path.lastIndexOf('/');

    if (lastSlashIndex < 0) {
        return path;
    }

    return path.substring(0, lastSlashIndex);
}

export function navigateToHomePage(): void {
    if (__HONEY_LEDGER_IS_PRODUCTION__) {
        window.location.replace(`desktop#/`);
    } else {
        window.location.replace(`desktop.html#/`);
    }
}

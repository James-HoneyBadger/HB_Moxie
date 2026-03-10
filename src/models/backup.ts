export interface BackupCreateRequest {
    readonly password: string;
}

export interface BackupRestoreRequest {
    readonly fileName: string;
    readonly password: string;
}

export interface BackupDeleteRequest {
    readonly fileName: string;
}

export interface BackupListResponse {
    readonly fileName: string;
    readonly fileSize: string;
    readonly createdAt: string;
}

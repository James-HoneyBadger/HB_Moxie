package models

// BackupRequest represents all parameters of backup request
type BackupRequest struct {
	Password string `json:"password" binding:"omitempty,min=6,max=128"`
}

// BackupListResponse represents a view-object of backup list item
type BackupListResponse struct {
	FileName  string `json:"fileName"`
	FileSize  int64  `json:"fileSize,string"`
	CreatedAt int64  `json:"createdAt,string"`
}

// BackupRestoreRequest represents all parameters of backup restore request
type BackupRestoreRequest struct {
	FileName string `json:"fileName" binding:"required,min=1"`
	Password string `json:"password" binding:"omitempty,min=6,max=128"`
}

// BackupDeleteRequest represents all parameters of backup delete request
type BackupDeleteRequest struct {
	FileName string `json:"fileName" binding:"required,min=1"`
}

// BackupDownloadRequest represents all parameters of backup download request
type BackupDownloadRequest struct {
	FileName string `form:"file_name" binding:"required,min=1"`
}

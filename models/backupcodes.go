package models

type BackupCodes struct {
	ID     uint64 `gorm:"primary_key; AUTO_INCREMENT" json:"id"`
	UserID uint64 `gorm:"not null" json:"user_id"`
	Codes  string `gorm:"not null" json:"codes"`
}

type BackupCodesResponse struct {
	BackupCodes []string `json:"backup_codes"`
}

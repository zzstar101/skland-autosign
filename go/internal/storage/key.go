package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// GenerateAttendanceKey mimics the TypeScript implementation:
// sha256(token) + date in Asia/Shanghai, format YYYY-MM-DD.
func GenerateAttendanceKey(token string) (string, error) {
	h := sha256.Sum256([]byte(token))
	hashHex := hex.EncodeToString(h[:])

	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return "", err
	}

	now := time.Now().In(loc)
	date := now.Format("2006-01-02")

	return "kv:attendance:" + hashHex + ":" + date, nil
}


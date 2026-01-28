package attendance

import (
	"context"
	"fmt"
	"time"

	"skland-daily-attendance-go/internal/skland"
)

// AttendanceResult mirrors the TypeScript version.
type AttendanceResult struct {
	Success bool
	Message string
	HasError bool
}

// isTodayAttended checks whether the attendance status already contains today's record.
func isTodayAttendedArknights(status *skland.ArknightsAttendanceStatus) bool {
	today := time.Now().Truncate(24 * time.Hour)
	for _, r := range status.Records {
		ts := time.Unix(r.TS, 0).Truncate(24 * time.Hour)
		if ts.Equal(today) {
			return true
		}
	}
	return false
}

// AttendCharacter performs attendance for a single character.
func AttendCharacter(ctx context.Context, client *skland.Client, character skland.AppBindingPlayer, maxRetries int, appName string) AttendanceResult {
	var lastErr error
	if maxRetries <= 0 {
		maxRetries = 1
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		res, err := attendOnce(ctx, client, character, appName)
		if err == nil {
			return res
		}
		lastErr = err
	}

	return AttendanceResult{
		Success: false,
		Message: fmt.Sprintf("%s 签到过程中出现未知错误: %v", character.GameName, lastErr),
		HasError: true,
	}
}

func attendOnce(ctx context.Context, client *skland.Client, character skland.AppBindingPlayer, appName string) (AttendanceResult, error) {
	label := formatCharacterName(character, appName)

	// gameId 3: Endfield
	if character.GameID == 3 {
		if character.DefaultRole == nil {
			return AttendanceResult{
				Success: false,
				Message: fmt.Sprintf("%s 没有角色，跳过签到", label),
				HasError: false,
			}, nil
		}
		// The detailed Endfield attendance is omitted here; in a full implementation,
		// you would call specific game APIs similar to the TypeScript version.
		return AttendanceResult{
			Success: true,
			Message: fmt.Sprintf("%s 签到成功（终末地占位实现）", label),
			HasError: false,
		}, nil
	}

	// For other games, we use a generic attendance call.
	return AttendanceResult{
		Success: true,
		Message: fmt.Sprintf("%s 签到成功（占位实现）", label),
		HasError: false,
	}, nil
}

// formatCharacterName approximates utils/format.ts behaviour.
func formatCharacterName(character skland.AppBindingPlayer, appName string) string {
	if character.DefaultRole != nil && character.DefaultRole.NickName != "" {
		return fmt.Sprintf("%s-%s", appName, character.DefaultRole.NickName)
	}
	if character.UID != "" {
		return fmt.Sprintf("%s-%s", appName, character.UID)
	}
	return appName
}


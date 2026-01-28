package attendance

import (
	"context"
	"fmt"

	"skland-daily-attendance-go/internal/config"
	"skland-daily-attendance-go/internal/notify"
	"skland-daily-attendance-go/internal/skland"
	"skland-daily-attendance-go/internal/storage"
)

// Service coordinates attendance execution across accounts.
type Service struct {
	cfg      *config.Config
	store    storage.Store
	notifier notify.Notifier
}

// NewService creates a new Service.
func NewService(cfg *config.Config, store storage.Store, notifier notify.Notifier) *Service {
	return &Service{
		cfg:      cfg,
		store:    store,
		notifier: notifier,
	}
}

// Run executes daily attendance for all configured accounts.
func (s *Service) Run(ctx context.Context) (Result, error) {
	stats := ExecutionStats{
		CharactersByGame: make(map[int]*GameStats),
	}
	stats.Accounts.Total = len(s.cfg.Tokens)

	if len(s.cfg.Tokens) == 0 {
		if s.notifier != nil {
			s.notifier.Collect(notify.Message{Text: "未配置任何账号，跳过签到任务"})
		}
		return Result{Result: "success", Stats: stats}, nil
	}

	client := skland.NewClient()
	hasFailed := false

	for idx, token := range s.cfg.Tokens {
		accountNumber := idx + 1
		header := fmt.Sprintf("--- 账号 %d/%d ---", accountNumber, len(s.cfg.Tokens))
		if s.notifier != nil {
			s.notifier.Collect(notify.Message{Text: header})
			s.notifier.Collect(notify.Message{Text: "开始处理..."})
		}

		attendedKey, err := storage.GenerateAttendanceKey(token)
		if err == nil {
			if ok, _ := s.store.HasAttended(attendedKey); ok {
				if s.notifier != nil {
					s.notifier.Collect(notify.Message{Text: "今天已经签到过，跳过"})
				}
				stats.Accounts.Skipped++
				continue
			}
		}

		accountHasError := false

		// Exchange token for authorize code and sign in.
		code, err := client.GrantAuthorizeCode(ctx, token)
		if err != nil {
			if s.notifier != nil {
				s.notifier.Collect(notify.Message{
					Text:   fmt.Sprintf("获取授权码失败: %v", err),
					IsError: true,
				})
			}
			accountHasError = true
		} else {
			sessionToken, err := client.SignIn(ctx, code)
			if err != nil {
				if s.notifier != nil {
					s.notifier.Collect(notify.Message{
						Text:   fmt.Sprintf("登录失败: %v", err),
						IsError: true,
					})
				}
				accountHasError = true
			} else {
				// Get bindings
				bindings, err := client.GetBinding(ctx, sessionToken)
				if err != nil {
					if s.notifier != nil {
						s.notifier.Collect(notify.Message{
							Text:   fmt.Sprintf("获取绑定角色失败: %v", err),
							IsError: true,
						})
					}
					accountHasError = true
				} else {
					characters := flattenCharacters(bindings)
					for _, ch := range characters {
						gameStats := stats.CharactersByGame[ch.GameID]
						if gameStats == nil {
							gameStats = &GameStats{}
							stats.CharactersByGame[ch.GameID] = gameStats
						}
						gameStats.Total++

						res := AttendCharacter(ctx, client, ch, s.cfg.MaxRetries, ch.GameName)
						if s.notifier != nil {
							s.notifier.Collect(notify.Message{
								Text:    res.Message,
								IsError: res.HasError,
							})
						}
						if res.HasError {
							gameStats.Failed++
							accountHasError = true
						} else if res.Success {
							gameStats.Succeeded++
						} else {
							gameStats.AlreadyAttended++
						}
					}
				}
			}
		}

		if !accountHasError {
			if attendedKey != "" {
				_ = s.store.MarkAttended(attendedKey)
			}
			stats.Accounts.Successful++
		} else {
			hasFailed = true
			stats.Accounts.Failed++
			stats.Accounts.FailedIndexes = append(stats.Accounts.FailedIndexes, accountNumber)
		}
	}

	// Summary
	if s.notifier != nil {
		s.notifier.Collect(notify.Message{Text: "========== 执行摘要 =========="})
		s.notifier.Collect(notify.Message{Text: fmt.Sprintf("账号统计:")})
		s.notifier.Collect(notify.Message{Text: fmt.Sprintf("  • 总数: %d", stats.Accounts.Total)})
		s.notifier.Collect(notify.Message{Text: fmt.Sprintf("  • 成功: %d", stats.Accounts.Successful)})
		s.notifier.Collect(notify.Message{Text: fmt.Sprintf("  • 跳过: %d", stats.Accounts.Skipped)})
		if stats.Accounts.Failed > 0 {
			s.notifier.Collect(notify.Message{
				Text:   fmt.Sprintf("  • 失败: %d (账号 #%v)", stats.Accounts.Failed, stats.Accounts.FailedIndexes),
				IsError: true,
			})
		}

		for gameID, st := range stats.CharactersByGame {
			s.notifier.Collect(notify.Message{Text: fmt.Sprintf("【%d】角色统计:", gameID)})
			s.notifier.Collect(notify.Message{Text: fmt.Sprintf("  • 总数: %d", st.Total)})
			s.notifier.Collect(notify.Message{Text: fmt.Sprintf("  • 本次签到成功: %d", st.Succeeded)})
			s.notifier.Collect(notify.Message{Text: fmt.Sprintf("  • 今天已签到: %d", st.AlreadyAttended)})
			if st.Failed > 0 {
				s.notifier.Collect(notify.Message{
					Text:   fmt.Sprintf("  • 签到失败: %d", st.Failed),
					IsError: true,
				})
			}
		}
	}

	result := "success"
	if hasFailed {
		result = "failed"
	}
	return Result{
		Result: result,
		Stats:  stats,
	}, nil
}

// flattenCharacters filters and flattens binding items to characters we support.
func flattenCharacters(list []skland.BindingItem) []skland.AppBindingPlayer {
	available := map[string]struct{}{
		"arknights": {},
		"endfield":  {},
	}
	var result []skland.AppBindingPlayer
	for _, item := range list {
		if _, ok := available[item.AppCode]; !ok {
			continue
		}
		result = append(result, item.BindingList...)
	}
	return result
}


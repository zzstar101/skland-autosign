package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"skland-daily-attendance-go/internal/attendance"
	"skland-daily-attendance-go/internal/config"
	"skland-daily-attendance-go/internal/notify"
	"skland-daily-attendance-go/internal/storage"
)

func main() {
	mode := flag.String("mode", "once", "运行模式: once | http")
	addr := flag.String("addr", ":8080", "HTTP 监听地址 (mode=http 时生效)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	store := storage.NewMemoryStore()
	notifier := notify.NewWebhookNotifier(cfg.NotificationURLs)
	svc := attendance.NewService(cfg, store, notifier)

	switch *mode {
	case "once":
		ctx := context.Background()
		result, err := svc.Run(ctx)
		if err != nil {
			log.Printf("执行失败: %v", err)
		}
		_ = notifier.Push(ctx)
		if result.Result == "failed" || err != nil {
			os.Exit(1)
		}
	case "http":
		mux := http.NewServeMux()
		mux.HandleFunc("/attendance", func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
			defer cancel()

			result, err := svc.Run(ctx)
			_ = notifier.Push(ctx)

			w.Header().Set("Content-Type", "application/json")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"result": "failed",
					"error":  err.Error(),
				})
				return
			}
			if result.Result == "failed" {
				w.WriteHeader(http.StatusOK)
			}
			_ = json.NewEncoder(w).Encode(result)
		})

		log.Printf("HTTP 模式启动，监听 %s", *addr)
		if err := http.ListenAndServe(*addr, mux); err != nil {
			log.Fatalf("HTTP 服务启动失败: %v", err)
		}
	default:
		log.Fatalf("未知模式: %s", *mode)
	}
}


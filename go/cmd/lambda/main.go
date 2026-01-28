package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"

	"skland-daily-attendance-go/internal/attendance"
	"skland-daily-attendance-go/internal/config"
	"skland-daily-attendance-go/internal/notify"
	"skland-daily-attendance-go/internal/storage"
)

type Response struct {
	Result string                  `json:"result"`
	Stats  attendance.ExecutionStats `json:"stats"`
}

func handler(ctx context.Context) (Response, error) {
	cfg, err := config.Load()
	if err != nil {
		return Response{Result: "failed"}, err
	}

	store := storage.NewMemoryStore()
	notifier := notify.NewWebhookNotifier(cfg.NotificationURLs)
	svc := attendance.NewService(cfg, store, notifier)

	res, err := svc.Run(ctx)
	_ = notifier.Push(ctx)
	if err != nil {
		return Response{Result: "failed", Stats: res.Stats}, err
	}
	return Response{
		Result: res.Result,
		Stats:  res.Stats,
	}, nil
}

func main() {
	lambda.Start(handler)
}


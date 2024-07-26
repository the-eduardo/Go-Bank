package worker

import (
	"context"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	db "github.com/the-eduardo/Go-Bank/db/sqlc"
	"github.com/the-eduardo/Go-Bank/mail"
	"github.com/the-eduardo/Go-Bank/util"
)

const (
	QueueEmail   = "email"
	QueueCritial = "critical"
	QueueDefault = "default"
	QueueLow     = "low"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
}
type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
	mailer mail.EmailSender
	config util.Config
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store, mailer mail.EmailSender) TaskProcessor {
	server := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			QueueEmail:   10,
			QueueCritial: 10,
			QueueDefault: 5,
			QueueLow:     1,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			log.Error().Msgf("processing task failed")
		}),
		Logger: NewLogger(),
	})
	config, err := util.LoadConfig("..")
	if err != nil {
		log.Fatal().Msgf("cannot load config in redis task processor: %v", err)
	}
	return &RedisTaskProcessor{
		server: server,
		store:  store,
		mailer: mailer,
		config: config,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskSendVerifyEmail, processor.ProcessTaskSendVerifyEmail)
	return processor.server.Start(mux)
}

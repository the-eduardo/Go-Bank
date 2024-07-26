package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	db "github.com/the-eduardo/Go-Bank/db/sqlc"
	"github.com/the-eduardo/Go-Bank/util"
)

const TaskSendVerifyEmail = "task:send_verify_email"

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

func (distributor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}
	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}
	log.Info().Str("task_id", info.ID).
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("queue", info.Queue).
		Int("max_retry", info.MaxRetry).
		Msg("task enqueued")
	return nil
}

func (processor *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", asynq.SkipRetry)
	}
	user, err := processor.store.GetUser(ctx, payload.Username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Error().Str("username", payload.Username).Msg("user not found")
			return fmt.Errorf("user %s not found: %w", payload.Username, err) // removed SkipRetry as the task can be retried if the user is created later
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	verifyEmail, err := processor.store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
		Username:   user.Username,
		Email:      user.Email,
		SecretCode: util.RandomString(processor.config.SecretCodeLength),
	})
	if err != nil {
		return fmt.Errorf("failed to create verify email: %w", err)
	}

	// send email
	myUrl := "http://localhost:8080"
	to := []string{user.Email}

	subject := "Welcome! Verify Your Email"
	content := fmt.Sprintf(`
<h1>Welcome to Go - Test Email!</h1>
<p>Hi %s,</p>
<p>Thank you for signing up. Please verify your email by clicking the link below:</p>
<p><a href="%s/v1/verify_email?email_id=%d&secret_code=%s">Verify Email</a></p>
<p>This link will expire in 15 minutes.</p>
`, user.Username, myUrl, verifyEmail.ID, verifyEmail.SecretCode)
	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	log.Info().
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("username", user.Username).
		Str("email", user.Email).
		Msg("processed task")
	return nil
}

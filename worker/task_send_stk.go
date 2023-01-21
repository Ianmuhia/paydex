package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"paydex/mpesa"
	"time"

	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
	"golang.org/x/exp/slog"
)

const TaskSendSTK = "task:send_stk"

type STKRequest struct {
	Amount      string
	Description string
	PhoneNumber string
}

func (distributor *RedisTaskDistributor) DistributeTaskSendSTKPush(
	ctx context.Context,
	payload *STKRequest,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendSTK, jsonPayload, opts...)
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}
	slog.Info("enqueued task", "type", task.Type(), "payload", string(task.Payload()), "queue", info.Queue, "max_retry", info.MaxRetry)
	return nil
}

// ProcessTaskSendVerifyEmail send mail with the provided data.
func (processor *RedisTaskProcessor) ProcessTaskSendSTKPush(ctx context.Context, task *asynq.Task) error {
	var payload STKRequest
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	val := mpesa.StKPushRequestBody{
		BusinessShortCode: processor.c.Mpesa.ShortCode,
		Amount:            payload.Amount,
		PhoneNumber:       payload.PhoneNumber,
		CallBackURL:       processor.c.Mpesa.CallbackURL,
		AccountReference:  processor.c.Mpesa.BusinessName,
		TransactionDesc:   payload.Description,
	}
	ct, cancelFunc := context.WithTimeout(ctx, 10*time.Second)
	defer cancelFunc()
	data, err := processor.mpesa.StkPushRequest(ct, val)
	if err != nil {
		log.Print(err)
		return errors.Wrap(asynq.SkipRetry, "MpesaService.MpesaPay")
	}

	if data.ResponseCode != "0" {
		return errors.Wrap(asynq.SkipRetry, "MpesaService.MpesaPay")
	}
	slog.Info("processed task", "type", task.Type(), "payload", string(task.Payload()))
	return nil
}

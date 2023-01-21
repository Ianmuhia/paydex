package worker

import (
	"context"
	"paydex/config"
	"paydex/mpesa"

	"time"

	"github.com/hibiken/asynq"
	"golang.org/x/exp/slog"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type TaskProcessor interface {
	Start() error
	StartScheduler() error
	ProcessTaskSendSTKPush(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server   *asynq.Server
	mpesa    *mpesa.Mpesa
	c        *config.Config
	redisOpt asynq.RedisClientOpt
}

func NewRedisTaskProcessor(
	redisOpt asynq.RedisClientOpt,
	c *config.Config,

) TaskProcessor {
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				slog.Error("process task failed", err, "type", task.Type(), "payload", task.Payload())
			}),
			Logger: NewLogger(),
		},
	)

	client := mpesa.New(
		c.Mpesa.ConsumerKey,
		c.Mpesa.ConsumerSecret,
		mpesa.WithTimeout(10*time.Second),
		mpesa.WithCache(true),
		mpesa.WithPassKey(c.Mpesa.PassKey),
		mpesa.WithB2CShortCode(c.Mpesa.ShortCode),
	)
	return &RedisTaskProcessor{
		server: server,
		mpesa:  client,
		c:      c,
	}
}

func (processor *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskSendSTK, processor.ProcessTaskSendSTKPush)

	return processor.server.Start(mux)
}

func (processor *RedisTaskProcessor) StartScheduler() error {
	l, err := time.LoadLocation("Africa/Nairobi")
	if err != nil {
		return err
	}
	mux := asynq.NewScheduler(processor.redisOpt, &asynq.SchedulerOpts{
		Location: l,
	})
	// entry, err := mux.Register("@every 1m", asynq.NewTask(TaskUpdatePayment, nil))
	// if err != nil {
	// 	return err
	// }
	// log.Printf("task scheduled [%s]", entry)
	return mux.Run()
}

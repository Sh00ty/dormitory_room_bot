package main

import (
	"context"
	"os"
	"time"

	"github.com/Sh00ty/dormitory_room_bot/internal/config"
	tgapi "github.com/Sh00ty/dormitory_room_bot/internal/infra/frameworks/tg"
	creditsRepo "github.com/Sh00ty/dormitory_room_bot/internal/infra/repositories/credits"
	listsRepo "github.com/Sh00ty/dormitory_room_bot/internal/infra/repositories/lists"
	taskRepo "github.com/Sh00ty/dormitory_room_bot/internal/infra/repositories/tasks"
	userServiceRepo "github.com/Sh00ty/dormitory_room_bot/internal/infra/repositories/users"
	"github.com/Sh00ty/dormitory_room_bot/internal/logger"
	_ "github.com/Sh00ty/dormitory_room_bot/internal/metrics"
	recallerlib "github.com/Sh00ty/dormitory_room_bot/internal/recaller"
	pgxbalancer "github.com/Sh00ty/dormitory_room_bot/internal/transaction_balancer"
	credits "github.com/Sh00ty/dormitory_room_bot/internal/usecases/credits"
	"github.com/Sh00ty/dormitory_room_bot/internal/usecases/lists"
	tasks "github.com/Sh00ty/dormitory_room_bot/internal/usecases/tasks"
	users "github.com/Sh00ty/dormitory_room_bot/internal/usecases/users"
	pool "github.com/Sh00ty/dormitory_room_bot/internal/worker_pool"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	environment  = "stg"
	unsentMsgKey = "unsent"
)

func main() {
	// ждем пока поднимется вся инфра
	time.Sleep(5 * time.Second)
	ctx := context.Background()
	err := config.Init(environment)
	if err != nil {
		logger.Fataf(err.Error())
	}

	balancer, err := pgxbalancer.New(ctx,
		config.Get(config.PgHostConfigKey).String(),
		config.Get(config.PgPasswordConfigKey).String(),
		config.Get(config.PgDbNameConfigKey).String(),
		config.Get(config.PgUserConfigKey).String(),
		config.Get(config.PgPortConfigKey).Uint16())
	if err != nil {
		logger.Fataf("can't connect to postgres : %v", err)
	}

	listRepo, mongoDisconnect, err := listsRepo.NewListsRepository(ctx,
		config.Get(config.MgHostConfigKey).String(),
		config.Get(config.MgDbNameConfigKey).String(),
		config.Get(config.MgPasswordConfigKey).String(),
		config.Get(config.MgUserConfigKey).String(),
	)
	if err != nil {
		logger.Fataf("can't connect to mongo : %v", err)
	}
	defer func() {
		err2 := mongoDisconnect(ctx)
		if err2 != nil {
			logger.Errorf("can't disconnect from mongo : %v", err)
		}
	}()

	taskManager := tasks.NewUsecase(taskRepo.NewRepo(balancer))
	userSvc := users.New(userServiceRepo.NewRepo(balancer))
	credisSvc := credits.NewUseCase(creditsRepo.NewDebitRepository(balancer), userSvc)
	listManager := lists.NewListManager(listRepo)

	shedInterval := config.Get(config.TaskSheduleTimeConfigKey).Uint64()

	bot := tgapi.Init(config.Get(config.TokenBotConfigKey).String(),
		userSvc,
		credisSvc,
		credits.NewDomestosResolver(),
		taskManager,
		listManager)

	closer := tgapi.GetNotifications(time.Duration(shedInterval) * time.Second)
	defer closer()

	recaller, err := recallerlib.Init[tgbotapi.MessageConfig](ctx,
		bot,
		config.Get(config.RecallerSearchTimeoutConfigKey).Duration()*time.Second,
		config.Get(config.RecallerBaseTimeoutConfigKey).Duration()*time.Millisecond,
		config.Get(config.RecallLimitConfigKey).Uint(),
		recallerlib.WithRedis[tgbotapi.MessageConfig](config.Get(config.RedisHostConfigKey).String(),
			config.Get(config.RedisPasswordConfigKey).String(), unsentMsgKey),
		recallerlib.WithLinear[tgbotapi.MessageConfig](),
		recallerlib.WithDeadChannel[tgbotapi.MessageConfig](10),
	)

	if err != nil {
		logger.Fataf("can't init recaller : %v", err)
	}

	go func() {
		for msg := range recaller.GetDeadChan() {
			logger.Errorf("[DEAD]: message %s dead", msg.Text)
		}
	}()

	tgPool := pool.CreateWorkerPool(
		config.Get(config.WokerPoolBufferSizeConfigKey).Uint64(),
		config.Get(config.WokerPoolMaxWorkerCountConfigKey).Uint64(),
		time.Minute,
		pool.WithRoundNRobin[tgbotapi.MessageConfig](),
	)

	go grasefulShutdown()

	tgapi.MessageProcceser(tgPool, recaller, 0)
}

func grasefulShutdown() {
	logger.Infof("---- app started -----")
	signalChan := make(chan os.Signal, 1)
	<-signalChan
	tgapi.Shutdown()
}

package main

import (
	"context"
	"os"
	"time"

	"gitlab.com/Sh00ty/dormitory_room_bot/internal/config"
	credits "gitlab.com/Sh00ty/dormitory_room_bot/internal/domain/credits"
	creditsRepo "gitlab.com/Sh00ty/dormitory_room_bot/internal/domain/credits/infra/repo_postgres"
	creditsTg "gitlab.com/Sh00ty/dormitory_room_bot/internal/domain/credits/infra/tg"
	"gitlab.com/Sh00ty/dormitory_room_bot/internal/domain/lists"
	listsRepo "gitlab.com/Sh00ty/dormitory_room_bot/internal/domain/lists/infra/repo_mongo"
	listsTg "gitlab.com/Sh00ty/dormitory_room_bot/internal/domain/lists/infra/tg"
	tasks "gitlab.com/Sh00ty/dormitory_room_bot/internal/domain/tasks"
	taskRepo "gitlab.com/Sh00ty/dormitory_room_bot/internal/domain/tasks/infra/repo_postgres"
	taskTg "gitlab.com/Sh00ty/dormitory_room_bot/internal/domain/tasks/infra/tg"
	users "gitlab.com/Sh00ty/dormitory_room_bot/internal/domain/users"
	userRepo "gitlab.com/Sh00ty/dormitory_room_bot/internal/domain/users/repo_postgres"
	userTg "gitlab.com/Sh00ty/dormitory_room_bot/internal/domain/users/tg"
	tgapi "gitlab.com/Sh00ty/dormitory_room_bot/internal/infra/frameworks/tg"
	_ "gitlab.com/Sh00ty/dormitory_room_bot/internal/metrics"
	"gitlab.com/Sh00ty/dormitory_room_bot/pkg/logger"
	"gitlab.com/Sh00ty/dormitory_room_bot/pkg/pgxbalancer"
	tgproc "gitlab.com/Sh00ty/dormitory_room_bot/pkg/tgproc"
)

func main() {
	ctx := context.Background()

	err := config.Init(os.Getenv("dormitory_env"))
	if err != nil {
		logger.Fataf(err.Error())
	}

	balancer, err := pgxbalancer.New(ctx,
		config.Get(config.PgHostConfigKey).String(),
		os.Getenv("pg_pass"),
		config.Get(config.PgDbNameConfigKey).String(),
		config.Get(config.PgUserConfigKey).String(),
		config.Get(config.PgPortConfigKey).Uint16())
	if err != nil {
		logger.Fataf("can't connect to postgres : %v", err)
	}

	listRepo, mongoDisconnect, err := listsRepo.NewListsRepository(ctx,
		config.Get(config.MgHostConfigKey).String(),
		config.Get(config.MgDbNameConfigKey).String(),
		os.Getenv("mongo_pass"),
		config.Get(config.MgUserConfigKey).String(),
	)
	if err != nil {
		logger.Fataf("can't connect to mongo : %v", err)
	}

	bot, err := tgproc.NewBot(
		os.Getenv("bot_token"),
		0,
		tgproc.WithResend(
			config.Get(config.RecallLimitConfigKey).Uint(),
			config.Get(config.RecallerBaseTimeoutConfigKey).Duration(),
		),
		tgproc.WithWorlerPool(
			config.Get(config.WokerPoolMaxWorkerCountConfigKey).Uint64(),
			config.Get(config.WokerPoolBufferSizeConfigKey).Uint64(),
			time.Minute,
		),
	)
	if err != nil {
		err2 := mongoDisconnect(ctx)
		if err2 != nil {
			logger.Errorf("failed to disconnect from mongodb: %v", err2)
		}
		logger.Fataf("cat't start tg bot: %v", err)
	}

	// others
	bot.MessageHandleFunc("help", tgproc.MessageHandleFunc(tgapi.Help))

	// tasks
	taskManager := tasks.NewUsecase(taskRepo.NewRepo(balancer))
	taskBot := taskTg.NewTgbot(taskManager, bot.GetMessageSender())
	shedInterval := config.Get(config.TaskSheduleTimeConfigKey).Uint64()
	closer := taskBot.GetNotifications(time.Duration(shedInterval) * time.Second)
	// message handlers
	bot.MessageHandleFunc("task", tgproc.MessageHandleFunc(taskBot.CreateDefaultTask))
	bot.MessageHandleFunc("subt", tgproc.MessageHandleFunc(taskBot.CreateSubsTask))
	bot.MessageHandleFunc("moment", tgproc.MessageHandleFunc(taskBot.CreateOneShotTask))
	bot.MessageHandleFunc("tasks", tgproc.MessageHandleFunc(taskBot.GetAllTasks))
	// button handlers
	bot.ButtonHandleFunc("subt", tgproc.ButtonHandleFunc(taskBot.Subscribe))
	bot.ButtonHandleFunc("unsubt", tgproc.ButtonHandleFunc(taskBot.UnSubscribe))
	bot.ButtonHandleFunc("get_t", tgproc.ButtonHandleFunc(taskBot.GetTask))
	bot.ButtonHandleFunc("change_w", tgproc.ButtonHandleFunc(taskBot.ChangeWorker))
	bot.ButtonHandleFunc("del_t", tgproc.ButtonHandleFunc(taskBot.DeleteTask))

	// list
	listManager := lists.NewListManager(listRepo)
	listBot := listsTg.New(listManager)
	// message handlers
	bot.MessageHandleFunc("lists", tgproc.MessageHandleFunc(listBot.GetAllChannelLists))
	bot.MessageHandleFunc("createl", tgproc.MessageHandleFunc(listBot.CreateList))
	bot.MessageHandleFunc("delit", tgproc.MessageHandleFunc(listBot.DeleteItem))
	bot.MessageHandleFunc("addit", tgproc.MessageHandleFunc(listBot.AddItem))
	bot.MessageHandleFunc("randit", tgproc.MessageHandleFunc(listBot.GetRandomItem))
	// button handlers
	bot.ButtonHandleFunc("dell", tgproc.ButtonHandleFunc(listBot.DeleteList))
	bot.ButtonHandleFunc("getl", tgproc.ButtonHandleFunc(listBot.GetList))
	bot.ButtonHandleFunc("randit", tgproc.ButtonHandleFunc(listBot.GetRandomItemQuery))

	// credits + users
	userSvc := users.New(userRepo.NewRepo(balancer))
	credisSvc := credits.NewUseCase(creditsRepo.NewDebitRepository(balancer), userSvc, credits.NewDomestosResolver())
	userBot := userTg.New(userSvc, credisSvc)
	bot.MessageHandleFunc("register", tgproc.MessageHandleFunc(userBot.Register))

	creditBot := creditsTg.New(credisSvc)
	bot.MessageHandleFunc("credit", tgproc.MessageHandleFunc(creditBot.Owe))
	bot.MessageHandleFunc("cancel", tgproc.MessageHandleFunc(creditBot.OweCancel))
	bot.MessageHandleFunc("checkout", tgproc.MessageHandleFunc(creditBot.Checkout))
	bot.MessageHandleFunc("bank", tgproc.MessageHandleFunc(creditBot.GetResults))

	go func() {
		logger.Infof("---- app started -----")
		signalChan := make(chan os.Signal, 1)
		<-signalChan
		bot.Stop()
		closer()
		err2 := mongoDisconnect(ctx)
		if err2 != nil {
			logger.Errorf("can't disconnect from mongo : %v", err)
		}
	}()

	bot.Start()
}

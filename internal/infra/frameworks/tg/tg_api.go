package tg

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	credits "github.com/Sh00ty/dormitory_room_bot/internal/interfaces/usecases/credits"
	listManager "github.com/Sh00ty/dormitory_room_bot/internal/interfaces/usecases/lists"
	taskManager "github.com/Sh00ty/dormitory_room_bot/internal/interfaces/usecases/tasks"
	userService "github.com/Sh00ty/dormitory_room_bot/internal/interfaces/usecases/users"
	localerrors "github.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	"github.com/Sh00ty/dormitory_room_bot/internal/logger"
	recallerlib "github.com/Sh00ty/dormitory_room_bot/internal/recaller"
	pool "github.com/Sh00ty/dormitory_room_bot/internal/worker_pool"
	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type tgbot struct {
	bot            *tgbotapi.BotAPI
	userSvc        userService.UserService
	debitSvc       credits.UseCaseInterface
	debitReolver   credits.Resolver
	taskManager    taskManager.Usecase
	listManager    listManager.ListManager
	unsentMessages uint64
}

var (
	bot *tgbot

	recaller recallerlib.Recaller[tgbotapi.MessageConfig]

	handlers = map[string]botJob{
		// managment commands
		"register": {
			f:           Register,
			helpMessage: "/register регистрирует тебя в канале в котором ты это напишешь",
		},

		// task commands
		"task": {
			f: CreateDefaultTask,
			helpMessage: "/task some_task @w1 @w2 @w3 2 2h 04-11-2022+21:08 большое описание\nСоздает задачку где:\n\n" +
				"Все опции указанные ниже не обязательные, их можно писать, а можно и нет(но порядок в котором идут опции оч важен)\n\n" +
				"@w1 ... это возможные исполнители данной задачи\n" +
				"следующее число показывает сколько человек в моменте исполняет задачу\n" +
				"исполнителей можно поменять или прокрутить по циклу если их количество равно цифре после них(лучше попробовать чтобы понять)\n" +
				"далее идет интервал напоминания этой задачи, можно использовать 40s, 20m, 24h\n" +
				"04-11-2022+21:08 - точные дата и время в которое бот напомнит о задаче\n" +
				"после срабатывания точного времени начнет работать интервал\n" +
				"и наконец сколь угодно большое описание",
		},
		"moment": {
			f:           CreateOneShotTask,
			helpMessage: "/moment some_task @w1 @w2 @w3 2 2h 04-11-2022+21:08 большое описание\nСоздает задачку которая исчезает после напоминания",
		},
		"subt": {
			f:           CreateSubsTask,
			helpMessage: "/moment some_task @w1 @w2 @w3 2 2h 04-11-2022+21:08 большое описание\nСоздает задачку на которую можно подписываться и отписываться",
		},
		"sub": {
			f:           Subscribe,
			helpMessage: "/sub...",
		},
		"unsub": {
			f:           UnSubscribe,
			helpMessage: "/unsub...",
		},
		"get": {
			f:           GetTask,
			helpMessage: "/get some_task - получает задачку по ее id",
		},
		"change": {
			f:           ChangeWorker,
			helpMessage: "/change some_task - меняет исполнителей для данной задачи",
		},
		"delete": {
			f:           DeleteTask,
			helpMessage: "/delete some_task - удаляет задачку по ее id",
		},
		"tasks": {
			f:           GetAllTasks,
			helpMessage: "/tasks - выводит все имеющиеся в этом чате задачи",
		},

		// money commands
		"bank": {
			f:           GetResults,
			helpMessage: "/bank - выводит баланс каждого участиника чата\nЕлси баланс меньше нуля значит что ты кому-то дожен, если больше то наоборот",
		},
		"credit": {
			f:           Owe,
			helpMessage: "/credit @gera @steps @etc 3000\nГоворит что исполнители после /credit должны тому кто пишет 3000/(их количество)",
		},
		"checkout": {
			f:           Checkout,
			helpMessage: "/checkout - показывает кто кому сколько должен перевести внутри данного чата. После этого обнуляет все имеющиеся долги в этом чате",
		},

		// lists commands
		"createl": {
			f:           CreateList,
			helpMessage: "/lcreate list_id item1 item2 .. - создает список",
		},
		"addit": {
			f:           AddItem,
			helpMessage: "/addit list_id item - добавляет элемент в список",
		},
		"getl": {
			f:           GetList,
			helpMessage: "/getl list_id - выводит список",
		},
		"dell": {
			f:           DeleteList,
			helpMessage: "/dell list_id - удаляет список с заданным id",
		},
		"delit": {
			f:           DeleteItem,
			helpMessage: "/delit list_id index - удаляет из списка элемент с заданным номером",
		},
		"lists": {
			f:           GetAllChannelLists,
			helpMessage: "/lists - выводит все имеющиеся списки",
		},
		"randit": {
			f:           GetRandomItem,
			helpMessage: "/randit list_id (количество опционально)- выводит случайный элемент из списка",
		},
	}
)

func Init(token string, u userService.UserService, c credits.UseCaseInterface, dr credits.Resolver, tm taskManager.Usecase, lm listManager.ListManager) *tgbot {
	// тк help сообщение нельзя добавить нарямую в handlers
	handlers["help"] = botJob{f: Help}

	bot = &tgbot{
		userSvc:      u,
		debitSvc:     c,
		debitReolver: dr,
		taskManager:  tm,
		listManager:  lm,
	}
	b, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		logger.Fataf("can't register tgbot api client: %s", err)
	}
	bot.bot = b
	return bot
}

func Shutdown() {
	bot.bot.StopReceivingUpdates()
}

func MessageProcceser(pool pool.Pool[tgbotapi.MessageConfig], rec recallerlib.Recaller[tgbotapi.MessageConfig], updateTimeout int) {
	recaller = rec
	go func() {
		resChan := pool.GetResult()
		for res := range resChan {
			//логируем какие-то результаты
			logger.Infof("sended message to %v", res.ChatID)
		}
	}()

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = updateTimeout

	updates := bot.bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		if update.Message != nil {
			job, err := commandParser(update.Message)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
				msg.ParseMode = tgbotapi.ModeHTML
				_, err = bot.bot.Send(msg)
				if err != nil {
					logger.Errorf("can't send message %v", err)
				}
				continue
			}
			if job == nil {
				continue
			}
			worker := pool.Put(job, update.Message.Chat.ID)
			logger.Infof("worker %v", worker)
		} else if update.Poll != nil {
			logger.Info("poll")
		} else if update.CallbackQuery != nil {
			job, err := callbackQueryParser(update.CallbackQuery)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, err.Error())
				msg.ParseMode = tgbotapi.ModeHTML
				_, err = bot.bot.Send(msg)
				if err != nil {
					logger.Errorf("can't send message %v", err)
				}
				continue
			}
			worker := pool.Put(job, update.CallbackQuery.Message.Chat.ID)
			logger.Infof("worker %v", worker)
		}
	}
	pool.Close()
}

type botJob struct {
	tgMessage   *tgbotapi.Message
	f           func(ctx context.Context, tgMessage *tgbotapi.Message) tgbotapi.MessageConfig
	helpMessage string
}

// Call - функция для отправки сообщений(называется так из-за реколлера)
func (b tgbot) Call(ctx context.Context, tgMessage tgbotapi.MessageConfig) error {
	tgMessage.ParseMode = tgbotapi.ModeHTML
	_, err := b.bot.Send(tgMessage)
	return err
}

func (j botJob) Exec(ctx context.Context) tgbotapi.MessageConfig {
	msg := j.f(ctx, j.tgMessage)
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := bot.bot.Send(msg)
	if err != nil {
		logger.Errorf("can't send message: %v", err)
		atomic.AddUint64(&bot.unsentMessages, 1)
		err2 := recaller.SaveForReccal(ctx, strconv.FormatUint(bot.unsentMessages, 10), msg)
		if err2 != nil {
			logger.Errorf("failed to save for reccal %s with err=%v; cause=%v", msg, err, err2)
			return msg
		}
	}
	return msg
}

func callbackQueryParser(q *tgbotapi.CallbackQuery) (pool.Job[tgbotapi.MessageConfig], error) {
	comm, payload, _ := strings.Cut(q.Data, ":")

	logger.Infof("button query is : %s : by %s", q.Data, q.From.UserName)

	job, exists := handlers[comm]
	if !exists {
		return nil, fmt.Errorf("❌ эти кнопки сломаны или они уже не поддерживаюся ботом")
	}

	msg := q.Message
	msg.From = q.From
	msg.Entities = []tgbotapi.MessageEntity{{Type: "bot_command"}}
	// добавляем / чтобы он распарсил как команду
	msg.Text = "/" + payload
	job.tgMessage = msg
	return job, nil
}

func commandParser(msg *tgbotapi.Message) (pool.Job[tgbotapi.MessageConfig], error) {
	if !msg.IsCommand() {
		return nil, nil
	}
	command := msg.Command()

	logger.Infof("command is %s", command)

	job, exists := handlers[command]
	if !exists {
		return botJob{}, fmt.Errorf("❌ команда %s не зарегистрирована", command)
	}
	job.tgMessage = msg
	return job, nil
}

func Register(ctx context.Context, tgMessage *tgbotapi.Message) tgbotapi.MessageConfig {
	usr := entities.User{
		ID:        valueObjects.UserID("@" + tgMessage.From.UserName),
		Credit:    0,
		UserName:  "@" + tgMessage.From.UserName,
		ChannelID: valueObjects.ChannelID(tgMessage.Chat.ID),
	}

	err := bot.userSvc.Create(ctx, usr)
	if err != nil {
		if !errors.Is(err, localerrors.ErrAlreadyExists) {
			logger.Errorf("Register: %v", err)
			return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ пороизошла ошибка во время регистрации, повтори позже")
		}
	}

	err = bot.debitSvc.Create(ctx, usr.ChannelID, usr.ID, valueObjects.Money(usr.Credit))
	if err != nil {
		if errors.Is(err, localerrors.ErrAlreadyExists) {
			nmsg := tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ пользователь %s уже зарегистрирован в данном чате", usr.UserName))
			return nmsg
		}
		logger.Errorf("Register: %v", err)
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ пороизошла ошибка по время регистрации, повтори позже")
	}
	return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("✅ зарегистрировал: %s", usr.UserName))
}

func Help(ctx context.Context, msg *tgbotapi.Message) (res tgbotapi.MessageConfig) {
	resMsg := "По появившемя вопросам пиши <i>tw02h00ty</i>\n"
	for _, handler := range handlers {
		resMsg += "\n" + handler.helpMessage + "\n"
	}
	return tgbotapi.NewMessage(msg.Chat.ID, resMsg)
}

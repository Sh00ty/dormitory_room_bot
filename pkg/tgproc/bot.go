package tgproc

import (
	"context"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.com/Sh00ty/dormitory_room_bot/pkg/workerpool"
)

type (
	Bot struct {
		bot               *tgbotapi.BotAPI
		updateTimeout     int
		pool              workerpool.Pool[Messages]
		resendLimit       uint
		resendBaseTimeout time.Duration
		mux               tgMux
		logger            BotLogger
	}

	Option func(b *Bot)

	BotLogger interface {
		Errorf(message string, args ...interface{})
	}

	updateHandler struct {
		messageHandler MessageHandler
		buttonHandler  ButtonHandler
		tgUpdate       tgbotapi.Update
	}

	Messages []tgbotapi.MessageConfig

	MessageHandler interface {
		HandleMessage(ctx context.Context, tgMessage *tgbotapi.Message) Messages
	}

	ButtonHandler interface {
		HandleQuery(ctx context.Context, callBackQuery *tgbotapi.CallbackQuery) Messages
	}

	TgUser int64

	MessageSender interface {
		Send(tgbotapi.Chattable) (tgbotapi.Message, error)
	}
)

func NewBot(token string, updateTimeout int, options ...Option) (Bot, error) {
	bot := Bot{
		mux: tgMux{
			messageHandlers: make(map[string]MessageHandler),
			buttonHandlers:  make(map[string]ButtonHandler),
		},
	}
	apiBot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return bot, err
	}
	bot.bot = apiBot
	if bot.pool == nil {
		bot.pool = workerpool.Create(
			150,
			15,
			3*time.Second,
			workerpool.WithRoundNRobin[Messages](),
		)
	}
	if bot.resendLimit == 0 {
		bot.resendLimit = 1
	}
	if bot.logger == nil {
		bot.logger = newLogger()
	}
	return bot, nil
}

func (b *Bot) MessageHandleFunc(command string, handler MessageHandler) {
	b.mux.commandHandleFunc(command, handler)
}

func (b *Bot) ButtonHandleFunc(pattern string, handler ButtonHandler) {
	b.mux.buttonHandleFunc(pattern, handler)
}

func (j updateHandler) Exec(ctx context.Context) Messages {
	switch {
	case j.buttonHandler != nil:
		return j.buttonHandler.HandleQuery(ctx, j.tgUpdate.CallbackQuery)
	case j.messageHandler != nil:
		return j.messageHandler.HandleMessage(ctx, j.tgUpdate.Message)
	default:
		return Messages{tgbotapi.NewMessage(j.tgUpdate.FromChat().ChatConfig().ChatID, "у 2hooty кривые руки")}
	}
}

func (b Bot) Start() {
	go func() {
		for msgs := range b.pool.GetResult() {
			for _, msg := range msgs {
				for i := 0; i < int(b.resendLimit); i++ {
					_, err := b.bot.Send(msg)
					if err != nil {
						b.logger.Errorf("can't send error due to err=%v", err)
						time.Sleep(calculateTimeout(b.resendBaseTimeout, i))
						continue
					}
					break
				}
			}
		}
	}()

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = b.updateTimeout

	updates := b.bot.GetUpdatesChan(updateConfig)
	ctx := context.Background()
	for update := range updates {
		h, err := b.mux.getUpdateHandler(ctx, update)
		if err != nil {
			msg := tgbotapi.NewMessage(update.FromChat().ID, err.Error())
			_, err := b.bot.Send(msg)
			if err != nil {
				b.logger.Errorf("can't send error due to err=%v", err)
			}
		}
		if h == nil {
			continue
		}
		b.pool.Put(h, update.FromChat().ID)
	}
}

func (b *Bot) Stop() {
	b.bot.StopReceivingUpdates()
	b.pool.Close()
}

func (b Bot) GetMessageSender() MessageSender {
	return b.bot
}

func calculateTimeout(baseTimeout time.Duration, resendNum int) time.Duration {
	return baseTimeout * time.Duration(resendNum+1)
}

func MakeMessage(chatID int64, msgs ...string) Messages {
	res := make([]tgbotapi.MessageConfig, 0, len(msgs))
	for _, msg := range msgs {
		tgMsg := tgbotapi.NewMessage(chatID, msg)
		tgMsg.ParseMode = tgbotapi.ModeHTML
		res = append(res, tgMsg)
	}
	return res
}

type (
	MessageHandleFunc func(ctx context.Context, tgMessage *tgbotapi.Message) []tgbotapi.MessageConfig
	ButtonHandleFunc  func(ctx context.Context, tgMessage *tgbotapi.CallbackQuery) []tgbotapi.MessageConfig
)

func (f MessageHandleFunc) HandleMessage(ctx context.Context, tgMessage *tgbotapi.Message) Messages {
	return f(ctx, tgMessage)
}

func (f ButtonHandleFunc) HandleQuery(ctx context.Context, query *tgbotapi.CallbackQuery) Messages {
	return f(ctx, query)
}

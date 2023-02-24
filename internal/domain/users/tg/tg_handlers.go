package tg

import (
	"context"
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.com/Sh00ty/dormitory_room_bot/internal/domain/credits"
	user "gitlab.com/Sh00ty/dormitory_room_bot/internal/domain/users"
	"gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"
	localerrors "gitlab.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
	"gitlab.com/Sh00ty/dormitory_room_bot/pkg/logger"
	"gitlab.com/Sh00ty/dormitory_room_bot/pkg/tgproc"
)

type tgbot struct {
	userSvc       user.UserService
	creditManager credits.CreditManager
}

func New(uu user.UserService, cu credits.CreditManager) *tgbot {
	return &tgbot{userSvc: uu, creditManager: cu}
}

func (bot tgbot) Register(ctx context.Context, tgMessage *tgbotapi.Message) []tgbotapi.MessageConfig {
	usr := entities.User{
		ID:        valueObjects.UserID("@" + tgMessage.From.UserName),
		Credit:    0,
		UserName:  "@" + tgMessage.From.UserName,
		ChannelID: valueObjects.ChannelID(tgMessage.Chat.ID),
	}
	if err := bot.userSvc.Create(ctx, usr); err != nil {
		if !errors.Is(err, localerrors.ErrAlreadyExists) {
			return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ пороизошла ошибка во время регистрации, повтори позже")
		}
	}

	if err := bot.creditManager.Create(ctx, usr.ChannelID, usr.ID, valueObjects.Money(usr.Credit)); err != nil {
		if errors.Is(err, localerrors.ErrAlreadyExists) {
			return tgproc.MakeMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ пользователь %s уже зарегистрирован в данном чате", usr.UserName))
		}
		logger.Errorf("Register: %v", err)
		return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ пороизошла ошибка по время регистрации, повтори позже")
	}
	return tgproc.MakeMessage(tgMessage.Chat.ID, fmt.Sprintf("✅ зарегистрировал: %s", usr.UserName))
}

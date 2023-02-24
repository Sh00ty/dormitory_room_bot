package tgproc

import (
	"context"
	"errors"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	errUnknownCommand               = errors.New("неизвестная команда")
	errIncorrectEncodedCallBackData = errors.New("неправильно закодированная полезная нагрузка для кнопки")
)

type tgMux struct {
	messageHandlers map[string]MessageHandler
	buttonHandlers  map[string]ButtonHandler
}

func (m *tgMux) buttonHandleFunc(pattern string, handler ButtonHandler) {
	m.buttonHandlers[pattern] = handler
}

func (m *tgMux) commandHandleFunc(command string, handler MessageHandler) {
	m.messageHandlers[command] = handler
}

func (m tgMux) getUpdateHandler(ctx context.Context, update tgbotapi.Update) (*updateHandler, error) {
	switch {
	case update.Message != nil:
		if !update.Message.IsCommand() {
			return nil, nil
		}
		handler, exists := m.messageHandlers[update.Message.Command()]
		if !exists {
			return nil, errUnknownCommand
		}
		return &updateHandler{
			messageHandler: handler,
			tgUpdate:       update,
		}, nil
	case update.CallbackQuery != nil:
		handlerName, data, exists := strings.Cut(update.CallbackQuery.Data, ":")
		if !exists {
			return nil, errIncorrectEncodedCallBackData
		}

		handler, exists := m.buttonHandlers[handlerName]
		if !exists {
			return nil, errUnknownCommand
		}
		update.CallbackQuery.Data = data
		return &updateHandler{
			buttonHandler: handler,
			tgUpdate:      update,
		}, nil
	default:
		return nil, nil
	}
}

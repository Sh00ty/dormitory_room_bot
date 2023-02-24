package lists

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.com/Sh00ty/dormitory_room_bot/internal/domain/lists"
	"gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"
	localerrors "gitlab.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	metrics "gitlab.com/Sh00ty/dormitory_room_bot/internal/metrics"
	valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
	"gitlab.com/Sh00ty/dormitory_room_bot/pkg/logger"
	"gitlab.com/Sh00ty/dormitory_room_bot/pkg/tgproc"
)

type tgbot struct {
	listManager lists.ListManager
}

func New(l lists.ListManager) *tgbot {
	return &tgbot{l}
}

func (bot tgbot) CreateList(ctx context.Context, tgMessage *tgbotapi.Message) []tgbotapi.MessageConfig {

	if tgMessage.CommandArguments() == "" {
		return tgproc.MakeMessage(tgMessage.Chat.ID, "/createl list_id")
	}
	channelID := valueObjects.ChannelID(tgMessage.Chat.ID)

	list, err := listParser(tgMessage)
	if err != nil {
		// тут ошибки вроде читаемые
		return tgproc.MakeMessage(tgMessage.Chat.ID, err.Error())
	}
	if list.ID == "" {
		return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ нельзя сделать пустой список")
	}

	err = bot.listManager.CreateList(ctx, channelID, list)
	if err != nil {
		if errors.Is(err, localerrors.ErrAlreadyExists) {
			return tgproc.MakeMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ список <i>%s</i> уже существует", list.ID))
		}
		logger.Errorf("CreateList: %v", err)
		return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время создания списка, повторите позже")
	}

	metrics.TotalListCommands.WithLabelValues("/createl").Add(1)
	return tgproc.MakeMessage(tgMessage.Chat.ID, fmt.Sprintf("✅ список <i>%s</i> создан", list.ID))
}

func (bot tgbot) GetList(ctx context.Context, query *tgbotapi.CallbackQuery) []tgbotapi.MessageConfig {
	listID := valueObjects.ListID(query.Data)
	channelID := query.Message.Chat.ID

	if listID == "" {
		return tgproc.MakeMessage(channelID, "/getl list_id")
	}

	list, err := bot.listManager.GetList(ctx, valueObjects.ChannelID(channelID), listID)
	if err != nil {
		if errors.Is(err, localerrors.ErrDoesntExist) {
			return tgproc.MakeMessage(channelID, fmt.Sprintf("❌ списка под названием <i>%s</i> не существует", listID))
		}
		logger.Errorf("GetList: %v", err)
		return tgproc.MakeMessage(channelID, "❌ произошла ошибка во время получания списка, повторите позже")
	}

	resMsg := fmt.Sprintf("📃 <i>Список:</i>  <b>%s</b>\n", list.ID)
	for i, item := range list.Items {
		resMsg += fmt.Sprintf("   %d. %s\n   🙍 by <i>%s</i>\n\n", i+1, item.Name, item.Author)
	}
	resMsg += fmt.Sprintf("🙍 показан by <i>%s</i>", query.From.UserName)

	metrics.TotalListCommands.WithLabelValues("/getl").Add(1)
	return tgproc.MakeMessage(channelID, resMsg)
}

func (bot tgbot) DeleteList(ctx context.Context, query *tgbotapi.CallbackQuery) []tgbotapi.MessageConfig {
	listID := valueObjects.ListID(query.Data)
	channelID := query.Message.Chat.ID

	if listID == "" {
		return tgproc.MakeMessage(channelID, "/dell list_id")
	}

	err := bot.listManager.DeleteList(ctx, valueObjects.ChannelID(channelID), listID)
	if err != nil {
		if errors.Is(err, localerrors.ErrDoesntExist) {
			return tgproc.MakeMessage(channelID, fmt.Sprintf("❌ списка под названием <i>%s</i> не существует", listID))
		}
		logger.Errorf("DeleteList: %v", err)
		return tgproc.MakeMessage(channelID, "❌ произошла ошибка во время удаления списка, повторите позже")
	}

	metrics.TotalListCommands.WithLabelValues("/dell").Add(1)
	return tgproc.MakeMessage(channelID, fmt.Sprintf("🗑️ лист <i>%s</i> удален by <i>%s</i>", listID, query.From.UserName))
}

func (bot tgbot) DeleteItem(ctx context.Context, tgMessage *tgbotapi.Message) (result []tgbotapi.MessageConfig) {

	if tgMessage.CommandArguments() == "" {
		return tgproc.MakeMessage(tgMessage.Chat.ID, "/delit list_id 2")
	}

	channelID := valueObjects.ChannelID(tgMessage.Chat.ID)

	args := tgMessage.CommandArguments()
	listID, itemIndStr, _ := strings.Cut(args, " ")
	if listID == "" {
		return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ не нашел id списка")
	}
	if itemIndStr == "" {
		return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ не введен индекс элемента в списке")
	}

	itemInd, err := strconv.ParseUint(itemIndStr, 10, 64)
	if err != nil {
		return tgproc.MakeMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ <i>%s</i> это не число и тем более не верный индекс элемента в списке", itemIndStr))
	}

	// в человеческом мире индексы на 1 больше
	err = bot.listManager.DeleteByIndex(ctx, channelID, valueObjects.ListID(listID), uint(itemInd-1))
	if err != nil {
		if errors.Is(err, localerrors.ErrDoesntExist) {
			return tgproc.MakeMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ неверный элемент списка или индекс элемента списка: <i>%s</i>[%d]", listID, itemInd))
		}
		logger.Errorf("DeleteItem: %v", err)
		return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время удаления элемента списка, повторите позже")
	}

	metrics.TotalListCommands.WithLabelValues("/delit").Add(1)
	return tgproc.MakeMessage(tgMessage.Chat.ID, fmt.Sprintf("🗑️ <i>%s</i>[%d] удален by <i>%s</i>", listID, itemInd, tgMessage.From.UserName))
}

func (bot tgbot) AddItem(ctx context.Context, tgMessage *tgbotapi.Message) []tgbotapi.MessageConfig {

	if tgMessage.CommandArguments() == "" {
		return tgproc.MakeMessage(tgMessage.Chat.ID, "/addit list_id some item")
	}

	channelID := valueObjects.ChannelID(tgMessage.Chat.ID)

	args := tgMessage.CommandArguments()
	listID, itemID, _ := strings.Cut(args, " ")
	if listID == "" {
		return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ не нашел <b>id</b> списка")
	}
	if itemID == "" {
		return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ нельзя сделать пустой элемент")
	}

	err := bot.listManager.AddItem(ctx, channelID, valueObjects.ListID(listID), entities.Item{
		Name:   itemID,
		Author: tgMessage.From.UserName,
	})
	if err != nil {
		if errors.Is(err, localerrors.ErrDoesntExist) {
			return tgproc.MakeMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ списка под названием <i>%s</i> не существует", listID))
		}
		logger.Errorf("AddItem: %v", err)
		return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время получания списка, повторите позже")
	}

	metrics.TotalListCommands.WithLabelValues("/addit").Add(1)
	return tgproc.MakeMessage(tgMessage.Chat.ID, fmt.Sprintf("✅ <i>%s</i> создан в списке <i>%s</i>", itemID, listID))
}

func (bot tgbot) GetAllChannelLists(ctx context.Context, tgMessage *tgbotapi.Message) []tgbotapi.MessageConfig {

	channelID := valueObjects.ChannelID(tgMessage.Chat.ID)
	lists, err := bot.listManager.GetAllChannelLists(ctx, channelID)
	if err != nil {
		if errors.Is(err, localerrors.ErrDoesntExist) {
			return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ а списков еще нет")
		}
		logger.Errorf("GetAllChannelLists: %v", err)
		return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время получания списков, повторите позже")
	}
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(lists))
	for _, list := range lists {
		gbutton := tgbotapi.NewInlineKeyboardButtonData(string(list.ID), "getl:"+string(list.ID))
		rbutton := tgbotapi.NewInlineKeyboardButtonData("🎲", "randit:"+string(list.ID))
		dbutton := tgbotapi.NewInlineKeyboardButtonData("🗑", "dell:"+string(list.ID))

		row := tgbotapi.NewInlineKeyboardRow(gbutton, rbutton, dbutton)
		rows = append(rows, row)
	}
	result := tgbotapi.NewMessage(tgMessage.Chat.ID, "и вот же они, сиписьки")
	result.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	metrics.TotalListCommands.WithLabelValues("/lists").Add(1)
	return []tgbotapi.MessageConfig{result}
}

func (bot tgbot) GetRandomItem(ctx context.Context, tgMessage *tgbotapi.Message) []tgbotapi.MessageConfig {

	if tgMessage.CommandArguments() == "" {
		return tgproc.MakeMessage(tgMessage.Chat.ID, "/randit list_id")
	}

	args := tgMessage.CommandArguments()
	listID, countStr, _ := strings.Cut(args, " ")
	if listID == "" {
		return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ не нашел id списка")
	}

	var (
		count uint64 = 1
		err   error
	)

	if countStr != "" {
		count, err = strconv.ParseUint(countStr, 10, 64)
		if err != nil {
			return tgproc.MakeMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ %s это не число, а нужно либо ничего, либо число", countStr))
		}
	}

	channelID := valueObjects.ChannelID(tgMessage.Chat.ID)
	items, err := bot.listManager.GetRandomItems(ctx, channelID, valueObjects.ListID(listID), count)
	if err != nil {
		switch {
		case errors.Is(err, localerrors.ErrDoesntExist):
			return tgproc.MakeMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ списка под названием <i>%s</i> не существует", listID))
		case errors.Is(err, localerrors.ErrNoItems):
			return tgproc.MakeMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ в <i>%s</i> списке нет еще элементов", listID))
		default:
			logger.Errorf("GetRandomItem: %v", err)
			return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время получания слйчайного элемента списка, повторите позже")
		}
	}

	resMsg := fmt.Sprintf("🎲 случайные элементы из списка <i>%s</i>:\n", listID)
	for _, item := range items {
		resMsg += fmt.Sprintf("👉 <b>%s</b>\nby <i>%s</i>\n\n", item.Name, item.Author)
	}
	resMsg += fmt.Sprintf("🙍 сгенерирован по просьбе <i>%s</i>", tgMessage.From.UserName)

	metrics.TotalListCommands.WithLabelValues("/randit").Add(1)
	return tgproc.MakeMessage(tgMessage.Chat.ID, resMsg)
}

func (bot tgbot) GetRandomItemQuery(ctx context.Context, query *tgbotapi.CallbackQuery) []tgbotapi.MessageConfig {
	channelID := query.Message.Chat.ID
	listID := query.Data

	if listID == "" {
		return tgproc.MakeMessage(channelID, "❌ не нашел id списка")
	}

	items, err := bot.listManager.GetRandomItems(ctx, valueObjects.ChannelID(channelID), valueObjects.ListID(listID), 1)
	if err != nil {
		switch {
		case errors.Is(err, localerrors.ErrDoesntExist):
			return tgproc.MakeMessage(channelID, fmt.Sprintf("❌ списка под названием <i>%s</i> не существует", listID))
		case errors.Is(err, localerrors.ErrNoItems):
			return tgproc.MakeMessage(channelID, fmt.Sprintf("❌ в <i>%s</i> списке нет еще элементов", listID))
		default:
			logger.Errorf("GetRandomItem: %v", err)
			return tgproc.MakeMessage(channelID, "❌ произошла ошибка во время получания слйчайного элемента списка, повторите позже")
		}
	}

	resMsg := fmt.Sprintf("🎲 случайные элементы из списка <i>%s</i>:\n", listID)
	for _, item := range items {
		resMsg += fmt.Sprintf("👉 <b>%s</b>\nby <i>%s</i>\n\n", item.Name, item.Author)
	}
	resMsg += fmt.Sprintf("🙍 сгенерирован по просьбе <i>%s</i>", query.From.UserName)

	metrics.TotalListCommands.WithLabelValues("/randit").Add(1)
	return tgproc.MakeMessage(channelID, resMsg)
}

func listParser(tgMessage *tgbotapi.Message) (list entities.List, err error) {
	itemsStr := strings.Split(tgMessage.CommandArguments(), " ")
	if len(itemsStr) < 1 {
		return entities.List{}, fmt.Errorf("❌ не могу получить название списка")
	}

	list.ID = valueObjects.ListID(itemsStr[0])
	list.ChannelID = valueObjects.ChannelID(tgMessage.Chat.ID)
	list.Items = make([]entities.Item, 0)
	for i, item := range itemsStr {
		if i == 0 {
			continue
		}
		list.Items = append(list.Items, entities.Item{Author: tgMessage.From.UserName, Name: item})
	}
	return
}

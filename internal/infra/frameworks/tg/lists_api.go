package tg

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	localerrors "github.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	"github.com/Sh00ty/dormitory_room_bot/internal/logger"
	metrics "github.com/Sh00ty/dormitory_room_bot/internal/metrics"
	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func CreateList(ctx context.Context, tgMessage *tgbotapi.Message) (result tgbotapi.MessageConfig) {

	if tgMessage.CommandArguments() == "" {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "/createl list_id")
	}
	channelID := valueObjects.ChannelID(tgMessage.Chat.ID)

	list, err := listParser(tgMessage)
	if err != nil {
		// тут ошибки вроде читаемые
		return tgbotapi.NewMessage(tgMessage.Chat.ID, err.Error())
	}
	if list.ID == "" {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ нельзя сделать пустой список")
	}

	err = bot.listManager.CreateList(ctx, channelID, list)
	if err != nil {
		if errors.Is(err, localerrors.ErrAlreadyExists) {
			return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ список <i>%s</i> уже существует", list.ID))
		}
		logger.Errorf("CreateList: %v", err)
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время создания списка, повторите позже")
	}

	metrics.TotalListCommands.WithLabelValues("/createl").Add(1)
	return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("✅ список <i>%s</i> создан", list.ID))
}

func GetList(ctx context.Context, tgMessage *tgbotapi.Message) (result tgbotapi.MessageConfig) {

	if tgMessage.CommandArguments() == "" {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "/getl list_id")
	}

	channelID := valueObjects.ChannelID(tgMessage.Chat.ID)
	listID := valueObjects.ListID(tgMessage.CommandArguments())
	list, err := bot.listManager.GetList(ctx, channelID, listID)
	if err != nil {
		if errors.Is(err, localerrors.ErrDoesntExist) {
			return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ списка под названием <i>%s</i> не существует", listID))
		}
		logger.Errorf("GetList: %v", err)
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время получания списка, повторите позже")
	}

	resMsg := fmt.Sprintf("📃 <i>Список:</i>  <b>%s</b>\n", list.ID)
	for i, item := range list.Items {
		resMsg += fmt.Sprintf("   %d. %s\n   🙍 by <i>%s</i>\n\n", i+1, item.Name, item.Author)
	}
	resMsg += fmt.Sprintf("🙍 показан by <i>%s</i>", tgMessage.From.UserName)

	metrics.TotalListCommands.WithLabelValues("/getl").Add(1)
	return tgbotapi.NewMessage(tgMessage.Chat.ID, resMsg)
}

func DeleteList(ctx context.Context, tgMessage *tgbotapi.Message) (result tgbotapi.MessageConfig) {

	if tgMessage.CommandArguments() == "" {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "/dell list_id")
	}

	channelID := valueObjects.ChannelID(tgMessage.Chat.ID)
	listID := valueObjects.ListID(tgMessage.CommandArguments())
	err := bot.listManager.DeleteList(ctx, channelID, listID)
	if err != nil {
		if errors.Is(err, localerrors.ErrDoesntExist) {
			return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ списка под названием <i>%s</i> не существует", listID))
		}
		logger.Errorf("DeleteList: %v", err)
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время удаления списка, повторите позже")
	}

	metrics.TotalListCommands.WithLabelValues("/dell").Add(1)
	return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("🗑️ лист <i>%s</i> удален by <i>%s</i>", listID, tgMessage.From.UserName))
}

func DeleteItem(ctx context.Context, tgMessage *tgbotapi.Message) (result tgbotapi.MessageConfig) {

	if tgMessage.CommandArguments() == "" {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "/delit list_id 2")
	}

	channelID := valueObjects.ChannelID(tgMessage.Chat.ID)

	args := tgMessage.CommandArguments()
	listID, itemIndStr, _ := strings.Cut(args, " ")
	if listID == "" {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ не нашел id списка")
	}
	if itemIndStr == "" {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ не введен индекс элемента в списке")
	}

	itemInd, err := strconv.ParseUint(itemIndStr, 10, 64)
	if err != nil {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ <i>%s</i> это не число и тем более не верный индекс элемента в списке", itemIndStr))
	}

	// в человеческом мире индексы на 1 больше
	err = bot.listManager.DeleteByIndex(ctx, channelID, valueObjects.ListID(listID), uint(itemInd-1))
	if err != nil {
		if errors.Is(err, localerrors.ErrDoesntExist) {
			return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ неверный элемент списка или индекс элемента списка: <i>%s</i>[%d]", listID, itemInd))
		}
		logger.Errorf("DeleteItem: %v", err)
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время удаления элемента списка, повторите позже")
	}

	metrics.TotalListCommands.WithLabelValues("/delit").Add(1)
	return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("🗑️ <i>%s</i>[%d] удален by <i>%s</i>", listID, itemInd, tgMessage.From.UserName))
}

func AddItem(ctx context.Context, tgMessage *tgbotapi.Message) (result tgbotapi.MessageConfig) {

	if tgMessage.CommandArguments() == "" {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "/addit list_id some item")
	}

	channelID := valueObjects.ChannelID(tgMessage.Chat.ID)

	args := tgMessage.CommandArguments()
	listID, itemID, _ := strings.Cut(args, " ")
	if listID == "" {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ не нашел <b>id</b> списка")
	}
	if itemID == "" {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ нельзя сделать пустой элемент")
	}

	err := bot.listManager.AddItem(ctx, channelID, valueObjects.ListID(listID), entities.Item{
		Name:   itemID,
		Author: tgMessage.From.UserName,
	})
	if err != nil {
		if errors.Is(err, localerrors.ErrDoesntExist) {
			return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ списка под названием <i>%s</i> не существует", listID))
		}
		logger.Errorf("AddItem: %v", err)
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время получания списка, повторите позже")
	}

	metrics.TotalListCommands.WithLabelValues("/addit").Add(1)
	return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("✅ <i>%s</i> создан в списке <i>%s</i>", itemID, listID))
}

func GetAllChannelLists(ctx context.Context, tgMessage *tgbotapi.Message) (result tgbotapi.MessageConfig) {

	channelID := valueObjects.ChannelID(tgMessage.Chat.ID)
	lists, err := bot.listManager.GetAllChannelLists(ctx, channelID)
	if err != nil {
		if errors.Is(err, localerrors.ErrDoesntExist) {
			return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ а списков еще нет")
		}
		logger.Errorf("GetAllChannelLists: %v", err)
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время получания списков, повторите позже")
	}
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(lists))
	for _, list := range lists {
		gbutton := tgbotapi.NewInlineKeyboardButtonData(string(list.ID), "getl:"+string(list.ID))
		rbutton := tgbotapi.NewInlineKeyboardButtonData("🎲", "randit:"+string(list.ID))
		dbutton := tgbotapi.NewInlineKeyboardButtonData("🗑", "dell:"+string(list.ID))

		row := tgbotapi.NewInlineKeyboardRow(gbutton, rbutton, dbutton)
		rows = append(rows, row)
	}
	result = tgbotapi.NewMessage(tgMessage.Chat.ID, "и вот же они, сиписьки")
	result.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	metrics.TotalListCommands.WithLabelValues("/lists").Add(1)
	return
}

func GetRandomItem(ctx context.Context, tgMessage *tgbotapi.Message) (result tgbotapi.MessageConfig) {

	if tgMessage.CommandArguments() == "" {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "/randit list_id")
	}

	args := tgMessage.CommandArguments()
	listID, countStr, _ := strings.Cut(args, " ")
	if listID == "" {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ не нашел id списка")
	}

	var (
		count uint64 = 1
		err   error
	)

	if countStr != "" {
		count, err = strconv.ParseUint(countStr, 10, 64)
		if err != nil {
			return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ %s это не число, а нужно либо ничего, либо число", countStr))
		}
	}

	channelID := valueObjects.ChannelID(tgMessage.Chat.ID)
	items, err := bot.listManager.GetRandomItems(ctx, channelID, valueObjects.ListID(listID), count)
	if err != nil {
		switch {
		case errors.Is(err, localerrors.ErrDoesntExist):
			return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ списка под названием <i>%s</i> не существует", listID))
		case errors.Is(err, localerrors.ErrNoItems):
			return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ в <i>%s</i> списке нет еще элементов", listID))
		default:
			logger.Errorf("GetRandomItem: %v", err)
			return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время получания слйчайного элемента списка, повторите позже")
		}
	}

	resMsg := fmt.Sprintf("🎲 случайные элементы из списка <i>%s</i>:\n", listID)
	for _, item := range items {
		resMsg += fmt.Sprintf("👉 <b>%s</b>\nby <i>%s</i>\n\n", item.Name, item.Author)
	}
	resMsg += fmt.Sprintf("🙍 сгенерирован по просьбе <i>%s</i>", tgMessage.From.UserName)

	metrics.TotalListCommands.WithLabelValues("/randit").Add(1)
	return tgbotapi.NewMessage(tgMessage.Chat.ID, resMsg)
}

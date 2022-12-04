package tg

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	localerrors "github.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	"github.com/Sh00ty/dormitory_room_bot/internal/logger"
	scheduler "github.com/Sh00ty/dormitory_room_bot/internal/scheduler"
	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Subscribe(ctx context.Context, tgMessage *tgbotapi.Message) tgbotapi.MessageConfig {
	taskID := tgMessage.CommandArguments()
	taskID = strings.Trim(taskID, " ")
	err := bot.taskManager.Subscribe(ctx, valueObjects.TaskID(taskID), valueObjects.ChannelID(tgMessage.Chat.ID), "@"+valueObjects.UserID(tgMessage.From.UserName))
	if err != nil {
		switch {
		case errors.Is(err, localerrors.ErrDoesntExist):
			return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ задачи с id %s не найдено", taskID))
		case errors.Is(err, localerrors.ErrNotSubTask):
			return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ задача с %s не поддерживает подписки или отписки", taskID))
		case errors.Is(err, localerrors.ErrAlreadyExists):
			return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ ты уже подсисан на эту задачку")
		}
		logger.Errorf("Subscribe: %v", err)
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время подписки на задачу")
	}
	taskString := fmt.Sprintf("✅ %s подпислся на задачу %s", tgMessage.From.UserName, taskID)
	return tgbotapi.NewMessage(tgMessage.Chat.ID, taskString)
}

func UnSubscribe(ctx context.Context, tgMessage *tgbotapi.Message) tgbotapi.MessageConfig {
	taskID := tgMessage.CommandArguments()
	taskID = strings.Trim(taskID, " ")
	err := bot.taskManager.UnSubscribe(ctx, valueObjects.TaskID(taskID), valueObjects.ChannelID(tgMessage.Chat.ID), "@"+valueObjects.UserID(tgMessage.From.UserName))
	if err != nil {
		switch {
		case errors.Is(err, localerrors.ErrDoesntExist):
			return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ задачи с id %s не найдено", taskID))
		case errors.Is(err, localerrors.ErrNotSubTask):
			return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ задача с %s не поддерживает подписки или отписки", taskID))
		case errors.Is(err, localerrors.ErrDidntSubscribed):
			return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ ты ж не подписывался на эту задачку")
		case errors.Is(err, localerrors.ErrNoWorkers):
			return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ вообще нет никого кто подписан на это")
		}
		logger.Errorf("Subscribe: %v", err)
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время отписки от задачи")
	}
	taskString := fmt.Sprintf("✅ %s отписался от задачи %s", tgMessage.From.UserName, taskID)
	return tgbotapi.NewMessage(tgMessage.Chat.ID, taskString)
}

func createTask(ctx context.Context, tgMessage *tgbotapi.Message, typ entities.TaskType) tgbotapi.MessageConfig {
	taskID, title, workers, _, notifyInterval, notifyTime := taskParser(tgMessage)
	if taskID == "" {
		switch typ {
		case entities.Default:
			return tgbotapi.NewMessage(tgMessage.Chat.ID, "пример:\n/task t_id @w1 @w2 @w3 2 24h 05-11-2022+21:27 desc")
		}
	}

	options := ""

	if len(workers) == 0 {
		options += "\n❌ исполнители"
	} else {
		options += fmt.Sprintf("\n✅ обнаружено <b>%d</b> исполнителей", len(workers))
	}

	workerCount := uint32(len(workers))

	isNotified := false
	if notifyInterval == nil {
		options += "\n❌ повтор нотификации"
	} else {
		isNotified = true
		options += fmt.Sprintf("\n⏱ повторение для нотификации: %s", notifyInterval.String())
	}

	if notifyTime == nil {
		options += "\n❌ точная дата нотификации"
	} else {
		isNotified = true
		options += fmt.Sprintf("\n⏱ дата нотификации: %s", notifyTime.Add(3*time.Hour).Format("02-01-2022+18:47"))
	}

	if !isNotified && typ == entities.OneShot {
		momentErr := "\n❌Момент не может быть без интервала и без времени напоминания одновременно, должно быть хотя бы одно"
		return tgbotapi.NewMessage(tgMessage.Chat.ID, momentErr)
	}
	err := bot.taskManager.Create(ctx, taskID, typ, title, tgMessage.From.UserName, workers, workerCount, notifyInterval, notifyTime, valueObjects.ChannelID(tgMessage.Chat.ID))
	if err != nil {
		if errors.Is(err, localerrors.ErrAlreadyExists) {
			return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ задача с данным <i>id</i> уже существует")
		}
		logger.Errorf("CreateTask: %v", err)
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время создания задачи, повторите позже")
	}
	return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("✅ <i>Задача:</i>  <b>%s</b> успешно создана\n%s", taskID, options))
}

func CreateOneShotTask(ctx context.Context, tgMessage *tgbotapi.Message) tgbotapi.MessageConfig {
	return createTask(ctx, tgMessage, entities.OneShot)
}

func CreateSubsTask(ctx context.Context, tgMessage *tgbotapi.Message) tgbotapi.MessageConfig {
	return createTask(ctx, tgMessage, entities.Subscribed)
}

func CreateDefaultTask(ctx context.Context, tgMessage *tgbotapi.Message) tgbotapi.MessageConfig {
	return createTask(ctx, tgMessage, entities.Default)
}

func GetTask(ctx context.Context, tgMessage *tgbotapi.Message) tgbotapi.MessageConfig {

	taskID := tgMessage.CommandArguments()
	taskID = strings.Trim(taskID, " ")
	task, err := bot.taskManager.Get(ctx, valueObjects.TaskID(taskID), valueObjects.ChannelID(tgMessage.Chat.ID))
	if err != nil {
		if errors.Is(err, localerrors.ErrDoesntExist) {
			return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ задачи с id <i>%s</i> не найдено", taskID))
		}
		logger.Errorf("GetTask: %v", err)
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время получения задачи")
	}

	taskString := GetTaskString(ctx, task)
	taskString += fmt.Sprintf("\n\n🙍 показана by <i>%s</i>", tgMessage.From.UserName)
	return tgbotapi.NewMessage(tgMessage.Chat.ID, taskString)
}

func DeleteTask(ctx context.Context, tgMessage *tgbotapi.Message) tgbotapi.MessageConfig {

	taskID := tgMessage.CommandArguments()
	taskID = strings.Trim(taskID, " ")
	err := bot.taskManager.Delete(ctx, valueObjects.TaskID(taskID), valueObjects.ChannelID(tgMessage.Chat.ID))
	if err != nil {
		if errors.Is(err, localerrors.ErrDoesntExist) {
			return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ задачи с id <i>%s</i> не найдено", taskID))
		}
		logger.Errorf("DeleteTask: %v", err)
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время удаления задачи")
	}

	return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("🗑️ задача <i>%s</i> удалена by <i>%s</i>", taskID, tgMessage.From.UserName))
}

func GetTaskString(ctx context.Context, task entities.Task) string {
	msg := fmt.Sprintf("📜 <i>Задача:</i>  <b>%s</b>\n", task.ID)

	workers, err := task.Getworkers()
	if err == nil && len(workers) != 0 {
		msg += "👨 <i>Исполнители:</i> "
		for _, w := range workers {
			msg += "\n       - " + string(w)
		}
	}
	if task.Title != "" {
		msg += fmt.Sprintf("\n📌 <i>Описание:</i>\n       %s", task.Title)
	}
	return msg
}

func GetNotifications(interval time.Duration) func() {
	ctx := context.Background()

	s := scheduler.NewTimeSheduller()
	closer := s.Start(interval, func() error {

		tasks, channelIDs, err := bot.taskManager.NotificationCheck(ctx)
		if err != nil {
			return err
		}
		type taskWIthChannel struct {
			channelID valueObjects.ChannelID
			taskID    valueObjects.TaskID
		}
		tasksToDelete := make([]taskWIthChannel, 0, len(tasks))
		for i, task := range tasks {
			result := GetTaskString(ctx, task)
			if task.IsOneShot() {
				tasksToDelete = append(tasksToDelete, taskWIthChannel{channelIDs[i], task.ID})
			}
			msg := tgbotapi.NewMessage(int64(channelIDs[i]), result)
			msg.ParseMode = tgbotapi.ModeHTML
			_, err = bot.bot.Send(msg)
			if err != nil {
				logger.Error(err.Error())
			}
		}

		for _, task := range tasksToDelete {
			err = bot.taskManager.Delete(ctx, task.taskID, task.channelID)
			if err != nil {
				// TODO: make some notification in channel, but not at this tread
				// put in rabbitmq (when it will be created sender service)
				logger.Errorf("can't delete one shot task : %w", err)
			}
		}
		return err
	})
	return closer
}

func ChangeWorker(ctx context.Context, tgMessage *tgbotapi.Message) tgbotapi.MessageConfig {

	taskID := tgMessage.CommandArguments()
	task, err := bot.taskManager.UpdateWorkers(ctx, valueObjects.TaskID(taskID), valueObjects.ChannelID(tgMessage.Chat.ID))
	if err != nil {
		switch {
		case errors.Is(err, localerrors.ErrDoesntExist):
			return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ задачи <i>%s</i> не существует", taskID))
		case errors.Is(err, localerrors.ErrNoWorkers):
			return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ у данной задачи <i>%s</i> нет исполнителей, камон", taskID))
		default:
			logger.Errorf("ChangeWorker: %v", err)
			return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время смены исполнителей, повторите позже")

		}
	}

	workers := task.Workers.GetWorkers()

	resMsg := "✅ успешно поменял работников на:"
	for _, w := range workers {
		resMsg += fmt.Sprintf("\n - %s", w)
	}
	resMsg += fmt.Sprintf("\n\nпо просьбе %s", tgMessage.From.UserName)
	return tgbotapi.NewMessage(tgMessage.Chat.ID, resMsg)
}

func GetAllTasks(ctx context.Context, tgMessage *tgbotapi.Message) (result tgbotapi.MessageConfig) {

	channelID := valueObjects.ChannelID(tgMessage.Chat.ID)
	tasks, err := bot.taskManager.GetAll(ctx, channelID)
	if err != nil {
		if errors.Is(err, localerrors.ErrDoesntExist) {
			return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ вы не делали никаких заданий, ничего не знаю")
		}
		logger.Errorf("GetAllTasks: %v", err)
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время получания ваших задач, повторите позже")
	}
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(tasks))
	for _, task := range tasks {
		gbutton := tgbotapi.NewInlineKeyboardButtonData(string(task.ID), "get:"+string(task.ID))
		dbutton := tgbotapi.NewInlineKeyboardButtonData("🗑", "delete:"+string(task.ID))
		row := tgbotapi.NewInlineKeyboardRow(gbutton, dbutton)
		_, err := task.Getworkers()
		if task.IsSubscribed() {
			sbutton := tgbotapi.NewInlineKeyboardButtonData("🔔", "sub:"+string(task.ID))
			row = append(row, sbutton)
			if err == nil {
				ubutton := tgbotapi.NewInlineKeyboardButtonData("🔕", "unsub:"+string(task.ID))
				row = append(row, ubutton)
			}
		}

		if err == nil {
			cbutton := tgbotapi.NewInlineKeyboardButtonData("🔁", "change:"+string(task.ID))
			row = append(row, cbutton)
		}
		rows = append(rows, row)
	}
	result = tgbotapi.NewMessage(tgMessage.Chat.ID, "<b>Ваши задачки</b> 🙊🙊🙊:")
	result.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	return
}

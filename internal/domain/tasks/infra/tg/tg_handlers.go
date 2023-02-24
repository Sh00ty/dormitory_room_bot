package tasks

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.com/Sh00ty/dormitory_room_bot/internal/domain/tasks"
	"gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"
	localerrors "gitlab.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
	"gitlab.com/Sh00ty/dormitory_room_bot/pkg/logger"
	scheduler "gitlab.com/Sh00ty/dormitory_room_bot/pkg/scheduler"
	"gitlab.com/Sh00ty/dormitory_room_bot/pkg/tgproc"
)

type tgbot struct {
	taskManager tasks.TaskManager
	sender      tgproc.MessageSender
}

func NewTgbot(tm tasks.TaskManager, s tgproc.MessageSender) *tgbot {
	return &tgbot{taskManager: tm, sender: s}
}

func (bot tgbot) Subscribe(ctx context.Context, query *tgbotapi.CallbackQuery) []tgbotapi.MessageConfig {
	taskID := valueObjects.TaskID(query.Data)
	channelID := query.Message.Chat.ID

	err := bot.taskManager.Subscribe(ctx, valueObjects.TaskID(taskID), valueObjects.ChannelID(channelID), "@"+valueObjects.UserID(query.From.UserName))
	if err != nil {
		switch {
		case errors.Is(err, localerrors.ErrDoesntExist):
			return tgproc.MakeMessage(channelID, fmt.Sprintf("❌ задачи с id %s не найдено", taskID))
		case errors.Is(err, localerrors.ErrNotSubTask):
			return tgproc.MakeMessage(channelID, fmt.Sprintf("❌ задача с %s не поддерживает подписки или отписки", taskID))
		case errors.Is(err, localerrors.ErrAlreadyExists):
			return tgproc.MakeMessage(channelID, "❌ ты уже подсисан на эту задачку")
		}
		logger.Errorf("Subscribe: %v", err)
		return tgproc.MakeMessage(channelID, "❌ произошла ошибка во время подписки на задачу")
	}
	taskString := fmt.Sprintf("✅ %s подпислся на задачу %s", query.From.UserName, taskID)
	return tgproc.MakeMessage(channelID, taskString)
}

func (bot tgbot) UnSubscribe(ctx context.Context, query *tgbotapi.CallbackQuery) []tgbotapi.MessageConfig {
	taskID := valueObjects.TaskID(query.Data)
	channelID := query.Message.Chat.ID

	err := bot.taskManager.UnSubscribe(ctx, valueObjects.TaskID(taskID), valueObjects.ChannelID(channelID), "@"+valueObjects.UserID(query.From.UserName))
	if err != nil {
		switch {
		case errors.Is(err, localerrors.ErrDoesntExist):
			return tgproc.MakeMessage(channelID, fmt.Sprintf("❌ задачи с id %s не найдено", taskID))
		case errors.Is(err, localerrors.ErrNotSubTask):
			return tgproc.MakeMessage(channelID, fmt.Sprintf("❌ задача с %s не поддерживает подписки или отписки", taskID))
		case errors.Is(err, localerrors.ErrDidntSubscribed):
			return tgproc.MakeMessage(channelID, "❌ ты ж не подписывался на эту задачку")
		case errors.Is(err, localerrors.ErrNoWorkers):
			return tgproc.MakeMessage(channelID, "❌ вообще нет никого кто подписан на это")
		}
		logger.Errorf("Subscribe: %v", err)
		return tgproc.MakeMessage(channelID, "❌ произошла ошибка во время отписки от задачи")
	}
	taskString := fmt.Sprintf("✅ %s отписался от задачи %s", query.From.UserName, taskID)
	return tgproc.MakeMessage(channelID, taskString)
}

func (bot tgbot) CreateOneShotTask(ctx context.Context, tgMessage *tgbotapi.Message) []tgbotapi.MessageConfig {
	return bot.createTask(ctx, tgMessage, entities.OneShot)
}

func (bot tgbot) CreateSubsTask(ctx context.Context, tgMessage *tgbotapi.Message) []tgbotapi.MessageConfig {
	return bot.createTask(ctx, tgMessage, entities.Subscribed)
}

func (bot tgbot) CreateDefaultTask(ctx context.Context, tgMessage *tgbotapi.Message) []tgbotapi.MessageConfig {
	return bot.createTask(ctx, tgMessage, entities.Default)
}

func (bot tgbot) GetTask(ctx context.Context, query *tgbotapi.CallbackQuery) []tgbotapi.MessageConfig {
	taskID := valueObjects.TaskID(query.Data)
	channelID := query.Message.Chat.ID

	task, err := bot.taskManager.Get(ctx, valueObjects.TaskID(taskID), valueObjects.ChannelID(channelID))
	if err != nil {
		if errors.Is(err, localerrors.ErrDoesntExist) {
			return tgproc.MakeMessage(channelID, fmt.Sprintf("❌ задачи с id <i>%s</i> не найдено", taskID))
		}
		logger.Errorf("GetTask: %v", err)
		return tgproc.MakeMessage(channelID, "❌ произошла ошибка во время получения задачи")
	}

	taskString := GetTaskString(ctx, task)
	taskString += fmt.Sprintf("\n\n🙍 показана by <i>%s</i>", query.From.UserName)
	return tgproc.MakeMessage(channelID, taskString)
}

func (bot tgbot) DeleteTask(ctx context.Context, query *tgbotapi.CallbackQuery) []tgbotapi.MessageConfig {
	taskID := valueObjects.TaskID(query.Data)
	channelID := query.Message.Chat.ID

	err := bot.taskManager.Delete(ctx, taskID, valueObjects.ChannelID(channelID))
	if err != nil {
		if errors.Is(err, localerrors.ErrDoesntExist) {
			return tgproc.MakeMessage(channelID, fmt.Sprintf("❌ задачи с id <i>%s</i> не найдено", taskID))
		}
		logger.Errorf("DeleteTask: %v", err)
		return tgproc.MakeMessage(channelID, "❌ произошла ошибка во время удаления задачи")
	}

	return tgproc.MakeMessage(channelID, fmt.Sprintf("🗑️ задача <i>%s</i> удалена by <i>%s</i>", taskID, query.From.UserName))
}

func (bot tgbot) ChangeWorker(ctx context.Context, query *tgbotapi.CallbackQuery) []tgbotapi.MessageConfig {
	taskID := valueObjects.TaskID(query.Data)
	channelID := query.Message.Chat.ID

	task, err := bot.taskManager.UpdateWorkers(ctx, valueObjects.TaskID(taskID), valueObjects.ChannelID(channelID))
	if err != nil {
		switch {
		case errors.Is(err, localerrors.ErrDoesntExist):
			return tgproc.MakeMessage(channelID, fmt.Sprintf("❌ задачи <i>%s</i> не существует", taskID))
		case errors.Is(err, localerrors.ErrNoWorkers):
			return tgproc.MakeMessage(channelID, fmt.Sprintf("❌ у данной задачи <i>%s</i> нет исполнителей, камон", taskID))
		default:
			logger.Errorf("ChangeWorker: %v", err)
			return tgproc.MakeMessage(channelID, "❌ произошла ошибка во время смены исполнителей, повторите позже")

		}
	}

	workers := task.Workers.GetWorkers()

	resMsg := "✅ успешно поменял работников на:"
	for _, w := range workers {
		resMsg += fmt.Sprintf("\n - %s", w)
	}
	resMsg += fmt.Sprintf("\n\nпо просьбе %s", query.From.UserName)
	return tgproc.MakeMessage(channelID, resMsg)
}

func (bot tgbot) GetAllTasks(ctx context.Context, tgMessage *tgbotapi.Message) (result []tgbotapi.MessageConfig) {

	channelID := valueObjects.ChannelID(tgMessage.Chat.ID)
	tasks, err := bot.taskManager.GetAll(ctx, channelID)
	if err != nil {
		if errors.Is(err, localerrors.ErrDoesntExist) {
			return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ вы не делали никаких заданий, ничего не знаю")
		}
		logger.Errorf("GetAllTasks: %v", err)
		return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время получания ваших задач, повторите позже")
	}
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(tasks))
	for _, task := range tasks {
		gbutton := tgbotapi.NewInlineKeyboardButtonData(string(task.ID), "get_t:"+string(task.ID))
		dbutton := tgbotapi.NewInlineKeyboardButtonData("🗑", "del_t:"+string(task.ID))
		row := tgbotapi.NewInlineKeyboardRow(gbutton, dbutton)
		_, err := task.Getworkers()
		if task.IsSubscribed() {
			sbutton := tgbotapi.NewInlineKeyboardButtonData("🔔", "subt:"+string(task.ID))
			row = append(row, sbutton)
			if err == nil {
				ubutton := tgbotapi.NewInlineKeyboardButtonData("🔕", "unsubt:"+string(task.ID))
				row = append(row, ubutton)
			}
		}

		if err == nil {
			cbutton := tgbotapi.NewInlineKeyboardButtonData("🔁", "change_w:"+string(task.ID))
			row = append(row, cbutton)
		}
		rows = append(rows, row)
	}
	tmpRes := tgbotapi.NewMessage(tgMessage.Chat.ID, "<b>Ваши задачки</b> 🙊🙊🙊:")
	tmpRes.ParseMode = tgbotapi.ModeHTML
	tmpRes.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	result = append(result, tmpRes)
	return
}

func (bot tgbot) createTask(ctx context.Context, tgMessage *tgbotapi.Message, typ entities.TaskType) []tgbotapi.MessageConfig {
	taskID, title, workers, workerCount, notifyInterval, notifyTime := taskParser(tgMessage)
	if taskID == "" {
		switch typ {
		case entities.Default:
			return tgproc.MakeMessage(tgMessage.Chat.ID, "пример:\n/task t_id @w1 @w2 @w3 2 24h 05-11-2022+21:27 desc")
		case entities.OneShot:
			return tgproc.MakeMessage(tgMessage.Chat.ID, "пример:\n/moment t_id @w1 @w2 @w3 2 24h 05-11-2022+21:27 desc")
		case entities.Subscribed:
			return tgproc.MakeMessage(tgMessage.Chat.ID, "пример:\n/subt t_id @w1 @w2 @w3 2 24h 05-11-2022+21:27 desc")
		}
	}

	options := ""

	if len(workers) == 0 {
		options += "\n❌ исполнители"
	} else {
		options += fmt.Sprintf("\n✅ обнаружено <b>%d</b> исполнителей", len(workers))
	}

	if typ == entities.Subscribed {
		workerCount = uint32(len(workers))
	}

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
		return tgproc.MakeMessage(tgMessage.Chat.ID, momentErr)
	}
	err := bot.taskManager.Create(ctx, taskID, typ, title, tgMessage.From.UserName, workers, workerCount, notifyInterval, notifyTime, valueObjects.ChannelID(tgMessage.Chat.ID))
	if err != nil {
		if errors.Is(err, localerrors.ErrAlreadyExists) {
			return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ задача с данным <i>id</i> уже существует")
		}
		logger.Errorf("CreateTask: %v", err)
		return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время создания задачи, повторите позже")
	}
	return tgproc.MakeMessage(tgMessage.Chat.ID, fmt.Sprintf("✅ <i>Задача:</i>  <b>%s</b> успешно создана\n%s", taskID, options))
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

func (bot tgbot) GetNotifications(interval time.Duration) func() {
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
			_, err = bot.sender.Send(msg)
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

type parseStage int

const (
	taskIDStage parseStage = iota
	workersStage
	workerCountStage
	notifyIntervalStage
	notifyTimeStage
	titleStage
)

func taskParser(tgMessage *tgbotapi.Message) (taskID valueObjects.TaskID,
	title string,
	workers []valueObjects.UserID,
	workerCount uint32,
	notifyInterval *time.Duration,
	notifyTime *time.Time) {

	var (
		exists   bool
		curStage = taskIDStage
		after    = tgMessage.CommandArguments() + " "
		before   = ""
		iscut    = true
	)

Loop:
	for {
		if iscut {
			before, after, exists = strings.Cut(after, " ")

			if !exists || before == "" {
				return
			}
		}

		switch curStage {
		case taskIDStage:
			taskID = valueObjects.TaskID(before)
			curStage++
		case workersStage:
			if before[:1] != "@" {
				curStage++
				iscut = false
				continue Loop
			}
			workers = append(workers, valueObjects.UserID(before))
		case workerCountStage:
			tmp, err := strconv.ParseUint(before, 10, 64)
			curStage++
			iscut = err == nil
			if err == nil {
				workerCount = uint32(tmp)
			}
		case notifyIntervalStage:
			tmp, err := time.ParseDuration(before)
			curStage++
			iscut = err == nil
			if err == nil {
				notifyInterval = &tmp
			}
		case notifyTimeStage:

			// day-mounth-year+hour:minute
			t, err := time.ParseInLocation("02-01-2006+15:04", before, time.Local)
			iscut = err == nil
			curStage++
			if err == nil {
				// какие-то проблемы с временной локалью
				t = t.Add(-3 * time.Hour)
				notifyTime = &t
			}
		case titleStage:
			title = before + " " + after
			return
		default:
			return
		}

	}
}

package tg

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type parseStage int

const (
	creditorStage = iota
	creditStage
)

func oweParser(args string) (users []valueObjects.UserID, credit int64, err error) {
	args = strings.ReplaceAll(args, "  ", " ")

	var (
		exists bool
		iscut  = true
		after  = args + " "
		before = ""

		curStage = creditorStage
	)

	for {
		if iscut {
			before, after, exists = strings.Cut(after, " ")

			if !exists || before == "" {
				return
			}
		}

		switch curStage {
		case creditorStage:
			if before[:1] != "@" {
				curStage++
				iscut = false
				continue
			}
			users = append(users, valueObjects.UserID(before))
		case creditStage:
			credit, err = strconv.ParseInt(before, 10, 64)
			if err != nil {
				return nil, 0, fmt.Errorf("❌ бро %s - это не число, удали телегу пж", before)
			}
			curStage++
		default:
			return
		}
	}
}

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

package tg

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	localerrors "github.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	"github.com/Sh00ty/dormitory_room_bot/internal/logger"
	metrics "github.com/Sh00ty/dormitory_room_bot/internal/metrics"
	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// в таск мэнэджере нет возможности атомарно добвавить всем юзерам долги, поэтому нельзя несколько сразу
// Андрей должен сделать и переделать эту фн для многих
func Owe(ctx context.Context, tgMessage *tgbotapi.Message) tgbotapi.MessageConfig {

	buyer := entities.User{
		ID:        valueObjects.UserID("@" + tgMessage.From.UserName),
		ChannelID: valueObjects.ChannelID(tgMessage.Chat.ID),
	}

	args := tgMessage.CommandArguments()

	payers, credit, err := oweParser(args)
	if err != nil {
		// внутри все ошибки человеческие, но лучше поправить это
		return tgbotapi.NewMessage(tgMessage.Chat.ID, err.Error())
	}
	if len(payers) == 0 {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ должники не обнаружены")
	}
	if credit == 0 {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ го задолженнность будет не нулевая")
	}
	credit = int64(math.Ceil(float64(credit) / float64(len(payers))))

	err = bot.debitSvc.CreditTransaction(ctx, buyer.ChannelID, buyer.ID, payers, valueObjects.Money(credit))
	if err != nil {
		if errors.As(err, &localerrors.UpdateCreditError{}) {
			updateErr := err.(localerrors.UpdateCreditError)
			if errors.Is(updateErr.Err, localerrors.ErrDoesntExist) {
				return tgbotapi.NewMessage(tgMessage.Chat.ID,
					fmt.Sprintf("❌ %s не зарегистрирован\nтранзакция провалена", updateErr.User))
			}
			return tgbotapi.NewMessage(tgMessage.Chat.ID, fmt.Sprintf("❌ не могу обновить счет %s\nтранзакция провалена", updateErr.User))
		}
	}

	resMsg := fmt.Sprintf("[%s]\n", strings.Repeat("+", len(payers)))
	for _, payer := range payers {
		resMsg += fmt.Sprintf("%s должен @%s -> %d\n", payer, tgMessage.From.UserName, credit)
	}
	resMsg += "✅ транзакция успешно выполнена"

	metrics.TotalCreditsCommands.WithLabelValues("/credit").Add(1)
	metrics.TotalCredidSum.WithLabelValues(strconv.FormatInt(int64(buyer.ChannelID), 10)).Add(float64(credit))
	return tgbotapi.NewMessage(tgMessage.Chat.ID, resMsg)
}

func Checkout(ctx context.Context, tgMessage *tgbotapi.Message) tgbotapi.MessageConfig {

	channelID := valueObjects.ChannelID(tgMessage.Chat.ID)

	credits, err := bot.debitSvc.GetAll(ctx, channelID)
	if err != nil {
		logger.Errorf("Checkout: %v", err)
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время получения балансов, повторите позже")
	}
	if len(credits) == 0 {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ в этом канале нет зарегистрированных юзеров")
	}
	transactions := bot.debitReolver.ResolveCredits(credits)

	msg := "✅ DUE TO GOLD STEP ALGO:\n"

	if len(transactions) == 0 {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ нет транзакций")
	}

	for _, tx := range transactions {
		msg += fmt.Sprintf("\n%s должен %s %d рублей", tx.From, tx.To, tx.Amount)
	}

	err = bot.debitSvc.ClearBalances(ctx, channelID)
	if err != nil {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ не удалось почистить балансы, попробуйте позже")
	}

	metrics.TotalCreditsCommands.WithLabelValues("/checkout").Add(1)
	return tgbotapi.NewMessage(tgMessage.Chat.ID, msg+"\n✅ я почистил балансы")
}

func GetResults(ctx context.Context, tgMessage *tgbotapi.Message) tgbotapi.MessageConfig {

	channelID := valueObjects.ChannelID(tgMessage.Chat.ID)

	credits, err := bot.debitSvc.GetAll(ctx, channelID)
	if err != nil {
		logger.Errorf("GetResults: %v", err)
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время получения балансов, попробуйте позже")
	}

	if len(credits) == 0 {
		return tgbotapi.NewMessage(tgMessage.Chat.ID, "❌ нет зарегистрированных юзеров для этого канала")
	}
	msg := "💰 <b>Балансы</b>:"

	for i, Credit := range credits {
		msg += fmt.Sprintf("\n%d. %s ---> %d₽", i, Credit.UserID[1:], Credit.Credit)
	}

	metrics.TotalCreditsCommands.WithLabelValues("/bank").Add(1)
	return tgbotapi.NewMessage(tgMessage.Chat.ID, msg)
}

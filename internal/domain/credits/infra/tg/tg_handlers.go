package credit

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.com/Sh00ty/dormitory_room_bot/internal/domain/credits"
	"gitlab.com/Sh00ty/dormitory_room_bot/internal/entities"
	localerrors "gitlab.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	metrics "gitlab.com/Sh00ty/dormitory_room_bot/internal/metrics"
	valueObjects "gitlab.com/Sh00ty/dormitory_room_bot/internal/value_objects"
	"gitlab.com/Sh00ty/dormitory_room_bot/pkg/calculator"
	"gitlab.com/Sh00ty/dormitory_room_bot/pkg/logger"
	"gitlab.com/Sh00ty/dormitory_room_bot/pkg/tgproc"
)

type tgbot struct {
	creditManager credits.CreditManager
	oweCalc       calculator.Calculator
}

func New(u credits.CreditManager) *tgbot {
	return &tgbot{u, calculator.New()}
}

var (
	// чтобы не вводить свой ник можно писать такие плэйсхолдеры
	selfCreditPlaceholders = []valueObjects.UserID{"@me", "@i", "@я"}
)

func (bot tgbot) Owe(ctx context.Context, tgMessage *tgbotapi.Message) []tgbotapi.MessageConfig {
	channelID := tgMessage.Chat.ID
	buyer := entities.User{
		ID:        valueObjects.UserID("@" + tgMessage.From.UserName),
		ChannelID: valueObjects.ChannelID(channelID),
	}

	args := tgMessage.CommandArguments()

	payers, credit, err := bot.oweParser(args)
	if err != nil {
		// внутри все ошибки человеческие, но лучше поправить это
		return tgproc.MakeMessage(channelID, err.Error())
	}
	if len(payers) == 0 {
		return tgproc.MakeMessage(channelID, "❌ должники не обнаружены")
	}
	if credit == 0 {
		return tgproc.MakeMessage(channelID, "❌ го задолженнность будет не нулевая")
	}

	for i, payer := range payers {
		for _, placeholder := range selfCreditPlaceholders {
			if payer == placeholder {
				payers[i] = buyer.ID
			}
		}
	}

	credit = int64(math.Ceil(float64(credit) / float64(len(payers))))

	if err = bot.creditManager.CreditTransaction(ctx, buyer.ChannelID, buyer.ID, payers, valueObjects.Money(credit)); err != nil {
		if errors.As(err, &localerrors.UpdateCreditError{}) {
			updateErr := err.(localerrors.UpdateCreditError)
			if errors.Is(updateErr.Err, localerrors.ErrDoesntExist) {
				return tgproc.MakeMessage(channelID,
					fmt.Sprintf("❌ %s не зарегистрирован\nтранзакция провалена", updateErr.User))
			}
			return tgproc.MakeMessage(channelID, fmt.Sprintf("❌ не могу обновить счет %s\nтранзакция провалена", updateErr.User))
		}
	}

	resMsg := ""
	for _, payer := range payers {
		if credit > 0 {
			resMsg += fmt.Sprintf("%s должен @%s -> %d\n", payer, tgMessage.From.UserName, credit)
		} else {
			// разварачиваем чтобы было удобнее смотреть
			resMsg += fmt.Sprintf("@%s должен %s -> %d\n", tgMessage.From.UserName, payer, -credit)
		}
	}
	resMsg += "✅ транзакция успешно выполнена"

	metrics.TotalCreditsCommands.WithLabelValues("/credit").Add(1)
	metrics.TotalCredidSum.WithLabelValues(strconv.FormatInt(channelID, 10)).Add(float64(credit))
	return tgproc.MakeMessage(channelID, resMsg)
}

type parseStage int

const (
	creditorStage parseStage = iota
	creditStage
	calculatorStage
)

func (bot tgbot) oweParser(args string) (users []valueObjects.UserID, credit int64, err error) {

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
			if strings.HasPrefix(before, "{") {
				curStage++
				iscut = false
				continue
			}
			credit, err = strconv.ParseInt(before, 10, 64)
			if err != nil {
				return nil, 0, fmt.Errorf("❌ бро %s - это не число, удали телегу пж", before)
			}
			return
		case calculatorStage:
			tokens, err := parseCalc(before + after)
			if err != nil {
				logger.Errorf("oweParser : %s", err.Error())
				return nil, 0, err
			}
			if len(tokens) == 0 {
				return nil, 0, fmt.Errorf("❌ c выражением что-то не так, оно будто бы пустое для меня")
			}
			credit, err = bot.oweCalc.Calculate(tokens)
			if err != nil {
				logger.Errorf("oweParser : %s", err.Error())
				switch {
				case errors.Is(err, calculator.ErrInvalidBrackets):
					return nil, 0, fmt.Errorf("❌ не могу подсчитать это, нашел закрывающую скобку, а открывающую не нашел")
				case errors.Is(err, calculator.ErrZeroDivision):
					return nil, 0, fmt.Errorf("❌ как же я осуждаю деление на ноль, не буду это считать")
				case errors.Is(err, calculator.ErrInvalidExpr):
					return nil, 0, fmt.Errorf("❌ что-то не так с этим выражением, попробуй ввести что-то более адекватное для меня")
				default:
					return nil, 0, fmt.Errorf("❌ хз что случилось, но подсчитать я это не смогу")
				}
			}
			return users, credit, nil
		default:
			return
		}
	}
}

func parseCalc(args string) ([]calculator.Token, error) {
	tokens := make([]calculator.Token, 0, 5)
	num := int64(0)
	isNum := false
	logger.Infof("oweArgs : %s", args)
loop:
	for _, char := range args {
		if char == ' ' {
			if isNum {
				tokens = append(tokens, calculator.GetNumberToken(num))
				num = 0
				isNum = false
			}
			continue
		}
		if char <= '9' && char >= '0' {
			num = num*10 + int64(char-48)
			isNum = true
			continue
		}

		if isNum {
			tokens = append(tokens, calculator.GetNumberToken(num))
			num = 0
			isNum = false
		}
		switch char {
		case '{':
			continue
		case '}':
			break loop
		default:
			token, err := calculator.GetTokenByRune(char)
			if err != nil {
				return nil, fmt.Errorf("❌ что-то я не могу это дело распознать, мешает - \"%s\"", string(char))
			}
			tokens = append(tokens, token)
			continue
		}
	}
	return tokens, nil
}

func (bot tgbot) Checkout(ctx context.Context, tgMessage *tgbotapi.Message) (res []tgbotapi.MessageConfig) {
	channelID := tgMessage.Chat.ID

	transactions, err := bot.creditManager.Checkout(ctx, valueObjects.ChannelID(channelID))
	if err != nil {
		switch {
		case errors.Is(err, localerrors.ErrDoesntExist):
			return tgproc.MakeMessage(channelID, "❌ что-то не нашел в данном канале зарегестрированных пользователей")
		default:
			return tgproc.MakeMessage(channelID, "❌ произошла непредвиденная ситуация во время расчета транзакций")
		}
	}
	if len(transactions) == 0 {
		return tgproc.MakeMessage(channelID, "🦥 все пусто, нет транзакций для вас")
	}
	msg := "✅ DUE TO GOLD STEPOS ALGO:\n🕺Я ОБЧИСТИЛ ВАШИ БАЛАНСЫ🕺"
	var prevSender valueObjects.UserID
	for _, tx := range transactions {
		if prevSender != tx.From {
			smsg := tgbotapi.NewMessage(channelID, msg)
			smsg.ParseMode = tgbotapi.ModeHTML
			res = append(res, smsg)
			msg = ""
			prevSender = tx.From
		}
		msg += fmt.Sprintf("<b>%s</b> должен <b>%s</b> %d₽\n", tx.From, tx.To, tx.Amount)
	}
	smsg := tgbotapi.NewMessage(channelID, msg)
	smsg.ParseMode = tgbotapi.ModeHTML
	res = append(res, smsg)
	metrics.TotalCreditsCommands.WithLabelValues("/checkout").Add(1)
	return res
}

func (bot tgbot) GetResults(ctx context.Context, tgMessage *tgbotapi.Message) []tgbotapi.MessageConfig {

	channelID := valueObjects.ChannelID(tgMessage.Chat.ID)

	credits, err := bot.creditManager.GetAll(ctx, channelID)
	if err != nil {
		logger.Errorf("GetResults: %v", err)
		return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ произошла ошибка во время получения балансов, попробуйте позже")
	}

	if len(credits) == 0 {
		return tgproc.MakeMessage(tgMessage.Chat.ID, "❌ нет зарегистрированных юзеров для этого канала")
	}
	msg := "💰 <b>Балансы</b>:"

	for i, Credit := range credits {
		msg += fmt.Sprintf("\n%d. %s ---> %d₽", i, Credit.UserID[1:], Credit.Credit)
	}
	metrics.TotalCreditsCommands.WithLabelValues("/bank").Add(1)
	return tgproc.MakeMessage(tgMessage.Chat.ID, msg)
}

func (bot tgbot) OweCancel(ctx context.Context, tgMessage *tgbotapi.Message) []tgbotapi.MessageConfig {
	channelID := tgMessage.Chat.ID
	if tgMessage.ReplyToMessage == nil {
		return tgproc.MakeMessage(channelID, "❌ сначала ты должен ответить на сообщение с /credit которое ты хочешь отменить")
	}
	if tgMessage.ReplyToMessage.Command() != "credit" {
		logger.Info(tgMessage.Command())
		return tgproc.MakeMessage(channelID, "❌ надо отметить сообщение с /credit, чтобы отменить долг")
	}
	if tgMessage.From.UserName != tgMessage.ReplyToMessage.From.UserName {
		return tgproc.MakeMessage(channelID, "❌ давай ты будешь отменять только свои долги, а то запахнет скамом")
	}

	buyer := entities.User{
		ID:        valueObjects.UserID("@" + tgMessage.From.UserName),
		ChannelID: valueObjects.ChannelID(tgMessage.Chat.ID),
	}

	args := tgMessage.ReplyToMessage.CommandArguments()

	payers, credit, err := bot.oweParser(args)
	if err != nil {
		// внутри все ошибки человеческие, но лучше поправить это
		return tgproc.MakeMessage(channelID, err.Error())
	}
	if len(payers) == 0 {
		return tgproc.MakeMessage(channelID, "❌ как так то, я не смог обнаружить у этого дела должников")
	}
	if credit == 0 {
		return tgproc.MakeMessage(channelID, "❌ камон, ты же знаешь, что я не мог зачесть нулевой долг")
	}

	for i, payer := range payers {
		for _, placeholder := range selfCreditPlaceholders {
			if payer == placeholder {
				payers[i] = buyer.ID
			}
		}
	}

	credit = -int64(math.Ceil(float64(credit) / float64(len(payers))))

	if err = bot.creditManager.CreditTransaction(ctx, valueObjects.ChannelID(channelID), buyer.ID, payers, valueObjects.Money(credit)); err != nil {
		if errors.As(err, &localerrors.UpdateCreditError{}) {
			updateErr := err.(localerrors.UpdateCreditError)
			if errors.Is(updateErr.Err, localerrors.ErrDoesntExist) {
				return tgproc.MakeMessage(channelID,
					fmt.Sprintf("❌ как %s может быть не зарегистрированым\nтранзакция провалена", updateErr.User))
			}
			return tgproc.MakeMessage(channelID, fmt.Sprintf("❌ не могу обновить счет %s\nтранзакция провалена", updateErr.User))
		}
	}
	return tgproc.MakeMessage(channelID, "✅ поздравляю, я отменил выделенный тобою долг, можешь гордиться мной")
}

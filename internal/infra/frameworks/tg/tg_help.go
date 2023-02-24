package tg

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	tgproc "gitlab.com/Sh00ty/dormitory_room_bot/pkg/tgproc"
)

func Help(ctx context.Context, msg *tgbotapi.Message) []tgbotapi.MessageConfig {
	args := msg.CommandArguments()

	switch args {
	case "credits":
		return tgproc.MakeMessage(msg.Chat.ID,
			"<b>/credit @usr1 @usr2 1000</b> - запоминает, что юзеры указанные после команды все вместе должны тому кто написал сумму указанную в конце."+
				"В данном примере @usr1 и @usr2 должны по 500 рублей тому, кто написал данную команду\n"+
				"Так же можно указывать себя среди должников, чтобы разделить сумму на которую вы скидываетесь. Помимо своего ника для этого можно использовать @i, @me и @я\n"+
				"\n<b>/bank</b> -  выводит баланс каждого участника чата\n"+
				"\n<b>/checkout</b> - выводит список транзакций между участниками чата, чтобы погасить долги созданные с помощью команды /credit\n"+
				"\n<b>/cancel</b> -  отменяет <i>/credit</i> на который надо ответить этой командой\n",
		)
	case "tasks":
		return tgproc.MakeMessage(msg.Chat.ID,
			"напишите за меня плиз хелп по таскам",
		)
	case "lists":
		return tgproc.MakeMessage(msg.Chat.ID,
			"напишите за меня плиз хелп по спискам",
		)
	default:
		return tgproc.MakeMessage(msg.Chat.ID,
			"<i>/help credits</i> - команды связанные с денежными операциями",
			"<i>/help tasks</i> - команды связанные с задачками и распределением работы",
			"<i>/help lists</i> - команды связанные со списками",
			"По появившемя вопросам пиши <i>@tw02h00ty</i>",
		)
	}

}

package credits

import (
	"sort"

	moneyalgo "github.com/Sh00ty/dormitory_room_bot/internal/interfaces/usecases/credits"
	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
	valueObjects "github.com/Sh00ty/dormitory_room_bot/pkg/value_objects"
)

type para struct {
	value valueObjects.Money
	key   valueObjects.UserID
}

func NewDomestosResolver() moneyalgo.Resolver {
	return domestosResolver{}
}

type domestosResolver struct {
}

func (domestosResolver) ResolveCredits(data []entities.Credit) []entities.Transaction {
	if len(data) == 0 {
		return nil
	}
	var ind1, ind2, ind3 int
	cID := data[0].ChannelID
	transactions := make([]entities.Transaction, 0, len(data)-1)
	neg := make([]para, 0, len(data)-1)
	pos := make([]para, 0, len(data)-1)
	for _, v := range data {
		if v.Credit < 0 {
			neg = append(neg, para{
				value: v.Credit,
				key:   v.UserID,
			})
		} else if v.Credit > 0 {
			pos = append(pos, para{
				value: v.Credit,
				key:   v.UserID,
			})
		}
	}
	sort.Slice(neg, func(i, j int) bool {
		return neg[i].value < neg[j].value
	})

	sort.Slice(pos, func(i, j int) bool {
		return pos[i].value > pos[j].value
	})
	for ind1 < len(neg) {
		transactions = append(transactions, entities.Transaction{})
		transactions[ind3].From = neg[ind1].key
		transactions[ind3].To = pos[ind2].key
		transactions[ind3].ChatID = cID
		if -neg[ind1].value == pos[ind2].value {
			transactions[ind3].Amount = pos[ind2].value
			ind1++
			ind2++
		} else if -neg[ind1].value > pos[ind2].value {
			transactions[ind3].Amount = pos[ind2].value
			neg[ind1].value += pos[ind2].value
			ind2++
		} else if -neg[ind1].value < pos[ind2].value {
			transactions[ind3].Amount = -neg[ind1].value
			pos[ind2].value += neg[ind1].value
			ind1++
		}
		ind3++
	}
	return transactions
}

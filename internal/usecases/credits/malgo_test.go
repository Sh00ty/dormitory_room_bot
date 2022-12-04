package credits

import (
	"fmt"
	"testing"

	"github.com/Sh00ty/dormitory_room_bot/pkg/entities"
)

var d domestosResolver

func TestMoney(t *testing.T) {
	m := []entities.Credit{
		{
			Credit:    1050,
			UserID:    "1",
			ChannelID: 1,
		},
		{
			Credit:    -1050,
			UserID:    "2",
			ChannelID: 1,
		},
	}
	fmt.Println(d.ResolveCredits(m))
}

func TestMoney1(t *testing.T) {
	m := []entities.Credit{
		{
			Credit:    1050,
			UserID:    "1",
			ChannelID: 1,
		},
		{
			Credit:    -1050,
			UserID:    "2",
			ChannelID: 1,
		},
		{
			Credit:    -20,
			UserID:    "3",
			ChannelID: 1,
		},
		{
			Credit:    20,
			UserID:    "4",
			ChannelID: 1,
		},
	}
	fmt.Println(d.ResolveCredits(m))
}

func TestMoney2(t *testing.T) {

	fmt.Println(d.ResolveCredits(nil))
}

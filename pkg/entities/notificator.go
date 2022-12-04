package entities

import (
	"fmt"
	"time"
)

var NullDate = time.Date(2002, time.September, 11, 11, 11, 11, 11, time.Local)

type notificator struct {
	// время когда нужно будет отправить уведомление о задаче
	NotifyTime time.Time

	// интервал повторения для того чтобы обновить notifyTime по его истечении
	NotifyInterval time.Duration
}

func CreateNotificator(notifyInterval *time.Duration, notifyTime *time.Time) *notificator {
	var n notificator

	if notifyTime != nil && notifyTime.After(time.Now()) {
		n.NotifyTime = *notifyTime
	} else {
		n.NotifyTime = time.Now()
	}

	if notifyInterval != nil {
		n.NotifyInterval = *notifyInterval
	}
	return &n
}

// Notificate если время для оповещения наступило, то возвращает true и сам обновляется
func (n *notificator) UpdateNotificationDate() {
	if n.NotifyTime.Equal(NullDate) {
		return
	}
	tn := time.Now()
	if tn.After(n.NotifyTime) {
		if n.NotifyInterval == 0 {
			n.NotifyTime = NullDate
			return
		}
		for n.NotifyTime.Before(tn) {
			n.NotifyTime = n.NotifyTime.Add(n.NotifyInterval)
		}
	}
}

func (n notificator) GetNextNotifyTime() (time.Time, bool) {
	if n.NotifyTime.Before(NullDate.Add(time.Hour)) {
		fmt.Println("false")
		return NullDate, false
	}
	return n.NotifyTime, true
}

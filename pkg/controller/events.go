package controller

type Subscription struct {
	SubId  string
	UserId string
	F      func(eventType string, id string)
}

func (this *Controller) Trigger(userid string, eventType string, id string) {
	this.subMux.Lock()
	defer this.subMux.Unlock()
	for _, sub := range this.subscriptions {
		if sub.UserId == userid {
			this.trigger(sub.F, eventType, id)
		}
	}
}

func (this *Controller) trigger(f func(eventType string, id string), eventType string, id string) {
	go f(eventType, id)
}

func (this *Controller) Unsubscribe(subId string) {
	this.subMux.Lock()
	defer this.subMux.Unlock()
	newList := []Subscription{}
	for _, sub := range this.subscriptions {
		if sub.SubId != subId {
			newList = append(newList, sub)
		}
	}
	this.subscriptions = newList
}

func (this *Controller) Subscribe(subId string, userId string, f func(eventType string, id string)) {
	this.subMux.Lock()
	defer this.subMux.Unlock()
	this.subscriptions = append(this.subscriptions, Subscription{
		SubId:  subId,
		UserId: userId,
		F:      f,
	})
}

package database

const (
	wantsToSubscribe = iota + 1
	timeAdded
	subscribed
)

type Subscription struct {
	ChatID    int     `bson:"chat_id"`
	Time      string  `bson:"time"`
	Longitude float64 `bson:"longitude"`
	Latitude  float64 `bson:"latitude"`
	UserState int     `bson:"user_state"`
}

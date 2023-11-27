package database

var newSubscription = Subscription{0, "", 0, 0}

type Subscription struct {
	ChatID    int     `bson:"chat_id"`
	Time      string  `bson:"time"`
	Longitude float64 `bson:"longitude"`
	Latitude  float64 `bson:"latitude"`
}

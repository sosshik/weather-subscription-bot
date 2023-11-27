package database

import (
	"context"
	"fmt"
	"time"

	weatherapi "git.foxminded.ua/foxstudent106264/task-2.5/pkg/weather_api"
	api "git.foxminded.ua/foxstudent106264/tgapi"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	Db            *mongo.Client
	DbName        string
	Collection    string
	Subscriptions map[int]*Subscription
}

func (d *Database) HandleSubscribeCommand(a *api.Api, update api.Update) {

	a.SendMessageWithLog("Enter your preferred notification time in the format HH:MM and send your location using Telegram's built-in function", update.Message.Chat.Id)

}

func (d *Database) HandleUnsubscribeCommand(a *api.Api, update api.Update) {
	subscription := Subscription{}
	collection := d.Db.Database(d.DbName).Collection(d.Collection)
	err := collection.FindOneAndDelete(context.Background(), bson.M{"chat_id": update.Message.Chat.Id}).Decode(&subscription)

	if err == nil {
		a.SendMessageWithLog("You have been unsubscribed.", update.Message.Chat.Id)
	} else {
		a.SendMessageWithLog("You are not subscribed.", update.Message.Chat.Id)
	}
}

func (d *Database) HandleUserInput(a *api.Api, update api.Update) {
	if update.Message.Text != "" {
		if _, ok := d.Subscriptions[update.Message.Chat.Id]; !ok {
			d.Subscriptions[update.Message.Chat.Id] = &Subscription{}
			d.Subscriptions[update.Message.Chat.Id].ChatID = update.Message.Chat.Id
		}
		_, err := time.Parse("15:04", update.Message.Text)
		if err != nil {
			a.SendMessageWithLog("Invalid time format. Please write /subscribe again and use HH:MM format.", update.Message.Chat.Id)
			return
		}

		d.Subscriptions[update.Message.Chat.Id].Time = update.Message.Text

		if newSubscription.Latitude == 0 && newSubscription.Longitude == 0 {
			a.SendMessageWithLog("Please send your location using Telegram's built-in function.", update.Message.Chat.Id)
		}

	} else if update.Message.Location.Longitude != 0 && update.Message.Location.Latitude != 0 {

		if _, ok := d.Subscriptions[update.Message.Chat.Id]; !ok {
			d.Subscriptions[update.Message.Chat.Id] = &Subscription{}
			d.Subscriptions[update.Message.Chat.Id].ChatID = update.Message.Chat.Id
		}

		d.Subscriptions[update.Message.Chat.Id].Longitude = update.Message.Location.Longitude
		d.Subscriptions[update.Message.Chat.Id].Latitude = update.Message.Location.Latitude

		if d.Subscriptions[update.Message.Chat.Id].Time == "" {
			a.SendMessageWithLog("Please your preferred notification time in the format HH:MM.", update.Message.Chat.Id)
		}

	}

	if d.Subscriptions[update.Message.Chat.Id].Time != "" && d.Subscriptions[update.Message.Chat.Id].Longitude != 0 && d.Subscriptions[update.Message.Chat.Id].Latitude != 0 {

		subscription := Subscription{}
		collection := d.Db.Database(d.DbName).Collection(d.Collection)
		err := collection.FindOne(context.Background(), bson.M{"chat_id": update.Message.Chat.Id}).Decode(&subscription)

		if err == nil {
			a.SendMessageWithLog(fmt.Sprintf("You are already subscribed. Time is set: %s", subscription.Time), update.Message.Chat.Id)
			return
		}

		_, err = collection.InsertOne(context.Background(), d.Subscriptions[update.Message.Chat.Id])
		if err != nil {
			log.Warnf("Unable to subscribe user with chat id%d: %s", update.Message.Chat.Id, err)
			a.SendMessageWithLog("Unable to subscribe you", update.Message.Chat.Id)
			return
		} else {
			log.Warnf("chat id %d was succesfully subscribed for time %s", update.Message.Chat.Id, newSubscription.Time)
			a.SendMessageWithLog(fmt.Sprintf("You succesfully subscribed for time %s", d.Subscriptions[update.Message.Chat.Id].Time), update.Message.Chat.Id)
		}
	}
}

func (d *Database) NotifyUser(a *api.Api, w weatherapi.WeatherAPI) {
	for {
		findOptions := options.Find()
		collection := d.Db.Database(d.DbName).Collection(d.Collection)
		cur, err := collection.Find(context.TODO(), bson.D{{}}, findOptions)
		if err != nil {
			log.Panic("database error: %w", err)
		}

		var subscriptions []Subscription

		for cur.Next(context.TODO()) {

			var elem Subscription
			err := cur.Decode(&elem)
			if err != nil {
				log.Warnf("unable to decode BSON at %s: %s", cur.Current, err)
				return
			}

			subscriptions = append(subscriptions, elem)
		}
		for _, sub := range subscriptions {
			notif, err := time.Parse("15:04", sub.Time)
			if err != nil {
				log.Warnf("unable to parse time for chat id %d: %s", sub.ChatID, err)
				continue
			}
			if time.Now().Hour() == notif.UTC().Hour() && time.Now().UTC().Minute() == notif.UTC().Minute() {

				message, err := w.GetForecast(sub.Latitude, sub.Longitude)
				if err != nil {
					log.Warnf("unable to get forecast for chat id %d: %s", sub.ChatID, err)
					continue
				}
				a.SendMessageWithLog(message, sub.ChatID)
			}
		}
		time.Sleep(59 * time.Second)
	}
}

package database

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	api "github.com/sosshik/tg-api"
	weatherapi "github.com/sosshik/weather-subscription-bot/pkg/weather_api"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	Db         *mongo.Client
	DbName     string
	Collection string
}

func ConnectionCheck(client *mongo.Client, clientOptions *options.ClientOptions) {
	for {
		time.Sleep(5 * time.Second)
		if err := client.Ping(context.Background(), nil); err != nil {
			log.Warnf("Lost connection to MongoDB. Attempting to reconnect.")
			err := client.Disconnect(context.Background())
			if err != nil {
				log.Warnf("Error while disconecting: %s", err)
				continue
			}
			client, err = mongo.Connect(context.Background(), clientOptions)
			if err != nil {
				log.Warnf("Failed to reconnect: %s", err)
			} else {
				log.Infof("Reconnected to MongoDB!")
			}
		}
	}
}

func (d *Database) HandleStartCommand(a *api.Api, update api.Update) {

	subscription := Subscription{}
	collection := d.Db.Database(d.DbName).Collection(d.Collection)
	err := collection.FindOne(context.Background(), bson.M{"chat_id": update.Message.Chat.Id}).Decode(&subscription)

	if err != nil {

		if err == mongo.ErrNoDocuments {
			keyboard := a.CreateKeyboard([]string{"/subscribe"})
			a.SendMessageAndKeyboardWithLog("*Hello, that's Weather Forecast Bot! Write /subscribe to subscribe for weather forecast.*", update.Message.Chat.Id, keyboard)
			return
		}

		keyboard := a.CreateKeyboard([]string{"/start"})
		a.SendMessageAndKeyboardWithLog("*Hello, that's Weather Forecast Bot! Error happend, please write /start again.*", update.Message.Chat.Id, keyboard)
		return

	}

	switch subscription.UserState {
	case wantsToSubscribe:
		keyboard := a.CreateKeyboard([]string{"/settime"})
		a.SendMessageAndKeyboardWithLog("*Hello, that's Weather Forecast Bot! You aready started subscription process. Please write /settime command and enter preferred notification time.*", update.Message.Chat.Id, keyboard)
		return
	case timeAdded:
		keyboard := a.CreateKeyboard([]string{"/setlocation"})
		a.SendMessageAndKeyboardWithLog("*Hello, that's Weather Forecast Bot! You aready started subscription process. Please write /setlocation command and enter send yoour loaction using Telegram's built-in function.*", update.Message.Chat.Id, keyboard)
		return
	case subscribed:
		keyboard := a.CreateKeyboard([]string{"/unsubscribe"})
		a.SendMessageAndKeyboardWithLog(fmt.Sprintf("*Hello, that's Weather Forecast Bot! You are already subscribed. Time is set: %s. If you want to unsubscribe write /unsubscribe command*", subscription.Time), update.Message.Chat.Id, keyboard)
		return
	}

}

func (d *Database) HandleSubscribeCommand(a *api.Api, update api.Update) {

	subscription := Subscription{}
	collection := d.Db.Database(d.DbName).Collection(d.Collection)
	err := collection.FindOne(context.Background(), bson.M{"chat_id": update.Message.Chat.Id}).Decode(&subscription)
	if err == nil {
		keyboard := a.CreateKeyboard([]string{"/unsubscribe"})
		a.SendMessageAndKeyboardWithLog(fmt.Sprintf("*You are already subscribed. Time is set: %s. If you want to unsubscribe write /unsubscribe command*", subscription.Time), update.Message.Chat.Id, keyboard)
		return
	}

	subscription.ChatID = update.Message.Chat.Id
	subscription.UserState = wantsToSubscribe
	_, err = collection.InsertOne(context.Background(), subscription)
	if err != nil {
		log.Warnf("unable to insert user into collection: %s", err)
		return
	}
	keyboard := a.CreateKeyboard([]string{"/settime"})
	a.SendMessageAndKeyboardWithLog("*Please write /settime command and after that enter your preferred notification time in the format HH:MM.*", update.Message.Chat.Id, keyboard)
}

func (d *Database) HandleSetTimeCommand(a *api.Api, update api.Update) {
	subscription := Subscription{}
	collection := d.Db.Database(d.DbName).Collection(d.Collection)
	err := collection.FindOne(context.Background(), bson.M{"chat_id": update.Message.Chat.Id}).Decode(&subscription)
	if err != nil {
		keyboard := a.CreateKeyboard([]string{"/subscribe"})
		a.SendMessageAndKeyboardWithLog("*You didn't start subscripton process. Please write /subscribe to start subscription*", update.Message.Chat.Id, keyboard)
		return
	}

	switch subscription.UserState {
	case wantsToSubscribe:
		a.SendMessageWithLog("*Please enter your preferred notification time in the format HH:MM.*", update.Message.Chat.Id)
		return
	case timeAdded:
		keyboard := a.CreateKeyboard([]string{"/setlocation"})
		a.SendMessageAndKeyboardWithLog(fmt.Sprintf("*You already set time: %s. Please write /setlocation and send your location to continue subscription process.*", subscription.Time), update.Message.Chat.Id, keyboard)
		return
	case subscribed:
		keyboard := a.CreateKeyboard([]string{"/unsubscribe"})
		a.SendMessageAndKeyboardWithLog(fmt.Sprintf("*You are already subscribed. Time is set: %s. If you want to unsubscribe write /unsubscribe command*", subscription.Time), update.Message.Chat.Id, keyboard)
		return
	}

}

func (d *Database) HandleSetLocationCommand(a *api.Api, update api.Update) {
	subscription := Subscription{}
	collection := d.Db.Database(d.DbName).Collection(d.Collection)
	err := collection.FindOne(context.Background(), bson.M{"chat_id": update.Message.Chat.Id}).Decode(&subscription)
	if err != nil {
		keyboard := a.CreateKeyboard([]string{"/subscribe"})
		a.SendMessageAndKeyboardWithLog("*You didn't start subscripton process. Please write /subscribe to start subscription*", update.Message.Chat.Id, keyboard)
		return
	}

	switch subscription.UserState {
	case wantsToSubscribe:
		keyboard := a.CreateKeyboard([]string{"/settime"})
		a.SendMessageAndKeyboardWithLog("*You didn't set the time yet. Please write /settime command to set the time.*", update.Message.Chat.Id, keyboard)
		return
	case timeAdded:
		a.SendMessageWithLog("*Please send your location using Telegram's built-in function.*", update.Message.Chat.Id)
		return
	case subscribed:
		keyboard := a.CreateKeyboard([]string{"/unsubscribe"})
		a.SendMessageAndKeyboardWithLog(fmt.Sprintf("*You are already subscribed. Time is set: %s. If you want to unsubscribe write /unsubscribe command*", subscription.Time), update.Message.Chat.Id, keyboard)
		return
	}

}

func (d *Database) HandleUnsubscribeCommand(a *api.Api, update api.Update) {
	subscription := Subscription{}
	collection := d.Db.Database(d.DbName).Collection(d.Collection)
	err := collection.FindOneAndDelete(context.Background(), bson.M{"chat_id": update.Message.Chat.Id}).Decode(&subscription)

	if err == nil {
		a.SendMessageWithLog("*You have been unsubscribed.*", update.Message.Chat.Id)
	} else {
		a.SendMessageWithLog("*You are not subscribed.*", update.Message.Chat.Id)
	}
}

func (d *Database) HandleUserInput(a *api.Api, update api.Update) {
	subscription := Subscription{}
	collection := d.Db.Database(d.DbName).Collection(d.Collection)
	err := collection.FindOne(context.Background(), bson.M{"chat_id": update.Message.Chat.Id}).Decode(&subscription)
	if err != nil {
		keyboard := a.CreateKeyboard([]string{"/subscribe"})
		a.SendMessageAndKeyboardWithLog("*You didn't start subscripton process. Please write /subscribe to start subscription*", update.Message.Chat.Id, keyboard)
		return
	}

	switch subscription.UserState {
	case wantsToSubscribe:
		_, err := time.Parse("15:04", update.Message.Text)
		if err != nil {
			a.SendMessageWithLog("*Invalid time format. Please write time again and use HH:MM format.*", update.Message.Chat.Id)
			return
		}
		subscription.Time = update.Message.Text
		subscription.UserState = timeAdded
		collection.FindOneAndReplace(context.Background(), bson.M{"chat_id": update.Message.Chat.Id}, subscription)
		keyboard := a.CreateKeyboard([]string{"/setlocation"})
		a.SendMessageAndKeyboardWithLog("*Time was set. Please write /setlocation and send your location using Telegram's built-in function.*", update.Message.Chat.Id, keyboard)
		return
	case timeAdded:
		if update.Message.Location.Longitude != 0 && update.Message.Location.Latitude != 0 {
			subscription.Longitude = update.Message.Location.Longitude
			subscription.Latitude = update.Message.Location.Latitude
			subscription.UserState = subscribed
			collection.FindOneAndReplace(context.Background(), bson.M{"chat_id": update.Message.Chat.Id}, subscription)
			keyboard := a.CreateKeyboard([]string{"/unsubscribe"})
			a.SendMessageAndKeyboardWithLog("*You successfully subscribed! If you want to unsubscribe write /unsubscribe command*", update.Message.Chat.Id, keyboard)
		}
		return
	case subscribed:
		keyboard := a.CreateKeyboard([]string{"/unsubscribe"})
		a.SendMessageAndKeyboardWithLog(fmt.Sprintf("*You are already subscribed. Time is set: %s. If you want to unsubscribe write /unsubscribe command*", subscription.Time), update.Message.Chat.Id, keyboard)
		return
	}
}

func (d *Database) NotifyUser(a *api.Api, w weatherapi.WeatherAPI) {
	for {
		findOptions := options.Find()
		collection := d.Db.Database(d.DbName).Collection(d.Collection)
		cur, err := collection.Find(context.TODO(), bson.D{{}}, findOptions)
		if err != nil {
			log.Warnf("database error: %s", err)
			return
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
			if sub.UserState == subscribed {
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
		}
		time.Sleep(59 * time.Second)
	}
}

## Overview
This is my solution to problem **2.5 Subscription bot**. This is a source code for Telegram Bot that have 2 commands `/subscribe` and `/unsubscribe`. When you write `/subscribe` you need to dial the time and location and bot will send you weather forecast every day at chosen time. `/unsubscribe` will delete your subscription from DB.
## How to run

To run this bot you need to get Token from the BotFather. You can read about it [here](https://core.telegram.org/bots/features#botfather).

Clone the repo: 

    git clone https://git.foxminded.ua/foxstudent106264/task-2.5.git

Create `.env` file with parameters: 
- `TELEGRAM_TOKEN` - the token that you got from BotFather
- `WEATHER_TOKEN` - the token from [Weather API](https://openweathermap.org/api)
- `MONGO_ADDR` - your MongoDB connection string
- `PORT` - port where you wish to start the bot
- `LOG_LEVEL` - log level of `logrus` logger, by default it's `info`

You can deploy your server any way you want, but I find it really quick and easy to use [ngrok](https://ngrok.com/download).

Once you install ngrok, you can run this command on another terminal on your system and copy forwarding link([result of the command looks like this](https://www.sohamkamani.com/golang/telegram-bot/ngrok-screenshot.png)):

    ngrok http 8080

Then set the webhooks:

    curl -F "url=<your forwarding link>/"  https://api.telegram.org/bot<your_api_token>/setWebhook

In the end just run the app:

    go run holiday_bot.go
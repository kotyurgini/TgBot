package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"tgbot/internal/ai/openai"
	"tgbot/internal/config"
	"tgbot/internal/lib/logger/handlers/fileslog"
	"tgbot/internal/lib/logger/handlers/multihandler"
	"tgbot/internal/lib/logger/handlers/myslog"
	"tgbot/internal/lib/logger/sl"
	"tgbot/internal/localization"
	"tgbot/internal/maincontroller"
	"tgbot/internal/store"
	"tgbot/internal/store/db"
	"tgbot/migrator"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// go run ./cmd/ --config-path "./configs/local.yaml"
func main() {
	cfg := config.MustLoad()
	log := setupMySlog()

	m, err := migrator.NewSqliteMigrator(cfg.StoragePath, log)
	if err != nil {
		panic(err)
	}
	m.MustMigrate()

	localization.MustLoadMessages(localization.LangEN)

	dbDriver, err := db.NewDBDriver(cfg)
	if err != nil {
		log.Error("Could not create dbDriver", sl.Err(err))
		return
	}

	bot, err := tgbotapi.NewBotAPI(cfg.TgToken)
	if err != nil {
		log.Error("Could not create tg bot", sl.Err(err))
		return
	}
	commandsRu := tgbotapi.NewSetMyCommandsWithScopeAndLanguage(tgbotapi.NewBotCommandScopeDefault(), "ru",
		tgbotapi.BotCommand{Command: "new", Description: "Начать новый дилог"},
		tgbotapi.BotCommand{Command: "dialogs", Description: "Список диалогов"},
		tgbotapi.BotCommand{Command: "profile", Description: "Ваш профиль"},
	)

	commandsEn := tgbotapi.NewSetMyCommandsWithScopeAndLanguage(tgbotapi.NewBotCommandScopeDefault(), "en",
		tgbotapi.BotCommand{Command: "new", Description: "Start new dialog"},
		tgbotapi.BotCommand{Command: "dialogs", Description: "List of dialogs"},
		tgbotapi.BotCommand{Command: "profile", Description: "Your profile"},
	)
	_, _ = bot.Send(commandsEn)
	_, _ = bot.Send(commandsRu)

	bc := context.Background()
	ctx, cancel := context.WithCancel(bc)

	aiAPI := openai.New("https://api.openai.com/v1", cfg.OpenAiToken, time.Minute)

	st, err := store.New(dbDriver)
	if err != nil {
		log.Error("Could not create store", sl.Err(err))
		return
	}

	_, err = maincontroller.New(ctx, bot, st, aiAPI, log, cfg.TgAdmin)
	if err != nil {
		log.Error("Could not create main handler", sl.Err(err))
		return
	}
	log.Info("Loaded")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	s := <-c
	log.Info("Signal received", slog.Attr{Key: "signal", Value: slog.StringValue(s.String())})
	cancel()
	_ = st.Close()
	os.Exit(0)
}

func setupMySlog() *slog.Logger {
	fileHandler := fileslog.NewFileSlogHandler("./data/logs")

	opts := myslog.MyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{Level: slog.LevelDebug},
	}

	handler := opts.NewMyHandler(os.Stdout)
	multi := multihandler.NewMultiHandler(handler, fileHandler)
	return slog.New(multi)
}

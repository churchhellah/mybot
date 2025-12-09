package main

import (
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// BotUser представляет пользователя бота
type BotUser struct {
	ID           int64
	Username     string
	FirstName    string
	JoinedAt     time.Time
	MessageCount int
}

// Интерфейс для обработчиков команд
type CommandHandler interface {
	Handle(update *tgbotapi.Update, bot *tgbotapi.BotAPI) error
}

// Обработчик команды /start
type StartCommand struct {
	users map[int64]*BotUser
}

func (c *StartCommand) Handle(update *tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	userID := update.Message.From.ID

	// Создаем или получаем пользователя
	user, exists := c.users[userID]
	if !exists {
		user = &BotUser{
			ID:           update.Message.From.ID,
			Username:     update.Message.From.UserName,
			FirstName:    update.Message.From.FirstName,
			JoinedAt:     time.Now(),
			MessageCount: 0,
		}
		c.users[userID] = user
	}
	welcomeMsg := formatWelcomeMessage(user.Username)
	return sendMessageSafely(bot, update.Message.Chat.ID, welcomeMsg)
}

// Обработчик команды /help
type HelpCommand struct{}

func (c *HelpCommand) Handle(update *tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	helpText := "Доступные команды:\n" +
		"/start - Начать работу с ботом\n" +
		"/help - Показать эту справку"
	return sendMessageSafely(bot, update.Message.Chat.ID, helpText)
}

func createUserFromMessage(msg *tgbotapi.Message) BotUser {
	return BotUser{
		ID:           msg.From.ID,
		Username:     msg.From.UserName,
		FirstName:    msg.From.FirstName,
		JoinedAt:     time.Now(),
		MessageCount: 0,
	}
}

func incrementMessageCount(user *BotUser) {
	user.MessageCount++
}

// formatWelcomeMessage форматирует приветственное сообщение для пользователя
// Функция определена на уровне пакета (вне main) - это правильный способ в Go
func formatWelcomeMessage(username string) string {
	if username == "" {
		return "Привет! Добро пожаловать в моего бота! 🤖"
	}
	return fmt.Sprintf("Привет, %s! Добро пожаловать в моего бота! 🤖", username)
}

func sendMessageSafely(bot *tgbotapi.BotAPI, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message to chat %d: %v", chatID, err)
		return fmt.Errorf("send message: %w", err)
	}
	return nil
}

func main() {
	// Загружаем переменные из .env файла (если он существует)
	// Игнорируем ошибку, если файл не найден - это нормально
	_ = godotenv.Load()

	// Получаем токен из переменной окружения (безопаснее, чем хардкод)
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Panic("TELEGRAM_BOT_TOKEN environment variable is not set. " +
			"Please set it in .env file or export it: export TELEGRAM_BOT_TOKEN=your_token")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Инициализация хранилища пользователей
	users := make(map[int64]*BotUser)

	// Регистрация обработчиков команд
	handlers := map[string]CommandHandler{
		"start": &StartCommand{users: users},
		"help":  &HelpCommand{},
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			// Обновляем счетчик сообщений
			userID := update.Message.From.ID
			if user, exists := users[userID]; exists {
				incrementMessageCount(user)
			}

			// Обработка команд
			command := update.Message.Command()
			if handler, exists := handlers[command]; exists {
				if err := handler.Handle(&update, bot); err != nil {
					log.Printf("Error handling command %s: %v", command, err)
				}
			} else if command != "" {
				// Неизвестная команда
				if err := sendMessageSafely(bot, update.Message.Chat.ID,
					"Неизвестная команда. Используйте /help для списка команд."); err != nil {
					log.Printf("Error: %v", err)
				}
			}
		}
	}
}

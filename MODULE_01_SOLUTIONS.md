# 💡 Решения заданий Модуля 1

> ⚠️ **Важно:** Сначала попробуйте решить задания самостоятельно! Используйте эти решения только для проверки или если застряли.

---

## Задание 1.1: Форматирование приветственного сообщения

```go
package main

import "fmt"

func formatWelcomeMessage(username string) string {
    if username == "" {
        return "Привет! Добро пожаловать!"
    }
    return fmt.Sprintf("Привет, %s! Добро пожаловать в моего бота! 🤖", username)
}

// Использование в main:
case "start":
    username := update.Message.From.UserName
    welcomeMsg := formatWelcomeMessage(username)
    msg = tgbotapi.NewMessage(update.Message.Chat.ID, welcomeMsg)
```

**Что изучили:**
- Функции с параметрами и возвращаемыми значениями
- `fmt.Sprintf` для форматирования строк
- Условные операторы

---

## Задание 1.2: Структура BotUser

```go
package main

import (
    "time"
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotUser struct {
    ID        int64
    Username  string
    FirstName string
    JoinedAt  time.Time
}

func createUserFromMessage(msg *tgbotapi.Message) BotUser {
    return BotUser{
        ID:        msg.From.ID,
        Username:  msg.From.UserName,
        FirstName: msg.From.FirstName,
        JoinedAt:  time.Now(),
    }
}

// Использование:
case "start":
    user := createUserFromMessage(update.Message)
    log.Printf("New user joined: %+v", user)
    msg = tgbotapi.NewMessage(update.Message.Chat.ID, 
        fmt.Sprintf("Привет, %s!", user.FirstName))
```

**Что изучили:**
- Определение структур
- Создание экземпляров структур
- Работа с пакетом `time`

---

## Задание 1.3: Счетчик сообщений с указателями

```go
type BotUser struct {
    ID           int64
    Username     string
    FirstName    string
    JoinedAt     time.Time
    MessageCount int // новое поле
}

func incrementMessageCount(user *BotUser) {
    user.MessageCount++
}

// Использование:
// Создаем map для хранения пользователей
users := make(map[int64]*BotUser)

// При получении сообщения:
userID := update.Message.From.ID
user, exists := users[userID]

if !exists {
    newUser := createUserFromMessage(update.Message)
    users[userID] = &newUser
    user = users[userID]
}

incrementMessageCount(user)
log.Printf("User %s sent %d messages", user.Username, user.MessageCount)
```

**Что изучили:**
- Указатели на структуры
- Изменение значений через указатели
- Работа с `map` в Go

---

## Задание 1.4: Интерфейс CommandHandler

```go
package main

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// Интерфейс для обработчиков команд
type CommandHandler interface {
    Handle(update *tgbotapi.Update, bot *tgbotapi.BotAPI) error
}

// Обработчик команды /start
type StartCommand struct{}

func (c *StartCommand) Handle(update *tgbotapi.Update, bot *tgbotapi.BotAPI) error {
    username := update.Message.From.UserName
    welcomeMsg := formatWelcomeMessage(username)
    msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcomeMsg)
    _, err := bot.Send(msg)
    return err
}

// Обработчик команды /help
type HelpCommand struct{}

func (c *HelpCommand) Handle(update *tgbotapi.Update, bot *tgbotapi.BotAPI) error {
    helpText := "Доступные команды:\n" +
        "/start - Начать работу с ботом\n" +
        "/help - Показать эту справку"
    msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpText)
    _, err := bot.Send(msg)
    return err
}

// Использование в main:
handlers := map[string]CommandHandler{
    "start": &StartCommand{},
    "help":  &HelpCommand{},
}

command := update.Message.Command()
if handler, exists := handlers[command]; exists {
    if err := handler.Handle(update, bot); err != nil {
        log.Printf("Error handling command %s: %v", command, err)
    }
} else {
    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда")
    bot.Send(msg)
}
```

**Что изучили:**
- Определение интерфейсов
- Неявная реализация интерфейсов
- Использование интерфейсов для полиморфизма
- Методы структур (receivers)

---

## Задание 1.5: Улучшенная обработка ошибок

```go
func sendMessageSafely(bot *tgbotapi.BotAPI, chatID int64, text string) error {
    msg := tgbotapi.NewMessage(chatID, text)
    _, err := bot.Send(msg)
    if err != nil {
        log.Printf("Failed to send message to chat %d: %v", chatID, err)
        return fmt.Errorf("send message: %w", err)
    }
    return nil
}

// Использование:
case "start":
    username := update.Message.From.UserName
    welcomeMsg := formatWelcomeMessage(username)
    if err := sendMessageSafely(bot, update.Message.Chat.ID, welcomeMsg); err != nil {
        log.Printf("Error: %v", err)
        // Можно попробовать отправить сообщение об ошибке или просто залогировать
    }
```

**Что изучили:**
- Правильная обработка ошибок
- `fmt.Errorf` с `%w` для обертывания ошибок
- Логирование вместо паники

---

## 🎯 Итоговое решение: Рефакторинг бота

```go
package main

import (
    "fmt"
    "log"
    "os"
    "time"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BotUser представляет пользователя бота
type BotUser struct {
    ID           int64
    Username     string
    FirstName    string
    JoinedAt     time.Time
    MessageCount int
}

// CommandHandler интерфейс для обработчиков команд
type CommandHandler interface {
    Handle(update *tgbotapi.Update, bot *tgbotapi.BotAPI) error
}

// StartCommand обрабатывает команду /start
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

// HelpCommand обрабатывает команду /help
type HelpCommand struct{}

func (c *HelpCommand) Handle(update *tgbotapi.Update, bot *tgbotapi.BotAPI) error {
    helpText := "Доступные команды:\n" +
        "/start - Начать работу с ботом\n" +
        "/help - Показать эту справку"
    return sendMessageSafely(bot, update.Message.Chat.ID, helpText)
}

// Вспомогательные функции
func formatWelcomeMessage(username string) string {
    if username == "" {
        return "Привет! Добро пожаловать!"
    }
    return fmt.Sprintf("Привет, %s! Добро пожаловать в моего бота! 🤖", username)
}

func sendMessageSafely(bot *tgbotapi.BotAPI, chatID int64, text string) error {
    msg := tgbotapi.NewMessage(chatID, text)
    _, err := bot.Send(msg)
    if err != nil {
        return fmt.Errorf("send message: %w", err)
    }
    return nil
}

func incrementMessageCount(user *BotUser) {
    user.MessageCount++
}

func main() {
    token := os.Getenv("TELEGRAM_BOT_TOKEN")
    if token == "" {
        log.Panic("TELEGRAM_BOT_TOKEN environment variable is not set")
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
                if err := handler.Handle(update, bot); err != nil {
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
```

**Что мы сделали:**
- ✅ Использовали структуры для данных
- ✅ Создали интерфейс для обработчиков команд
- ✅ Правильно обработали все ошибки
- ✅ Разделили код на логические функции
- ✅ Использовали указатели где необходимо
- ✅ Код стал более читаемым и расширяемым

---

## 🎓 Что дальше?

После выполнения всех заданий модуля 1 вы должны понимать:
- Как работают переменные и функции в Go
- Что такое структуры и как их использовать
- Зачем нужны указатели
- Как работают интерфейсы
- Как правильно обрабатывать ошибки

**Готовы к следующему модулю?** Переходите к изучению пакетов и модулей! 🚀


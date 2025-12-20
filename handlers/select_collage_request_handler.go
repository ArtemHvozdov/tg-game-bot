package handlers

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ArtemHvozdov/tg-game-bot.git/models"
	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"gopkg.in/telebot.v3"
)

// Global variable to store current admin working chat
var currentAdminWorkingChat = make(map[int64]int64) // adminID -> chatID

func SelectRequestHandler(bot *telebot.Bot) func(c telebot.Context) error {
    return func(c telebot.Context) error {
        user := c.Sender()
        chat := c.Chat()
        
        adminsIds := cfg.AdminIDs
        isAdminUser := utils.IsAdminByBot(adminsIds, user.ID)
        if !isAdminUser {
            msg := "Вибачте, ця команда доступна лише адміністраторам."
            bot.Send(chat, msg)
            return nil
        }

        // Get chat ID from command argument
        args := c.Args()
        if len(args) == 0 {
            msg := "Використання: /select_request <chat_id>\nПриклад: /select_request -1002617613395"
            bot.Send(chat, msg)
            return nil
        }

        // Parse chat ID from argument
        selectedChatID, err := strconv.ParseInt(args[0], 10, 64)
        if err != nil {
            msg := "Неправильний chat ID. Використовуйте число."
            bot.Send(chat, msg)
            return nil
        }

        // Get game ID by chat ID
        game, err := storage_db.GetGameByChatId(selectedChatID)
        if err != nil {
            utils.Logger.Errorf("Error getting game by chat ID %d: %v", selectedChatID, err)
            msg := "Не знайдено гру для вказаного chat ID."
            bot.Send(chat, msg)
            return nil
        }

        // Verify that this chat ID exists in collage requests
        requests, err := storage_db.GetAllCollageRequests()
        if err != nil {
            utils.Logger.Errorf("Error getting collage requests: %v", err)
            return nil
        }

        // Check if selected chat ID exists in requests
        var requestExists bool
        for _, request := range requests {
            if request.ChatID == selectedChatID {
                requestExists = true
                break
            }
        }

        if !requestExists {
            msg := fmt.Sprintf("Chat ID %d не знайдено у списку запитів.", selectedChatID)
            bot.Send(chat, msg)
            return nil
        }
        
        // Save current admin working chat
        currentAdminWorkingChat[user.ID] = selectedChatID

        // Send images from folder
        err = sendImagesFromFolder(bot, chat, selectedChatID)
        if err != nil {
            utils.Logger.Errorf("Error sending images: %v", err)
            msg := "Помилка при відправці зображень."
            bot.Send(chat, msg)
            return nil
        }

        // Update request status to "processing"
        err = storage_db.UpdateCollageRequestStatus(int64(game.ID), selectedChatID, models.StatusReqCollageProcessing)
        if err != nil {
            utils.Logger.Errorf("Error updating request status: %v", err)
        }

        msg := fmt.Sprintf("Зображення з чату %d відправлені. Тепер ви можете надіслати готовий колаж.", selectedChatID)
        bot.Send(chat, msg)

        return nil
    }
}

func sendImagesFromFolder(bot *telebot.Bot, adminChat *telebot.Chat, chatID int64) error {
    // Remove minus sign from chatID for folder path
    chatIDStr := strings.TrimPrefix(strconv.FormatInt(chatID, 10), "-")
    folderPath := fmt.Sprintf("./storage_temp/%s", chatIDStr)

    // Get list of all jpg files in folder
    files, err := filepath.Glob(filepath.Join(folderPath, "*.jpg"))
    if err != nil {
        return fmt.Errorf("failed to read folder %s: %w", folderPath, err)
    }

    if len(files) == 0 {
        return fmt.Errorf("no images found in folder %s", folderPath)
    }

    // Create album from images
    var album telebot.Album
    for _, filePath := range files {
        photo := &telebot.Photo{File: telebot.FromDisk(filePath)}
        album = append(album, photo)
    }

    // Send album
    _, err = bot.SendAlbum(adminChat, album)
    if err != nil {
        return fmt.Errorf("failed to send album: %w", err)
    }

    utils.Logger.Infof("Sent %d images from folder %s to admin chat", len(files), folderPath)
    return nil
}
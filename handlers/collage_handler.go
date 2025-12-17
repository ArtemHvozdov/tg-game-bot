// Collage handler for subtask 10 results
package handlers

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/ArtemHvozdov/tg-game-bot.git/storage_db"
	"github.com/ArtemHvozdov/tg-game-bot.git/utils"
	"github.com/fogleman/gg"
	xdraw "golang.org/x/image/draw"
	"gopkg.in/telebot.v3"
)

// CreateSubtask10Collage creates a collage from subtask 10 winning images
// Variant with context handler for calling in the chat
// func CreateSubtask10Collage(bot *telebot.Bot) func(c telebot.Context) error {
// 	return func(c telebot.Context) error {
// 		chat := c.Chat()

// 		// Get game by chat ID
// 		game, err := storage_db.GetGameByChatId(chat.ID)
// 		if err != nil {
// 			utils.Logger.Errorf("Failed to get game by chat ID %d: %v", chat.ID, err)
// 			return c.Send("Помилка отримання гри")
// 		}

// 		// Inform user
// 		c.Send("Чекайте-чекайте, меджик у процесі … 🧚✨")

// 		// Get winners array
// 		winners, err := storage_db.GetSubtask10WinnersArray(game.ID)
// 		if err != nil {
// 			utils.Logger.Errorf("Failed to get subtask 10 winners for game %d: %v", game.ID, err)
// 			return c.Send(fmt.Sprintf("❌ Помилка отримання результатів: %v", err))
// 		}

// 		// Generate collage
// 		err = CreateSubtask10CollageWithGG(winners)
// 		if err != nil {
// 			return c.Send(fmt.Sprintf("❌ Помилка створення колажу: %v", err))
// 		}

// 		// Check if file exists
// 		if _, err := os.Stat("subtask10_collage.jpg"); os.IsNotExist(err) {
// 			return c.Send("❌ Файл колажу не був створений")
// 		}
		
// 		// Send as photo
// 		photo := &telebot.Photo{
// 			File:    telebot.FromDisk("subtask10_collage.jpg"),
// 			Caption: "Готово! Ловіть колаж із відповідей, які набрали найбільшу кількість голосів. Схоже на те, що подобається вашій гьорлз бенд? 💅",
// 		}
// 		_, err = bot.Send(chat, photo)
// 		if err != nil {
// 			return c.Send(fmt.Sprintf("❌ Помилка відправки колажу: %v", err))
// 		}

// 		// Send as document (original without compression)
// 		document := &telebot.Document{
// 			File:     telebot.FromDisk("subtask10_collage.jpg"),
// 			MIME:     "image/jpeg",
// 			FileName: "subtask10_collage_2160x2160.jpg",
// 			Caption:  "Cупер висока якість для моїх aesthetic girls 🎀",
// 		}
// 		bot.Send(chat, document)

// 		// Clean up temporary file
// 		go func() {
// 			time.Sleep(5 * time.Second)
// 			os.Remove("subtask10_collage.jpg")
// 			fmt.Println("Temporary subtask 10 collage file cleaned up")
// 		}()

// 		return nil
// 	}
// }

// Variant withouut context for calling as external function
func CreateSubtask10Collage(bot *telebot.Bot, chatID int64) error {
	chat := &telebot.Chat{ID: chatID}
	
	// Get game by chat ID
	game, err := storage_db.GetGameByChatId(chat.ID)
	if err != nil {
		utils.Logger.Errorf("Failed to get game by chat ID %d: %v", chat.ID, err)
		return err
	}

	_, err = bot.Send(chat, "Чекайте-чекайте, меджик у процесі … 🧚✨")
	if err != nil {
		utils.Logger.Errorf("Failed to send waiting message to chat %d: %v", chat.ID, err)
	}

	// Get winners array
	winners, err := storage_db.GetSubtask10WinnersArray(game.ID)
	if err != nil {
		utils.Logger.Errorf("Failed to get subtask 10 winners for game %d: %v", game.ID, err)
		return err
	}

	// Generate collage
	err = CreateSubtask10CollageWithGG(winners)
	if err != nil {
		_, err = bot.Send(chat, "Упс... Щось я не можу створити коллаж. Що ж робити?")
		utils.Logger.Errorf("Failed to create subtask 10 collage: %v", err)
		return err
	}

	// Check if file exists
	if _, err := os.Stat("subtask10_collage.jpg"); os.IsNotExist(err) {
		utils.Logger.Errorf("Subtask 10 collage file does not exist after creation")
		return err
	}

	// Send as photo
	photo := &telebot.Photo{
		File:    telebot.FromDisk("subtask10_collage.jpg"),
		Caption: "Готово! Ловіть колаж із відповідей, які набрали найбільшу кількість голосів. Схоже на те, що подобається вашій гьорлз бенд? 💅",
	}
	_, err = bot.Send(chat, photo)
	if err != nil {
		utils.Logger.Errorf("Failed to send subtask 10 collage photo: %v", err)
		return err
	}

	// Send as document (original without compression)
	document := &telebot.Document{
		File:     telebot.FromDisk("subtask10_collage.jpg"),
		MIME:     "image/jpeg",
		FileName: "subtask10_collage_2160x2160.jpg",
		Caption:  "Cупер висока якість для моїх aesthetic girls 🎀",
	}
	bot.Send(chat, document)

	// Clean up temporary file
	go func() {
		time.Sleep(5 * time.Second)
		os.Remove("subtask10_collage.jpg")
		fmt.Println("Temporary subtask 10 collage file cleaned up")
	}()

	return nil
}


// CreateSubtask10CollageWithGG creates a 3×3 collage from winning images
func CreateSubtask10CollageWithGG(winners []string) error {
	fmt.Println("=== START CREATE SUBTASK 10 COLLAGE 2160×2160 ===")

	// Build full paths to images
	basePath := "internal/data/tasks/subtasks/subtask_10/assets_collage"
	var imagePaths []string

	for _, winner := range winners {
		fullPath := filepath.Join(basePath, winner)
		// Check if file exists
		if _, err := os.Stat(fullPath); err == nil {
			imagePaths = append(imagePaths, fullPath)
			fmt.Printf("Found winner image: %s\n", fullPath)
		} else {
			fmt.Printf("WARNING: Winner image not found: %s\n", fullPath)
			utils.Logger.Warnf("Winner image not found: %s", fullPath)
		}
	}

	if len(imagePaths) == 0 {
		return fmt.Errorf("no winner images found in %s", basePath)
	}

	// Limit to 9 images max (3×3 grid)
	if len(imagePaths) > 9 {
		imagePaths = imagePaths[:9]
	}

	fmt.Printf("Using %d images for subtask 10 collage\n", len(imagePaths))

	// Grid settings
	const (
		cellSize   = 720  // each image size
		gridSize   = 3    // 3×3 grid
		canvasSize = cellSize * gridSize // 2160×2160
	)

	// Create canvas
	dc := gg.NewContext(canvasSize, canvasSize)
	dc.SetRGB(0.05, 0.05, 0.05) // very dark background
	dc.Clear()

	// Draw images in 3×3 grid
	for i, path := range imagePaths {
		if i >= 9 {
			break
		}

		fmt.Printf("Processing subtask 10 image %d: %s\n", i+1, path)

		img, err := gg.LoadImage(path)
		if err != nil {
			fmt.Printf("ERROR loading %s: %v\n", path, err)
			continue
		}

		// Fit image into 720×720 square
		processed := fitImageToTileSubtask10(img, cellSize, cellSize)

		row := i / gridSize
		col := i % gridSize
		x := col * cellSize
		y := row * cellSize

		dc.DrawImage(processed, x, y)
	}

	// Save JPEG
	outFile, err := os.Create("subtask10_collage.jpg")
	if err != nil {
		return fmt.Errorf("failed to create collage file: %w", err)
	}
	defer outFile.Close()

	err = jpeg.Encode(outFile, dc.Image(), &jpeg.Options{Quality: 95})
	if err != nil {
		return fmt.Errorf("failed to encode subtask 10 collage jpeg: %w", err)
	}

	fmt.Println("=== SUBTASK 10 COLLAGE SAVED AS subtask10_collage.jpg (2160×2160) ===")
	return nil
}

// fitImageToTileSubtask10 fits image into cellSize×cellSize square
func fitImageToTileSubtask10(img image.Image, targetW, targetH int) image.Image {
	b := img.Bounds()
	iw, ih := b.Dx(), b.Dy()

	scale := math.Min(float64(targetW)/float64(iw), float64(targetH)/float64(ih)) * 0.98
	nw := int(float64(iw) * scale)
	nh := int(float64(ih) * scale)

	// Resize image
	resized := image.NewRGBA(image.Rect(0, 0, nw, nh))
	xdraw.CatmullRom.Scale(resized, resized.Bounds(), img, b, xdraw.Over, nil)

	// Create final square with background
	tile := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	bg := image.NewUniform(color.RGBA{255, 255, 255, 255}) // white background
	xdraw.Draw(tile, tile.Bounds(), bg, image.Point{}, xdraw.Src)

	// Center the image
	off := image.Pt((targetW-nw)/2, (targetH-nh)/2)
	xdraw.Draw(tile, resized.Bounds().Add(off), resized, image.Point{}, xdraw.Over)

	return tile
}
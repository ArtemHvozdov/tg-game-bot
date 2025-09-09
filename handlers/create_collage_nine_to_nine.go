package handlers

import (
	"fmt"
	"image"
	"image/jpeg"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/fogleman/gg"
	"gopkg.in/telebot.v3"
)

func CreateCollageFromResultsImageNine(bot *telebot.Bot) func(c telebot.Context) error {
	return func(c telebot.Context) error {
		chat := c.Chat()
		
		// Inform user that collage creation started
		c.Send("🎨 Створюю колаж...")
		
		// Create the collage
		err := CreateCollageWithGGNine1()
		if err != nil {
			return c.Send(fmt.Sprintf("❌ Помилка створення колажу: %v", err))
		}
		
		// Check if collage file was created
		if _, err := os.Stat("collage.jpg"); os.IsNotExist(err) {
			return c.Send("❌ Файл колажу не був створений")
		}
		
		// Send as photo
		photo := &telebot.Photo{
			File:    telebot.FromDisk("collage.jpg"),
			Caption: "🎨 Ваш колаж готовий!\n\n📱 Ідеально підходить для обоїв телефону",
		}
		
		_, err = bot.Send(chat, photo)
		if err != nil {
			return c.Send(fmt.Sprintf("❌ Помилка відправки колажу: %v", err))
		}
		
		// Also send as document for highest quality
		document := &telebot.Document{
			File:     telebot.FromDisk("collage.jpg"),
			MIME:     "image/jpeg",
			FileName: "wallpaper_collage.jpg",
			Caption:  "📱 Високоякісна версія для завантаження як обої",
		}
		
		bot.Send(chat, document)
		
		// Clean up - remove the collage file after sending
		go func() {
			time.Sleep(5 * time.Second)
			os.Remove("collage.jpg")
			fmt.Println("Temporary collage file cleaned up")
		}()
		
		return nil
	}
}

func CreateCollageWithGGNine1() error {
    fmt.Println("=== START CREATE COLLAGE ===")
    
    // Find all images in results folder
    imagePaths, err := filepath.Glob("results_4/*.jpg")
    if err != nil {
        return fmt.Errorf("error reading results folder: %w", err)
    }
    
    if len(imagePaths) == 0 {
        return fmt.Errorf("no images found in results folder")
    }
    
    // Limit to 9 images
    if len(imagePaths) > 9 {
        imagePaths = imagePaths[:9]
    }
    
    fmt.Printf("Finded %d images for collage\n", len(imagePaths))
    
    // Image cell sizes for 3x3 grid
    imageWidth := 360   // reduced from 540 to fit 3 columns
    imageHeight := 640  // keep same height for better proportions
    
    totalWidth := imageWidth * 3    // 1080
    totalHeight := imageHeight * 3  // 1920
    
    fmt.Printf("Create canvas: %dx%d\n", totalWidth, totalHeight)
    
    dc := gg.NewContext(totalWidth, totalHeight)
    
    // Very dark background
    dc.SetRGB(0.05, 0.05, 0.05)
    dc.Clear()
    
    // Positions for 3x3 grid - NO MARGINS
    positions := [][2]int{
        {0, 0},                                    // row 1, col 1
        {imageWidth, 0},                           // row 1, col 2
        {imageWidth * 2, 0},                       // row 1, col 3
        {0, imageHeight},                          // row 2, col 1
        {imageWidth, imageHeight},                 // row 2, col 2
        {imageWidth * 2, imageHeight},             // row 2, col 3
        {0, imageHeight * 2},                      // row 3, col 1
        {imageWidth, imageHeight * 2},             // row 3, col 2
        {imageWidth * 2, imageHeight * 2},         // row 3, col 3
    }
    
    // Process each image
    for i, path := range imagePaths {
        if i >= 9 { break }
        
        fmt.Printf("Image processing %d: %s\n", i+1, path)
        
        img, err := gg.LoadImage(path)
        if err != nil {
            fmt.Printf("ERROR download of image %s: %v\n", path, err)
            continue
        }
        
        // Enlarging the image scale
        processedImg := showBiggerImageWithLessMargin(img, imageWidth, imageHeight)
        
        x := positions[i][0]
        y := positions[i][1]
        
        fmt.Printf("Placing an image in position: %d, %d\n", x, y)
        
        dc.DrawImageAnchored(processedImg, x+imageWidth/2, y+imageHeight/2, 0.5, 0.5)
    }
    
    // Reduced text by 20%
    centerX := totalWidth / 2
    centerY := totalHeight / 2
    
    fmt.Println("Add text (reduced by 20%)...")
    
    // We draw the text graphically - each letter separately
    drawSmallerText(dc, centerX, centerY)
    
    fmt.Println("Save file...")
    
    // Save as JPEG
    img := dc.Image()
    
    file, err := os.Create("collage.jpg")
    if err != nil {
        return fmt.Errorf("failed to create collage file: %w", err)
    }
    defer file.Close()
    
    err = jpeg.Encode(file, img, &jpeg.Options{Quality: 95})
    if err != nil {
        return fmt.Errorf("failed to encode JPEG: %w", err)
    }
    
    fmt.Println("=== COLLAGE SAVED AS collage.jpg ===")
    return nil
}

// Show MORE of the image with less padding
func showBiggerImageWithLessMarginNine(img image.Image, targetWidth, targetHeight int) image.Image {
    bounds := img.Bounds()
    imgWidth := bounds.Dx()
    imgHeight := bounds.Dy()
    
    fmt.Printf("Original size: %dx%d\n", imgWidth, imgHeight)
    
    // Calculate the scale so that the image fills more space
    scaleX := float64(targetWidth) / float64(imgWidth)
    scaleY := float64(targetHeight) / float64(imgHeight)
    
    // Increase fill from 95% to 98% for large images
    scale := math.Min(scaleX, scaleY) * 0.98
    
    fmt.Printf("Zoom in: %.2f\n", scale)
    
    newWidth := int(float64(imgWidth) * scale)
    newHeight := int(float64(imgHeight) * scale)
    
    fmt.Printf("Zoom size: %dx%d\n", newWidth, newHeight)
    
    // Create context scope size
    dc := gg.NewContext(targetWidth, targetHeight)
    
    // Darker background for less contrast with padding
    dc.SetRGB(0.08, 0.08, 0.08)
    dc.Clear()
    
    // Center the image
    offsetX := (targetWidth - newWidth) / 2
    offsetY := (targetHeight - newHeight) / 2
    
    fmt.Printf("Bias: %d, %d\n", offsetX, offsetY)
    
    // Draw a scaled image Offset
    dc.DrawImageAnchored(img, offsetX+newWidth/2, offsetY+newHeight/2, 0.5, 0.5)
    
    return dc.Image()
}

func drawSmallerTextNine(dc *gg.Context, centerX, centerY int) {
    // Sizes for text
    letterHeight := 90.0  // old:150
    letterWidth := 50.0    // old: 100
    spacing := 14.0        // old: 20
    
    // text "MEMORIES"
    text := "MEMORIES"
    totalWidth := float64(len(text)) * (letterWidth + spacing) - spacing
    
    startX := float64(centerX) - totalWidth/2
    startY := float64(centerY)
    
    // Reduced background for text
    bgPadding := 45.0  // old 30
    dc.SetRGBA(0, 0, 0, 0.4)
    dc.DrawRoundedRectangle(
        startX - bgPadding,
        startY - letterHeight/2 - bgPadding,
        totalWidth + bgPadding*2,
        letterHeight + bgPadding*2,
        16,  // old 20
    )
    dc.Fill()
    
    // draw each letter as simple lines
    dc.SetRGB(1, 1, 1) // White color
    dc.SetLineWidth(6)  // old 8
    
    for i, char := range text {
        x := startX + float64(i) * (letterWidth + spacing)
        drawLetter(dc, char, x, startY, letterWidth, letterHeight)
    }
}


// draw each letter as simple lines
func drawLetterNine(dc *gg.Context, char rune, x, y, width, height float64) {
    halfWidth := width / 2
    halfHeight := height / 2
    
    switch char {
    case 'M':
        // letter M
        dc.DrawLine(x-halfWidth, y+halfHeight, x-halfWidth, y-halfHeight) 
        dc.DrawLine(x+halfWidth, y+halfHeight, x+halfWidth, y-halfHeight) 
        dc.DrawLine(x-halfWidth, y-halfHeight, x, y) 
        dc.DrawLine(x, y, x+halfWidth, y-halfHeight) 
        
    case 'E':
        // letter E
        dc.DrawLine(x-halfWidth, y-halfHeight, x-halfWidth, y+halfHeight) 
        dc.DrawLine(x-halfWidth, y-halfHeight, x+halfWidth, y-halfHeight)
        dc.DrawLine(x-halfWidth, y, x+halfWidth*0.7, y)
        dc.DrawLine(x-halfWidth, y+halfHeight, x+halfWidth, y+halfHeight)
        dc.Stroke()
        
    case 'O':
        // letter O - овал
        dc.DrawEllipse(x, y, halfWidth, halfHeight)
        dc.Stroke()
        
    case 'R':
        // letter R
        dc.DrawLine(x-halfWidth, y-halfHeight, x-halfWidth, y+halfHeight)
        dc.DrawLine(x-halfWidth, y-halfHeight, x+halfWidth, y-halfHeight)
        dc.DrawLine(x+halfWidth, y-halfHeight, x+halfWidth, y)
        dc.DrawLine(x+halfWidth, y, x-halfWidth, y)
        dc.DrawLine(x, y, x+halfWidth, y+halfHeight)
        dc.Stroke()
        
    case 'I':
        // letter I
        dc.DrawLine(x, y-halfHeight, x, y+halfHeight)
        dc.DrawLine(x-halfWidth*0.5, y-halfHeight, x+halfWidth*0.5, y-halfHeight)
        dc.DrawLine(x-halfWidth*0.5, y+halfHeight, x+halfWidth*0.5, y+halfHeight)
        dc.Stroke()
        
    case 'S':
        // letter S 
        dc.DrawLine(x+halfWidth, y-halfHeight, x-halfWidth, y-halfHeight)
        dc.DrawLine(x-halfWidth, y-halfHeight, x-halfWidth, y)
        dc.DrawLine(x-halfWidth, y, x+halfWidth, y)
        dc.DrawLine(x+halfWidth, y, x+halfWidth, y+halfHeight)
        dc.DrawLine(x+halfWidth, y+halfHeight, x-halfWidth, y+halfHeight)
        dc.Stroke()
        
    default:
        // Unknown letter - draw a rectangle
        dc.DrawRectangle(x-halfWidth, y-halfHeight, width, height)
        dc.Stroke()
    }
}


// Helper function to process image for square format
func processImageForSquareNine(img image.Image, targetSize int) image.Image {
    bounds := img.Bounds()
    imgWidth := bounds.Dx()
    imgHeight := bounds.Dy()
    
    // Create new square context
    dc := gg.NewContext(targetSize, targetSize)
    dc.SetRGB(1, 1, 1) // white background
    dc.Clear()
    
    // Calculate scaling to fit image in square while maintaining aspect ratio
    scale := math.Min(float64(targetSize)/float64(imgWidth), float64(targetSize)/float64(imgHeight))*0.7
    
    newWidth := int(float64(imgWidth) * scale)
    newHeight := int(float64(imgHeight) * scale)
    
    // Center the scaled image
    offsetX := (targetSize - newWidth) / 2
    offsetY := (targetSize - newHeight) / 2
    
    // Draw the scaled image
    dc.DrawImageAnchored(img, offsetX+newWidth/2, offsetY+newHeight/2, 0.5, 0.5)
    
    return dc.Image()
}

// Bot handler function for creating collage - updated for JPEG
func CreateCollageFromResultsNine1(bot *telebot.Bot) func(c telebot.Context) error {
    return func(c telebot.Context) error {
        chat := c.Chat()
        
        // Get all images from results folder
        imagePaths, err := filepath.Glob("results/*.jpg")
        if err != nil {
            return c.Send("❌ Помилка читання папки з результатами")
        }
        
        if len(imagePaths) == 0 {
            return c.Send("❌ Не знайдено зображень для створення колажу")
        }
        
        // Limit to 6 images
        if len(imagePaths) > 6 {
            imagePaths = imagePaths[:6]
        }
        
        c.Send(fmt.Sprintf("🎨 Створюю колаж з %d зображень...", len(imagePaths)))
        
        // Create collage
        err = CreateCollageWithGGNine1()
        if err != nil {
            return c.Send(fmt.Sprintf("❌ Помилка створення колажу: %v", err))
        }
        
        // Check if collage file was created
        if _, err := os.Stat("collage.jpg"); os.IsNotExist(err) {
            return c.Send("❌ Файл колажу не був створений")
        }
        
        // Send as photo (JPEG is better supported by Telegram)
        photo := &telebot.Photo{
            File:    telebot.FromDisk("collage.jpg"),
            Caption: "🎨 Ваш колаж готовий!\n\n📱 Ідеально підходить для обоїв телефону",
        }
        
        _, err = bot.Send(chat, photo)
        if err != nil {
            return c.Send(fmt.Sprintf("❌ Помилка відправки колажу: %v", err))
        }
        
        // Also send as document for highest quality
        document := &telebot.Document{
            File:     telebot.FromDisk("collage.jpg"),
            MIME:     "image/jpeg",
            FileName: "wallpaper_collage.jpg",
            Caption:  "📱 Високоякісна версія для завантаження як обої",
        }
        
        bot.Send(chat, document)
        
        // Clean up - remove the collage file after sending
        go func() {
            // Wait a bit to ensure file was sent, then clean up
            time.Sleep(5 * time.Second)
            os.Remove("collage.jpg")
            fmt.Println("Temporary collage file cleaned up")
        }()
        
        return nil
    }
}

// Optional: Function to create a quick preview collage (smaller size for testing)
func CreatePreviewCollageNine() error {
    imagePaths, err := filepath.Glob("results/*.jpg")
    if err != nil {
        return fmt.Errorf("error reading results folder: %w", err)
    }
    
    if len(imagePaths) == 0 {
        return fmt.Errorf("no images found in results folder")
    }
    
    // Limit to 6 images
    if len(imagePaths) > 6 {
        imagePaths = imagePaths[:6]
    }
    
    // Create smaller canvas for preview
    dc := gg.NewContext(540, 960) // Half size of main collage
    
    // Simple white background for preview
    dc.SetRGB(1, 1, 1)
    dc.Clear()
    
    // Calculate smaller layout
    imageSize := 160
    spacing := 20
    marginTop := 50
    marginSide := (540 - (2*imageSize + spacing)) / 2
    
    // Positions for 2x3 grid
    positions := [][2]int{
        {marginSide, marginTop},
        {marginSide + imageSize + spacing, marginTop},
        {marginSide, marginTop + imageSize + spacing},
        {marginSide + imageSize + spacing, marginTop + imageSize + spacing},
        {marginSide, marginTop + 2*(imageSize + spacing)},
        {marginSide + imageSize + spacing, marginTop + 2*(imageSize + spacing)},
    }
    
    for i, path := range imagePaths {
        if i >= 6 { break }
        
        img, err := gg.LoadImage(path)
        if err != nil {
            continue
        }
        
        processedImg := processImageForSquare(img, imageSize)
        
        x := positions[i][0]
        y := positions[i][1]
        
        dc.DrawImageAnchored(processedImg, x+imageSize/2, y+imageSize/2, 0.5, 0.5)
    }
    
    // Save preview as JPEG
    img := dc.Image()
    
    file, err := os.Create("collage_preview.jpg")
    if err != nil {
        return fmt.Errorf("failed to create preview file: %w", err)
    }
    defer file.Close()
    
    err = jpeg.Encode(file, img, &jpeg.Options{Quality: 85})
    if err != nil {
        return fmt.Errorf("failed to encode preview JPEG: %w", err)
    }
    
    fmt.Println("Preview collage saved as collage_preview.jpg")
    return nil
}
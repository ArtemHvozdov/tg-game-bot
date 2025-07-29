package btnmanager

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/telebot.v3"
)

// Get creates and returns an inline button by unique ID
func (m *Manager) Get(markup *telebot.ReplyMarkup, unique string, args ...any) telebot.Btn {
	config, exists := m.buttons[unique]
	if !exists {
		// Returning the error button via the standard API
		fmt.Printf("DEBUG: Button '%s' not found in manager\n", unique)
		return markup.Data(fmt.Sprintf("ERROR: Button '%s' not found", unique), "error")
	}

	//fmt.Printf("DEBUG: Found button '%s' with data template: '%s'\n", unique, config.Data)
	//fmt.Printf("DEBUG: Args passed: %v\n", args)

	// If a URL is specified, create a link button
	if config.URL != "" {
		//fmt.Printf("DEBUG: Creating URL button\n")
		return markup.URL(config.Text, config.URL)
	} else {
		// Otherwise, create a callback button
		var callbackData string
		if config.Data != "" {
			// Format callback_data with passed arguments
			callbackData = fmt.Sprintf(config.Data, args...)
			//fmt.Printf("DEBUG: Formatted callback_data: '%s'\n", callbackData)
		} else {
			// If data is not specified, use unique as callback_data
			callbackData = unique
			//fmt.Printf("DEBUG: Using unique as callback_data: '%s'\n", callbackData)
		}
		return markup.Data(config.Text, callbackData)
	}
}

// GetInlineButton creates a regular InlineButton structure (for cases where a structure is needed)
func (m *Manager) GetInlineButton(unique string, args ...any) telebot.InlineButton {
	config, exists := m.buttons[unique]
	if !exists {
		return telebot.InlineButton{
			Text: fmt.Sprintf("ERROR: Button '%s' not found", unique),
			Data: "error",
		}
	}

	button := telebot.InlineButton{
		Text: config.Text,
	}

	// If a URL is specified, create a link button
	if config.URL != "" {
		button.URL = config.URL
	} else {
		// Otherwise, create a callback button
		if config.Data != "" {
			// Format callback_data with passed arguments
			button.Data = fmt.Sprintf(config.Data, args...)
		} else {
			// If data is not specified, use unique as callback_data
			button.Data = unique
		}
	}

	return button
}

// GetConfig returns the configuration of a button by unique identifier
func (m *Manager) GetConfig(unique string) (ButtonConfig, bool) {
	config, exists := m.buttons[unique]
	return config, exists
}

// HasButton checks if a button with the specified id exists
func (m *Manager) HasButton(unique string) bool {
	_, exists := m.buttons[unique]
	return exists
}

// GetAllButtons returns all available buttons
func (m *Manager) GetAllButtons() map[string]ButtonConfig {
	result := make(map[string]ButtonConfig)
	for k, v := range m.buttons {
		result[k] = v
	}
	return result
}

// loadFromFile loads button configuration from JSON file
func (m *Manager) loadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read buttons config file: %w", err)
	}

	var configs []ButtonConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("failed to parse buttons config JSON: %w", err)
	}

	// Clearing existing buttons
	m.buttons = make(map[string]ButtonConfig)

	// Загружаем новые кнопки
	for _, config := range configs {
		if config.Unique == "" {
			continue // Skipping buttons without a unique identifier
		}
		m.buttons[config.Unique] = config
	}

	return nil
}
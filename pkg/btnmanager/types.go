package btnmanager

// ButtonConfig represents the button configuration from JSON
type ButtonConfig struct {
	Text   string `json:"text"`   // Button text
	Unique string `json:"unique"` // Unique identifier
	Data   string `json:"data"`   // Callback_data template
	URL    string `json:"url"`    // URL for button links (optional)
}

// Manager manages buttons loaded from configuration
type Manager struct {
	buttons map[string]ButtonConfig // unique -> ButtonConfig
}

// NewManager creates a new instance of the button manager
func NewManager() *Manager {
	return &Manager{
		buttons: make(map[string]ButtonConfig),
	}
}
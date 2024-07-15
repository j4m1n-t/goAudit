package internal

import (
	//Standard Library Imports//
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	// Fyne Imports//
	"fyne.io/fyne/v2"
)

type AppConfig struct {
	IconPath   string `json:"iconPath"`
	ConfigPath string `json:"config"`
}

var configPath string

func SetLog() {
	AppDataDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Printf("Failed to get user AppData dir: %v\n", err)
		return
	}
	logDir := filepath.Join(AppDataDir, "GoAD")
	err = os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
		return
	}
	logFile, err := os.OpenFile(filepath.Join(logDir, "GoAD.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Logging initialized")
}

func SaveConfig(config AppConfig) error {
	err := os.MkdirAll(filepath.Dir(configPath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(&config)
	if err != nil {
		return fmt.Errorf("failed to encode config file: %w", err)
	}

	return nil
}

func LoadConfig() AppConfig {
	config := AppConfig{}
	configDir, err := os.UserConfigDir()
	if err != nil {
		fyne.LogError("Failed to get user config dir.", err)
	}
	configPath = filepath.Join(configDir, "GoAD", "config.json")
	iconPath := filepath.Join(configDir, "GoAD", "goad2.ico")
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create the directory if it doesn't exist
			err = os.MkdirAll(filepath.Dir(configPath), os.ModePerm)
			if err != nil {
				log.Printf("Failed to create config directory: %v", err)
			}
			// Return default config as the file doesn't exist yet
			return config
		}
		log.Printf("Failed to open config file: %v", err)
		return config
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Printf("Failed to decode config file: %v", err)
		// Return default config if decoding fails
		return AppConfig{ConfigPath: configPath, IconPath: iconPath}
	}
	return config
}

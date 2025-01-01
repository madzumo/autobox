package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var (
	headerMenu = `

+========================================================+
|    _              _               ____     ___   __  __|
|   / \     _   _  | |_    ___     | __ )   / _ \  \ \/ /|
|  / _ \   | | | | | __|  / _ \    |  _ \  | | | |  \  / |
| / ___ \  | |_| | | |_  | (_) |   | |_) | | |_| |  /  \ |
|/_/   \_\  \__,_|  \__|  \___/    |____/   \___/  /_/\_\|
+========================================================+
												by madzumo
`

	settingsFileName = "settings.json"
)

type applicationMain struct {
	droplet  *Droplets
	settings *settingsConfig
	manifest string
}

type settingsConfig struct {
	DoAPI       string `json:"doAPI"`
	NumberBoxes int    `json:"boxes"`
	AwsKey      string `json:"awsKey"`
	AwsSecret   string `json:"awsSecret"`
}

func main() {

	settingsX, _ := getSettings()

	app := &applicationMain{
		droplet:  &Droplets{},
		settings: settingsX,
		manifest: fmt.Sprintf("\nManifest:\nDigital Ocean API:%s\nBoxes:%d\n", settingsX.DoAPI, settingsX.NumberBoxes),
	}

	ShowMenu(app)
}

func getSettings() (*settingsConfig, error) {

	configTemp := settingsConfig{
		DoAPI:       "APIkey",
		NumberBoxes: 1,
		AwsKey:      "awsKEY",
		AwsSecret:   "awsSECRET",
	}

	data, err := os.ReadFile(settingsFileName)
	if err != nil {
		return &configTemp, err
	}

	err = json.Unmarshal(data, &configTemp)
	return &configTemp, err
}

func saveSettings(config *settingsConfig) error {
	//convert to struct -> JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsFileName, data, 0644)
}

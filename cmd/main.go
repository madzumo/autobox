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
	settings *settingsConfig
	header   string
}

type settingsConfig struct {
	DoAPI       string `json:"doAPI"`
	NumberBoxes int    `json:"boxes"`
	AwsKey      string `json:"awsKey"`
	AwsSecret   string `json:"awsSecret"`
	Provider    string `json:"provider"` //digital, aws or linode
	URL         string `json:"url"`
	LinodeAPI   string `json:"linodeAPI"`
}

func main() {

	settingsX, err := getSettings()
	if err != nil {
		fmt.Printf("Error retrieving settings: %s", err)
	}
	app := &applicationMain{
		settings: settingsX,
	}
	app.updateHeader()
	ShowMenu(app)
}

func (app *applicationMain) updateHeader() {
	manifest := fmt.Sprintf("\nProvider: %s\nAPI: %.15s...\nBoxes: %d\nURL: %s", app.settings.Provider, app.settings.DoAPI, app.settings.NumberBoxes, app.settings.URL)
	app.header = lipHeaderStyle.Render(headerMenu) + lipManifestStyle.Render(manifest)
}

func getSettings() (settings *settingsConfig, err error) {

	settingsTemp := settingsConfig{
		DoAPI:       "APIkey",
		NumberBoxes: 1,
		AwsKey:      "awsKEY",
		AwsSecret:   "awsSECRET",
		Provider:    "digital",
		LinodeAPI:   "APIkey",
	}

	data, err := os.ReadFile(settingsFileName)
	if err != nil {
		return &settingsTemp, err
	}

	err = json.Unmarshal(data, &settingsTemp)
	if err != nil {
		return &settingsTemp, err
	}

	return &settingsTemp, nil
}

func saveSettings(config *settingsConfig) error {
	//convert to struct -> JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsFileName, data, 0644)
}

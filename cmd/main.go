package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
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
	Aws         *AWS     `json:"aws"`
	Digital     *Digital `json:"digital"`
	NumberBoxes int      `json:"boxes"`
	Provider    string   `json:"provider"`
	URL         string   `json:"url"`
	BatchTag    string   `json:"batchtag"`
}

func main() {

	app, err := getSettings()
	if err != nil {
		fmt.Printf("Error retrieving settings: %s", err)
	}

	ShowMenu(app)
}

func (app *applicationMain) getAppHeader() string {
	var manifest string
	pepa, _ := app.Aws.createEc2Client()
	ec2s, _ := app.Aws.getActiveEC2s(pepa)
	switch app.Provider {
	case "digital":
		manifest = fmt.Sprintf("\nProvider: %s\nRegion: %s\nAPI: %.10s...\nURL: %s\nDeploy Boxes: %d  Running: 0", app.Provider, app.Digital.Region, app.Digital.ApiToken, app.URL, app.NumberBoxes)
	case "aws":
		manifest = fmt.Sprintf("\nProvider: %s\nRegion: %s\nKey/Secret: %.8s.../%.8s...\nURL: %s\nDeploy Boxes: %d  Running: %d", app.Provider, app.Aws.Region, app.Aws.Key, app.Aws.Secret, app.URL, app.NumberBoxes, ec2s)
		// case "linode":
		// 	manifest = fmt.Sprintf("\nProvider: %s\nAPI: %.15s...\nBoxes: %d\nURL: %s", app.settings.Provider, app.settings.LinodeAPI, app.settings.NumberBoxes, app.settings.URL)
	}

	batchTagColorize := fmt.Sprintf("\nTag: %s", app.BatchTag)

	if app.Provider == "aws" {
		manifestColorFront = awsColorFront
	} else if app.Provider == "digital" {
		manifestColorFront = digitalColorFront
	}
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(headerColorFront)).Render(headerMenu) +
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(manifestColorFront)).Render(manifest) +
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(headerColorFront)).Render(batchTagColorize)
}

func getSettings() (appMain *applicationMain, err error) {
	app := &applicationMain{
		Aws:         &AWS{PemKeyFileName: "autobox", AmiID: "ami-036841078a4b68e14", InstanceType: "t3a.small"},
		Digital:     &Digital{InstanceSize: "s-1vcpu-2gb", ImageSlug: "ubuntu-24-10-x64", Tags: []string{"AUTO-BOX"}},
		NumberBoxes: 1,
		Provider:    "aws",
	}

	data, err := os.ReadFile(settingsFileName)
	if err != nil {
		return app, err
	}

	err = json.Unmarshal(data, &app)
	if err != nil {
		return nil, err
	}

	return app, nil
}

func saveSettings(appMain *applicationMain) error {
	//convert to struct -> JSON
	data, err := json.MarshalIndent(appMain, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsFileName, data, 0644)
}

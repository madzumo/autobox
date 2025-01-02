package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func (app *applicationMain) runVNCcommand() {
	// Define the absolute path to the executable
	executable := `C:\Program Files\TightVNC\tvnviewer.exe`

	// Define the arguments for the executable
	args := []string{
		"167.71.180.205::5901",
		"-password=prime6996",
	}

	// Create the command
	cmd := exec.Command(executable, args...)

	// Optional: Redirect output to the console
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing command: %v\n", err)
	} else {
		fmt.Println("Command executed successfully.")
	}
}

func (app *applicationMain) runPS1files() {
	// Folder containing the PowerShell scripts
	scriptsFolder := "./boxes" // Adjust the path as needed

	// Get all .ps1 files in the folder
	files, err := os.ReadDir(scriptsFolder)
	if err != nil {
		fmt.Println("Error reading scripts folder:", err)
		return
	}

	// Loop through each .ps1 file and execute it
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".ps1" {
			scriptPath := filepath.Join(scriptsFolder, file.Name())

			// Command to run the script in a new PowerShell window
			cmd := exec.Command("powershell", "-NoExit", "-File", scriptPath)

			// Start the PowerShell script in a separate window
			err := cmd.Start()
			if err != nil {
				fmt.Printf("Error starting script %s: %v\n", file.Name(), err)
				continue
			}

			fmt.Printf("Started script: %s\n", file.Name())
		}
	}
}

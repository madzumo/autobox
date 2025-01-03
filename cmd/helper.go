package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func (app *applicationMain) runVNC(ipDestination string) error {
	// Define the absolute path to the executable
	executable := `C:\Program Files\TightVNC\tvnviewer.exe`

	// Define the arguments for the executable
	args := []string{
		fmt.Sprintf("%s::5901", ipDestination),
		"-password=prime7",
	}

	// Create the command
	cmd := exec.Command(executable, args...)

	// Optional: Redirect output to the console
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err := cmd.Run()
	return err
	// if err := cmd.Run(); err != nil {
	// 	fmt.Printf("Error executing command: %v\n", err)
	// } else {
	// 	fmt.Println("Command executed successfully.")
	// }
}

func (app *applicationMain) runPS1files() error {
	// Folder containing the PowerShell scripts
	scriptsFolder := "./boxes" // Adjust the path as needed

	// Get all .ps1 files in the folder
	files, err := os.ReadDir(scriptsFolder)
	if err != nil {
		return err
	}

	// Loop through each .ps1 file and execute it
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".ps1" {
			scriptPath := filepath.Join(scriptsFolder, file.Name())

			// cmd := exec.Command("powershell", "-NoExit", "-File", scriptPath)
			// err := cmd.Start()

			psCommand := fmt.Sprintf(
				"Start-Process powershell.exe -WindowStyle Normal -ArgumentList '-NoExit','-File','%s'",
				scriptPath,
			)

			cmd := exec.Command(
				"powershell.exe",
				"-NoProfile",
				"-ExecutionPolicy", "Bypass",
				"-Command", psCommand,
			)

			if err := cmd.Start(); err != nil {
				return err
			}
			// fmt.Printf("Started script: %s\n", file.Name())
		}
	}
	return nil
}

func countNumberofFiles(folderPath string) (int, error) {
	// Check if the folder exists
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return 0, nil
	}

	// Open the folder
	dir, err := os.Open(folderPath)
	if err != nil {
		return 0, err
	}
	defer dir.Close()

	// Read the folder's contents
	files, err := dir.Readdir(0)
	if err != nil {
		return 0, err
	}

	return len(files), nil
}

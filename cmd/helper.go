package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func (app *applicationMain) createPostSCRIPT(ipAddress string, awsKeyName string) error {
	// scriptsFolder := fmt.Sprintf("./%s", app.settings.RegionDigital)
	scriptsFolder := app.Digital.Region
	commands := fmt.Sprintf(`ssh -o StrictHostKeyChecking=no root@%s "export URL='%s' && curl -sSL https://raw.githubusercontent.com/madzumo/autobox/main/scripts/startup.sh | bash"`,
		ipAddress, app.URL)

	if app.Provider == "aws" {
		scriptsFolder = app.Aws.Region
		scriptPath, _ := filepath.Abs(filepath.Join(fmt.Sprintf("./%s", scriptsFolder), awsKeyName))
		commands = fmt.Sprintf(`ssh -i "%s.pem" -o StrictHostKeyChecking=no ubuntu@%s "export URL='%s' && curl -sSL https://raw.githubusercontent.com/madzumo/autobox/main/scripts/startup.sh | bash"`,
			scriptPath, ipAddress, app.URL)
	}

	if app.checkFileNameExist(ipAddress, fmt.Sprintf("./%s", scriptsFolder)) {
		return nil
	}
	filename := fmt.Sprintf("%s-%s.ps1", app.BatchTag, ipAddress)

	// Ensure the directory exists
	err2 := os.MkdirAll(scriptsFolder, 0755)
	if err2 != nil {
		return err2
	}

	fullPath := fmt.Sprintf("%s/%s", scriptsFolder, filename)

	// Create or overwrite the .ps1 file in the current directory
	err := os.WriteFile(fullPath, []byte(commands), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (app *applicationMain) checkFileNameExist(fileString2Check, directory2Check string) bool {
	files, err := os.ReadDir(directory2Check)
	if err != nil {
		return true
	}
	found := false
	for _, file := range files {
		if strings.Contains(file.Name(), fileString2Check) {
			found = true
			break
		}
	}
	return found
}

// func (app *applicationMain) runPS1files2() error {
// 	scriptsFolder := fmt.Sprintf("./%s", app.settings.RegionDigital)
// 	if app.settings.Provider == "aws" {
// 		scriptsFolder = fmt.Sprintf("./%s", app.settings.RegionAws)
// 	}

// 	// Get all .ps1 files in the folder
// 	files, err := os.ReadDir(scriptsFolder)
// 	if err != nil {
// 		return err
// 	}

// 	// Loop through each .ps1 file and execute it
// 	for _, file := range files {
// 		if filepath.Ext(file.Name()) == ".ps1" {
// 			scriptPath := filepath.Join(scriptsFolder, file.Name())

// 			psCommand := fmt.Sprintf(
// 				"Start-Process powershell.exe -WindowStyle Normal -ArgumentList '-NoExit','-File','%s'",
// 				scriptPath,
// 			)

// 			cmd := exec.Command(
// 				"powershell.exe",
// 				"-NoProfile",
// 				"-ExecutionPolicy", "Bypass",
// 				"-Command", psCommand,
// 			)

// 			if err := cmd.Start(); err != nil {
// 				return err
// 			}
// 			// fmt.Printf("Started script: %s\n", file.Name())
// 		}
// 	}
// 	return nil
// }

// func (app *applicationMain) runPS1files() error {
// 	scriptsFolder := fmt.Sprintf("./%s", app.Digital.Region)
// 	if app.Provider == "aws" {
// 		scriptsFolder = fmt.Sprintf("./%s", app.Aws.Region)
// 	}

// 	files, err := os.ReadDir(scriptsFolder)
// 	if err != nil {
// 		return err
// 	}

// 	// Loop through each .ps1 file and execute it
// 	batchMatch := true
// 	for _, file := range files {
// 		if app.BatchTag != "" {
// 			batchMatch = strings.Contains(file.Name(), app.BatchTag)
// 		}
// 		if filepath.Ext(file.Name()) == ".ps1" && batchMatch {
// 			// Generate an absolute path for the script
// 			scriptPath, err := filepath.Abs(filepath.Join(scriptsFolder, file.Name()))
// 			if err != nil {
// 				return fmt.Errorf("failed to get absolute path for %s: %w", file.Name(), err)
// 			}

// 			// psCommand := fmt.Sprintf(
// 			// 	"Start-Process powershell.exe -WindowStyle Normal -ArgumentList '-NoExit','-File','%s'",
// 			// 	scriptPath,
// 			// )
// 			// cmd := exec.Command(
// 			// 	"powershell.exe",
// 			// 	"-NoProfile",
// 			// 	"-ExecutionPolicy", "Bypass",
// 			// 	"-Command", psCommand,
// 			// )

// 			// if err := cmd.Start(); err != nil {
// 			// 	return fmt.Errorf("failed to start script %s: %w", scriptPath, err)
// 			// }

// 			psCommand := fmt.Sprintf(
// 				"Start-Process -FilePath powershell.exe -ArgumentList '-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', '%s' -Wait -WindowStyle Normal",
// 				scriptPath,
// 			)
// 			cmd := exec.Command(
// 				"powershell.exe",
// 				"-NoProfile",
// 				"-ExecutionPolicy", "Bypass",
// 				"-Command", psCommand,
// 			)

// 			// Run the command
// 			err = cmd.Run()
// 			if err != nil {
// 				fmt.Printf("Error: %v\n", err)
// 			} else {
// 				fmt.Println("Post Launch script finished..")
// 			}
// 		}
// 	}
// 	return nil
// }

func (app *applicationMain) runPS1file(scriptPath, fileName string) error {

	// psCommand := fmt.Sprintf(
	// 	"Start-Process powershell.exe -WindowStyle Normal -ArgumentList '-NoExit','-File','%s'",
	// 	scriptPath,
	// )
	// cmd := exec.Command(
	// 	"powershell.exe",
	// 	"-NoProfile",
	// 	"-ExecutionPolicy", "Bypass",
	// 	"-Command", psCommand,
	// )

	// if err := cmd.Start(); err != nil {
	// 	return fmt.Errorf("failed to start script %s: %w", scriptPath, err)
	// }

	psCommand := fmt.Sprintf(
		"Start-Process -FilePath powershell.exe -ArgumentList '-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', '%s' -Wait -WindowStyle Normal",
		scriptPath,
	)
	cmd := exec.Command(
		"powershell.exe",
		"-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-Command", psCommand,
	)

	// Run the command
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error with %s:\n%s", fileName, err)
	} else {
		fmt.Printf("\n%s script finished", fileName)
	}

	return nil
}

func countNumberofFiles(folderPath string) (int, error) {
	// Check if the folder exists
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		return 0, nil
	}

	//read files in directory
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return 0, nil
	}

	return len(files), nil
}

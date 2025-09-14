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
	fmt.Printf("DEBUG createPostSCRIPT: Starting for IP %s, awsKeyName %s\n", ipAddress, awsKeyName)

	// Get current working directory
	currentDir, _ := os.Getwd()
	fmt.Printf("DEBUG createPostSCRIPT: Current working directory: %s\n", currentDir)

	// Create commands based on provider - SIMPLE APPROACH
	var commands string
	if app.Provider == "aws" {
		// PEM key is now in current directory, so reference it directly
		pemKeyPath := filepath.Join(currentDir, awsKeyName)
		fmt.Printf("DEBUG createPostSCRIPT: Using PEM key path: %s.pem\n", pemKeyPath)

		// Create simple SSH command that uses the EXISTING startup.sh script
		commands = fmt.Sprintf(`ssh -i "%s.pem" -o StrictHostKeyChecking=no ubuntu@%s "export URL='%s' && curl -sSL https://raw.githubusercontent.com/madzumo/autobox/main/scripts/startup.sh | bash"`,
			pemKeyPath, ipAddress, app.URL)
	} else {
		// Digital Ocean version
		commands = fmt.Sprintf(`ssh -o StrictHostKeyChecking=no root@%s "export URL='%s' && curl -sSL https://raw.githubusercontent.com/madzumo/autobox/main/scripts/startup.sh | bash"`,
			ipAddress, app.URL)
	}

	fmt.Printf("DEBUG createPostSCRIPT: Provider=%s, BatchTag='%s'\n", app.Provider, app.BatchTag)

	// Check if file already exists in current directory
	fileExists := app.checkFileNameExist(ipAddress, currentDir)
	fmt.Printf("DEBUG createPostSCRIPT: File exists check result: %v\n", fileExists)

	if fileExists {
		fmt.Printf("DEBUG createPostSCRIPT: File already exists for IP %s, skipping\n", ipAddress)
		return nil
	}

	filename := fmt.Sprintf("%s-%s.ps1", app.BatchTag, ipAddress)
	fmt.Printf("DEBUG createPostSCRIPT: Creating filename: %s\n", filename)

	fullPath := filepath.Join(currentDir, filename)
	fmt.Printf("DEBUG createPostSCRIPT: Full file path: %s\n", fullPath)
	fmt.Printf("DEBUG createPostSCRIPT: Command content length: %d bytes\n", len(commands))

	// Create or overwrite the .ps1 file in the current directory
	fmt.Printf("DEBUG createPostSCRIPT: About to write file...\n")
	err := os.WriteFile(fullPath, []byte(commands), 0644)
	if err != nil {
		fmt.Printf("DEBUG createPostSCRIPT: Error writing file: %v\n", err)
		fmt.Printf("DEBUG createPostSCRIPT: Error type: %T\n", err)
		return err
	}
	fmt.Printf("DEBUG createPostSCRIPT: File write completed\n")

	// Verify file was created
	if stat, err := os.Stat(fullPath); err != nil {
		fmt.Printf("DEBUG createPostSCRIPT: Failed to verify file creation: %v\n", err)
		return fmt.Errorf("file creation verification failed: %w", err)
	} else {
		fmt.Printf("DEBUG createPostSCRIPT: File verified - size: %d bytes, mode: %v\n", stat.Size(), stat.Mode())
	}

	fmt.Printf("DEBUG createPostSCRIPT: Successfully created file: %s\n", fullPath)
	return nil
}

func (app *applicationMain) checkFileNameExist(fileString2Check, directory2Check string) bool {
	fmt.Printf("DEBUG checkFileNameExist: Looking for '%s' in directory '%s'\n", fileString2Check, directory2Check)

	files, err := os.ReadDir(directory2Check)
	if err != nil {
		fmt.Printf("DEBUG checkFileNameExist: Error reading directory '%s': %v - returning true (skip creation)\n", directory2Check, err)
		return true
	}

	fmt.Printf("DEBUG checkFileNameExist: Found %d files in directory\n", len(files))
	found := false
	for _, file := range files {
		fmt.Printf("DEBUG checkFileNameExist: Checking file '%s' - contains '%s': %v\n", file.Name(), fileString2Check, strings.Contains(file.Name(), fileString2Check))
		if strings.Contains(file.Name(), fileString2Check) {
			found = true
			fmt.Printf("DEBUG checkFileNameExist: Match found! File '%s' contains '%s'\n", file.Name(), fileString2Check)
			break
		}
	}
	fmt.Printf("DEBUG checkFileNameExist: Final result: %v\n", found)
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
	fmt.Printf("DEBUG runPS1file: Executing script: %s\n", scriptPath)
	fmt.Printf("DEBUG runPS1file: Script file name: %s\n", fileName)

	// Launch PowerShell script in a NEW WINDOW and DO NOT WAIT for completion
	fmt.Printf("DEBUG runPS1file: Starting PowerShell script in new window (non-blocking)...\n")

	// Use Start-Process to launch in a new window without waiting
	psCommand := fmt.Sprintf(
		"Start-Process -FilePath 'powershell.exe' -ArgumentList '-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', '%s' -WindowStyle Normal",
		scriptPath, // Removed -Wait so it doesn't block
	)

	cmd := exec.Command(
		"powershell.exe",
		"-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-Command", psCommand,
	)

	// Set working directory
	cmd.Dir = filepath.Dir(scriptPath)
	fmt.Printf("DEBUG runPS1file: Working directory: %s\n", cmd.Dir)

	// Start the process and DO NOT WAIT for it to complete
	err := cmd.Start() // Use Start() instead of Run() or CombinedOutput()
	if err != nil {
		fmt.Printf("DEBUG runPS1file: Failed to start PowerShell window: %v\n", err)
		return fmt.Errorf("failed to start PowerShell script %s: %w", fileName, err)
	}

	fmt.Printf("DEBUG runPS1file: Successfully launched %s in new PowerShell window (PID: %d)\n", fileName, cmd.Process.Pid)
	fmt.Printf("DEBUG runPS1file: Script is running in background, not waiting for completion\n")

	return nil
}

// func countNumberofFiles(folderPath string) (int, error) {
// 	// Check if the folder exists
// 	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
// 		return 0, nil
// 	}

// 	//read files in directory
// 	files, err := os.ReadDir(folderPath)
// 	if err != nil {
// 		return 0, nil
// 	}

// 	return len(files), nil
// }

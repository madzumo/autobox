package main

import (
	"fmt"
	"os"
	"os/exec"
)

func runVNCcommand() {
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

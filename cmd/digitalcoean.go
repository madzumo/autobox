package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/digitalocean/godo"
)

var (
	startupScript = `
	export URL="https://twitch.tv/vortix93"
	curl -sSL https://raw.githubusercontent.com/madzumo/autobox/main/scripts/startup.sh | bash
	`
)

type Droplets struct {
	region string
	size   string
	image  string
}

func (app *applicationMain) createBox(token string) int {
	client := godo.NewFromToken(token)
	ctx := context.TODO()

	dropletName := fmt.Sprintf("autobox-%d", rand.Int())
	region := "nyc3"
	size := "s-1vcpu-2gb" //.018/hour
	// size := "s-1vcpu-1gb" // .009/hour
	imageSlug := "ubuntu-24-10-x64"
	// Retrieve all SSH keys on the account
	sshKeys, _, err := client.Keys.List(ctx, &godo.ListOptions{})
	if err != nil {
		log.Fatalf("Error retrieving SSH keys: %v", err)
	}

	// Convert to the required type for DropletCreateRequest
	var dropletSSHKeys []godo.DropletCreateSSHKey
	for _, key := range sshKeys {
		dropletSSHKeys = append(dropletSSHKeys, godo.DropletCreateSSHKey{
			ID: key.ID,
		})
	}

	createRequest := &godo.DropletCreateRequest{
		Name:   dropletName,
		Region: region,
		Size:   size,
		Image: godo.DropletCreateImage{
			Slug: imageSlug,
		},
		SSHKeys:  dropletSSHKeys,
		Backups:  false,
		Tags:     []string{"AUTO-BOX"},
		UserData: startupScript,
	}

	droplet, _, err := client.Droplets.Create(ctx, createRequest)
	if err != nil {
		fmt.Printf("Error creating droplet: %v", err)
		return 0
	}

	fmt.Printf("Droplet created! ID:%d, Name: %s\n", droplet.ID, droplet.Name)
	return droplet.ID
}

func (app *applicationMain) deleteBox(token string, dropletID int) {
	client := godo.NewFromToken(token)
	ctx := context.TODO()

	_, err := client.Droplets.Delete(ctx, dropletID)
	if err != nil {
		log.Fatalf("Error deleting droplet: %v", err)
	}

	fmt.Printf("Droplet with ID %d deleted successfully!\n", dropletID)
}

func (app *applicationMain) createFirewall(token string) error {
	// Define the firewall rule
	firewallRequest := &godo.FirewallRequest{
		Name: "autoBOX-Firewall",
		InboundRules: []godo.InboundRule{
			{
				Protocol:  "tcp",
				PortRange: "5901",
				Sources:   &godo.Sources{Addresses: []string{"0.0.0.0/0", "::/0"}},
			},
			{
				Protocol:  "tcp",
				PortRange: "80",
				Sources:   &godo.Sources{Addresses: []string{"0.0.0.0/0", "::/0"}},
			},
			{
				Protocol:  "tcp",
				PortRange: "443",
				Sources:   &godo.Sources{Addresses: []string{"0.0.0.0/0", "::/0"}},
			},
			{
				Protocol:  "tcp",
				PortRange: "22",
				Sources:   &godo.Sources{Addresses: []string{"0.0.0.0/0", "::/0"}},
			},
		},
		OutboundRules: []godo.OutboundRule{
			{
				Protocol:     "tcp",
				PortRange:    "1-65535", // Allow all TCP ports
				Destinations: &godo.Destinations{Addresses: []string{"0.0.0.0/0", "::/0"}},
			},
			{
				Protocol:     "udp",
				PortRange:    "1-65535", // Allow all UDP ports
				Destinations: &godo.Destinations{Addresses: []string{"0.0.0.0/0", "::/0"}},
			},
			{
				Protocol:     "icmp",
				Destinations: &godo.Destinations{Addresses: []string{"0.0.0.0/0", "::/0"}}, // Allow all ICMP traffic
			},
		},

		Tags: []string{"AUTO-BOX"},
	}

	client := godo.NewFromToken(token)
	ctx := context.Background()

	//1. check if the firewall exists
	firewalls, _, err := client.Firewalls.List(ctx, &godo.ListOptions{})
	if err != nil {
		return err
	}

	//look for an existing firewall with same name
	for _, fw := range firewalls {
		if fw.Name == "autoBOX-firewall" {
			fmt.Printf("Firewall already exists with ID: %s\n", fw.ID)
			return nil
		}
	}

	//2. create the firewall
	firewall, _, err := client.Firewalls.Create(ctx, firewallRequest)
	if err != nil {
		return err
	}
	fmt.Printf("Firewall created: %v\n", firewall)
	return nil
}

func (app *applicationMain) deleteFirewall(token string) error {
	client := godo.NewFromToken(token)
	ctx := context.Background()

	//list all firewalls
	firewalls, _, err := client.Firewalls.List(ctx, &godo.ListOptions{})
	if err != nil {
		return err
	}

	//loop through them & delete the one we want
	for _, fw := range firewalls {
		if fw.Name == "autoBOX-Firewall" {
			_, err := client.Firewalls.Delete(ctx, fw.ID)
			if err != nil {
				fmt.Printf("Error deleting firewall (ID:%s): %v\n", fw.ID, err)
			} else {
				fmt.Printf("Deleted firewall: %s (ID:%s)\n", fw.Name, fw.ID)
			}
		}
	}
	return nil
}

// func (app *applicationMain) addDropletToFirewall(token string, firewallID string, dropletID int) {
// 	client := godo.NewFromToken(token)
// 	_, err := client.Firewalls.AddDroplets(context.Background(), firewallID, dropletID)
// 	if err != nil {
// 		fmt.Printf("Error adding droplet %d to firewall: %v\n", dropletID, err)
// 		return
// 	}
// 	fmt.Printf("Droplet %d added to firewall successfully.\n", dropletID)
// }

func (app *applicationMain) createSSHkey(token string) int {
	client := godo.NewFromToken(token)
	ctx := context.TODO()

	keyCreateRequest := &godo.KeyCreateRequest{
		Name:      "gokey",
		PublicKey: "publickeyhere",
	}

	sshKey, _, err := client.Keys.Create(ctx, keyCreateRequest)

	if err != nil {
		log.Fatalf("Failed to create SSH key: %v", err)
	}

	fmt.Printf("SSH Key created: %v\n", sshKey.Name)
	return sshKey.ID
}

// func (app *applicationMain) saveFirewall(id string) error {
// 	execpath, _ := os.Executable()
// 	dir := filepath.Dir(execpath)
// 	filepath := filepath.Join(dir, "firewall.txt")

// 	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 06044)

// 	if err != nil {
// 		return fmt.Errorf("failed to open file: %w", err)
// 	}

// 	defer file.Close()

// 	_, err = file.WriteString(id + "\n")
// 	if err != nil {
// 		return fmt.Errorf("failed to write to file: %w", err)
// 	}
// 	return nil
// }

func (app *applicationMain) saveIDsLocal(id string) error {
	execpath, _ := os.Executable()
	dir := filepath.Dir(execpath)
	filepath := filepath.Join(dir, "ids.txt")

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 06044)

	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	defer file.Close()

	_, err = file.WriteString(id + "\n")
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}

func (app *applicationMain) getIDsLocal() ([]int, error) {
	execpath, _ := os.Executable()
	dir := filepath.Dir(execpath)
	path := filepath.Join(dir, "ids.txt")

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var ids []int
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			// Skip empty lines
			continue
		}

		// Try to convert line to integer
		val, err := strconv.Atoi(line)
		if err != nil {
			// If not an integer, ignore this line
			continue
		}

		ids = append(ids, val)
	}

	// Check for any scanning error
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// If no valid integers were found, return nil
	if len(ids) == 0 {
		return nil, nil
	}

	return ids, nil
}

func (app *applicationMain) deleteALLboxes(token string) {
	boxes, _ := app.getIDsLocal()

	for _, numID := range boxes {
		app.deleteBox(token, numID)
	}

}

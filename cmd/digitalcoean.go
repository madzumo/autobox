package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/digitalocean/godo"
)

var (
	region = "nyc3"
	size   = "s-1vcpu-2gb" //.018/hour
	// size      = "s-1vcpu-1gb" // .009/hour
	imageSlug = "ubuntu-24-10-x64"
	tags      = []string{"AUTO-BOX"}
)

func (app *applicationMain) createBox(token string) error {
	client := godo.NewFromToken(token)
	ctx := context.TODO()

	dropletName := fmt.Sprintf("autobox-%d", rand.Int())

	// Retrieve all SSH keys on the account
	sshKeys, _, err := client.Keys.List(ctx, &godo.ListOptions{})
	if err != nil {
		return err
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
		SSHKeys: dropletSSHKeys,
		Backups: false,
		Tags:    tags,
	}

	_, _, err2 := client.Droplets.Create(ctx, createRequest)
	if err2 != nil {
		return err2
	}
	// fmt.Println("Droplet created!")
	return nil
}

func (app *applicationMain) deleteBox(token string) error {
	client := godo.NewFromToken(token)
	ctx := context.TODO()

	tag := "AUTO-BOX"

	_, err := client.Droplets.DeleteByTag(ctx, tag)
	return err
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
			// fmt.Printf("Firewall already exists with ID: %s\n", fw.ID)
			return nil
		}
	}

	//2. create the firewall
	_, _, err2 := client.Firewalls.Create(ctx, firewallRequest)
	if err2 != nil {
		return err2
	}
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

// func (app *applicationMain) saveIDsLocal(id string) error {
// 	execpath, _ := os.Executable()
// 	dir := filepath.Dir(execpath)
// 	filepath := filepath.Join(dir, "ids.txt")

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

// func (app *applicationMain) getIDsLocal() ([]int, error) {
// 	execpath, _ := os.Executable()
// 	dir := filepath.Dir(execpath)
// 	path := filepath.Join(dir, "ids.txt")

// 	file, err := os.Open(path)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to open file: %w", err)
// 	}
// 	defer file.Close()

// 	var ids []int
// 	scanner := bufio.NewScanner(file)

// 	for scanner.Scan() {
// 		line := strings.TrimSpace(scanner.Text())
// 		if line == "" {
// 			// Skip empty lines
// 			continue
// 		}

// 		// Try to convert line to integer
// 		val, err := strconv.Atoi(line)
// 		if err != nil {
// 			// If not an integer, ignore this line
// 			continue
// 		}

// 		ids = append(ids, val)
// 	}

// 	// Check for any scanning error
// 	if err := scanner.Err(); err != nil {
// 		return nil, fmt.Errorf("failed to read file: %w", err)
// 	}

// 	// If no valid integers were found, return nil
// 	if len(ids) == 0 {
// 		return nil, nil
// 	}

// 	return ids, nil
// }

func (app *applicationMain) createPostSCRIPT(dropletIP string, position int) error {
	// Replace the IP with the provided dropletIP
	commands := fmt.Sprintf(`
ssh -o StrictHostKeyChecking=no root@%s "export URL='%s' && curl -sSL https://raw.githubusercontent.com/madzumo/autobox/main/scripts/startup.sh | bash"
`, dropletIP, app.settings.URL)

	// File name for the PowerShell script
	filename := fmt.Sprintf("%d-%s.ps1", position, dropletIP)

	// Ensure the directory exists
	err2 := os.MkdirAll("boxes", 0755)
	if err2 != nil {
		return err2
	}

	//check if this IP has been saved as a script yet
	// files, err := os.ReadDir("./boxes")

	// Full path for the file
	fullPath := fmt.Sprintf("%s/%s", "boxes", filename)

	// Create or overwrite the .ps1 file in the current directory
	err := os.WriteFile(fullPath, []byte(commands), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (app *applicationMain) compileIPaddresses() (ips []string, err error) {
	client := godo.NewFromToken(app.settings.DoAPI)
	ctx := context.TODO()
	tag := "AUTO-BOX"

	// List droplets by tag
	droplets, _, err := client.Droplets.ListByTag(ctx, tag, &godo.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Retrieve public IP addresses of each droplet
	for _, droplet := range droplets {
		for _, network := range droplet.Networks.V4 {
			if network.Type == "public" {
				//fmt.Printf("Droplet: %s, Public IP: %s\n", droplet.Name, network.IPAddress)
				ips = append(ips, network.IPAddress)
			}
		}
	}

	return ips, nil
}

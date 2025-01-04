package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/digitalocean/godo"
)

type Digital struct {
	ApiToken     string   `json:"apiToken"`
	Region       string   `json:"region"`
	InstanceSize string   `json:"instanceSize"`
	ImageSlug    string   `json:"imageslug"`
	Tags         []string `json:"tags"`
}

func (d *Digital) createBox() error {
	client := godo.NewFromToken(d.ApiToken)
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
		Region: d.Region,
		Size:   d.InstanceSize,
		Image: godo.DropletCreateImage{
			Slug: d.ImageSlug,
		},
		SSHKeys: dropletSSHKeys,
		Backups: false,
		Tags:    d.Tags,
	}

	_, _, err2 := client.Droplets.Create(ctx, createRequest)
	if err2 != nil {
		return err2
	}
	// fmt.Println("Droplet created!")
	return nil
}

func (d *Digital) deleteBox() error {
	client := godo.NewFromToken(d.ApiToken)
	ctx := context.TODO()

	tag := "AUTO-BOX"

	_, err := client.Droplets.DeleteByTag(ctx, tag)
	return err
}

func (d *Digital) createFirewall() error {
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

	client := godo.NewFromToken(d.ApiToken)
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

func (d *Digital) deleteFirewall() error {
	client := godo.NewFromToken(d.ApiToken)
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

func (d *Digital) createSSHkey() int {
	client := godo.NewFromToken(d.ApiToken)
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

func (d *Digital) compileIPaddressesDigital() (ips []string, err error) {
	client := godo.NewFromToken(d.ApiToken)
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

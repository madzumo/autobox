package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/digitalocean/godo"
)

var (
	startupScript = `
#!/bin/bash
curl -sSL https://raw.githubusercontent.com/madzumo/autobox/main/scripts/startup.sh | bash
`
)

func createBox(token string) int {
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
		log.Fatalf("Error creating droplet: %v", err)
	}

	fmt.Printf("Droplet created! ID:%d, Name: %s\n", droplet.ID, droplet.Name)

	return droplet.ID
}

func deleteBox(token string, dropletID int) {
	client := godo.NewFromToken(token)
	ctx := context.TODO()

	_, err := client.Droplets.Delete(ctx, dropletID)
	if err != nil {
		log.Fatalf("Error deleting droplet: %v", err)
	}

	fmt.Printf("Droplet with ID %d deleted successfully!\n", dropletID)
}

func createFirewall(token string) string {
	// Define the firewall rule
	firewallRequest := &godo.FirewallRequest{
		Name: "pepita-Firewall",
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
	}

	// Create the firewall
	client := godo.NewFromToken(token)
	firewall, _, err := client.Firewalls.Create(context.Background(), firewallRequest)
	if err != nil {
		fmt.Printf("Error creating firewall: %v\n", err)
		return ""
	}
	fmt.Printf("Firewall created: %v\n", firewall)
	return firewall.ID
}

func addDropletToFirewall(token string, firewallID string, dropletID int) {
	client := godo.NewFromToken(token)
	_, err := client.Firewalls.AddDroplets(context.Background(), firewallID, dropletID)
	if err != nil {
		fmt.Printf("Error adding droplet %d to firewall: %v\n", dropletID, err)
		return
	}
	fmt.Printf("Droplet %d added to firewall successfully.\n", dropletID)
}

func createSSHkey() int {
	client := godo.NewFromToken(doAPI)
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

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"stuffs.dev/deployfast/internal/config"
	"stuffs.dev/deployfast/internal/ssh"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	var rootCmd = &cobra.Command{
		Use:   "deployfast",
		Short: "DeployFast is a CLI tool for SSH operations",
		Long:  `A Fast and Flexible CLI tool for SSH operations built with Go.`,
	}

	var provisionCmd = &cobra.Command{
		Use:   "provision",
		Short: "Provision the remote server",
		Run: func(cmd *cobra.Command, args []string) {
			// Create SSH client
			sshClient, err := ssh.NewSSHClient(cfg.SSH)
			if err != nil {
				log.Fatalf("Failed to create SSH client: %v", err)
			}
			defer sshClient.Close()

			// Run provision script
			localScriptPath := "templates/provision.sh"
			remoteScriptPath := "/tmp/provision.sh"
			err = sshClient.RunRemoteScript(localScriptPath, remoteScriptPath)
			if err != nil {
				log.Fatalf("Failed to run provision script: %v", err)
			}

			fmt.Println("Provision script executed successfully")
		},
	}

	rootCmd.AddCommand(provisionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

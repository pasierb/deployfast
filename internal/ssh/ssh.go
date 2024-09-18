package ssh

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	"stuffs.dev/deployfast/internal/config"
)

// SSHClient struct to hold the ssh.Client
type SSHClient struct {
	client *ssh.Client
}

// NewSSHClient function to create a new SSHClient
func NewSSHClient(cfg config.SSHConfig) (*SSHClient, error) {
	// Prompt for password
	fmt.Printf("Enter SSH password for %s@%s: ", cfg.User, cfg.Host)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // Print a newline after the password input

	// SSH connection configuration
	config := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(string(password)),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to the SSH server
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), config)
	if err != nil {
		return nil, err
	}

	return &SSHClient{client: client}, nil
}

// RunCommand method for SSHClient
func (s *SSHClient) RunCommand(command string) (string, error) {
	// Create a session
	session, err := s.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	// Run the command
	fmt.Printf("Executing command: %s\n", command)
	output, err := session.CombinedOutput(command)
	if err != nil {
		fmt.Printf("Command failed: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
		return "", err
	}

	fmt.Printf("Command output: %s\n", string(output))
	return string(output), nil
}

// Close method for SSHClient
func (s *SSHClient) Close() error {
	return s.client.Close()
}

// TransferFile transfers a local file to the remote server using SFTP
func (s *SSHClient) TransferFile(localPath, remotePath string) error {
	// Create a new SFTP client
	sftpClient, err := sftp.NewClient(s.client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	// Open the local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// Create the remote file
	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer remoteFile.Close()

	// Copy the file content
	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// RunProvisionScript transfers and executes the provision.sh script
func (s *SSHClient) RunProvisionScript(localScriptPath string) (string, error) {
	remoteScriptPath := "/tmp/provision.sh"

	// Transfer the script
	err := s.TransferFile(localScriptPath, remoteScriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to transfer provision script: %w", err)
	}

	// Make the script executable
	_, err = s.RunCommand(fmt.Sprintf("chmod +x %s", remoteScriptPath))
	if err != nil {
		return "", fmt.Errorf("failed to make script executable: %w", err)
	}

	// Try to run the script with sudo
	output, err := s.RunCommand(fmt.Sprintf("sudo bash %s", remoteScriptPath))
	if err != nil {
		fmt.Printf("Failed to run script with sudo: %v\n", err)
		fmt.Println("Attempting to run without sudo...")

		// If sudo fails, try running without sudo
		output, err = s.RunCommand(fmt.Sprintf("bash %s", remoteScriptPath))
		if err != nil {
			return "", fmt.Errorf("failed to run provision script: %w", err)
		}
	}

	// Clean up the remote script
	_, cleanupErr := s.RunCommand(fmt.Sprintf("rm %s", remoteScriptPath))
	if cleanupErr != nil {
		fmt.Printf("Warning: Failed to remove remote script: %v\n", cleanupErr)
	}

	return output, nil
}

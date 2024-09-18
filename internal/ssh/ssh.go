package ssh

import (
	"bufio"
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

// RunRemoteScript transfers and executes a script on the remote server
func (s *SSHClient) RunRemoteScript(localScriptPath, remoteScriptPath string) error {
	// Transfer the script
	err := s.TransferFile(localScriptPath, remoteScriptPath)
	if err != nil {
		return fmt.Errorf("failed to transfer script: %w", err)
	}

	// Make the script executable
	_, err = s.RunCommand(fmt.Sprintf("chmod +x %s", remoteScriptPath))
	if err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	// Create a session
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Set up pipe for remote command's stdout
	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to set up stdout pipe: %w", err)
	}

	// Set up pipe for remote command's stderr
	stderr, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to set up stderr pipe: %w", err)
	}

	// Start the remote command
	cmd := fmt.Sprintf("sudo bash %s || bash %s", remoteScriptPath, remoteScriptPath)
	err = session.Start(cmd)
	if err != nil {
		return fmt.Errorf("failed to start provision script: %w", err)
	}

	// Stream output
	go streamOutput(stdout, os.Stdout)
	go streamOutput(stderr, os.Stderr)

	// Wait for the command to finish
	err = session.Wait()
	if err != nil {
		return fmt.Errorf("provision script failed: %w", err)
	}

	// Clean up the remote script
	_, cleanupErr := s.RunCommand(fmt.Sprintf("rm %s", remoteScriptPath))
	if cleanupErr != nil {
		fmt.Printf("Warning: Failed to remove remote script: %v\n", cleanupErr)
	}

	return nil
}

// streamOutput reads from src and writes to dst line by line
func streamOutput(src io.Reader, dst io.Writer) {
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		fmt.Fprintln(dst, scanner.Text())
	}
}

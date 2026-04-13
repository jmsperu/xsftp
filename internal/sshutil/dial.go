package sshutil

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"

	"github.com/jmsperu/sftpgo/internal/config"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
)

// Dial establishes an SSH connection and returns an SFTP client.
func Dial(conn config.Connection) (*ssh.Client, *sftp.Client, error) {
	authMethods := buildAuthMethods(conn)

	if len(authMethods) == 0 {
		return nil, nil, fmt.Errorf("no authentication method available")
	}

	sshCfg := &ssh.ClientConfig{
		User:            conn.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", conn.Host, conn.Port)
	sshClient, err := ssh.Dial("tcp", addr, sshCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("SSH dial %s: %w", addr, err)
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, nil, fmt.Errorf("SFTP client: %w", err)
	}

	return sshClient, sftpClient, nil
}

func buildAuthMethods(conn config.Connection) []ssh.AuthMethod {
	var methods []ssh.AuthMethod

	// 1. Explicit key file
	if conn.KeyFile != "" {
		if m := keyFileAuth(conn.KeyFile); m != nil {
			methods = append(methods, m)
		}
	}

	// 2. SSH agent
	if m := agentAuth(); m != nil {
		methods = append(methods, m)
	}

	// 3. Default key files
	home, _ := os.UserHomeDir()
	for _, name := range []string{"id_ed25519", "id_rsa", "id_ecdsa"} {
		p := filepath.Join(home, ".ssh", name)
		if _, err := os.Stat(p); err == nil {
			if m := keyFileAuth(p); m != nil {
				methods = append(methods, m)
			}
		}
	}

	// 4. Password (interactive prompt)
	methods = append(methods, ssh.PasswordCallback(func() (string, error) {
		fmt.Printf("Password for %s@%s: ", conn.User, conn.Host)
		pw, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return "", err
		}
		return string(pw), nil
	}))

	// 5. Keyboard-interactive (some servers use this instead of password)
	methods = append(methods, ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
		answers := make([]string, len(questions))
		for i, q := range questions {
			fmt.Print(q)
			if echos[i] {
				var ans string
				fmt.Scanln(&ans)
				answers[i] = ans
			} else {
				pw, err := term.ReadPassword(int(syscall.Stdin))
				fmt.Println()
				if err != nil {
					return nil, err
				}
				answers[i] = string(pw)
			}
		}
		return answers, nil
	}))

	return methods
}

func keyFileAuth(path string) ssh.AuthMethod {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		// Try with passphrase
		fmt.Printf("Passphrase for %s: ", path)
		pw, err2 := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err2 != nil {
			return nil
		}
		signer, err = ssh.ParsePrivateKeyWithPassphrase(data, pw)
		if err != nil {
			return nil
		}
	}
	return ssh.PublicKeys(signer)
}

func agentAuth() ssh.AuthMethod {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		return nil
	}
	conn, err := net.Dial("unix", sock)
	if err != nil {
		return nil
	}
	return ssh.PublicKeysCallback(agent.NewClient(conn).Signers)
}

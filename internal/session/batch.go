package session

import (
	"fmt"
	"strings"

	"github.com/jmsperu/sftpgo/internal/config"
	"github.com/jmsperu/sftpgo/internal/sshutil"
	"github.com/pkg/sftp"
)

func Batch(conn config.Connection, commands []string) error {
	sshClient, sftpClient, err := sshutil.Dial(conn)
	if err != nil {
		return err
	}
	defer sshClient.Close()
	defer sftpClient.Close()

	cwd, _ := sftpClient.Getwd()
	fmt.Printf("Connected to %s@%s:%d\n", conn.User, conn.Host, conn.Port)

	for i, line := range commands {
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]
		args := parts[1:]

		fmt.Printf("[%d/%d] %s\n", i+1, len(commands), line)

		switch cmd {
		case "cd":
			if len(args) > 0 {
				target := resolvePath(cwd, args[0])
				if _, err := sftpClient.Stat(target); err != nil {
					fmt.Printf("  Error: %v\n", err)
				} else {
					cwd = target
				}
			}

		case "ls":
			listRemote(sftpClient, cwd, args)

		case "pwd":
			fmt.Println(cwd)

		case "mkdir":
			if len(args) > 0 {
				target := resolvePath(cwd, args[0])
				if err := sftpClient.MkdirAll(target); err != nil {
					fmt.Printf("  Error: %v\n", err)
				}
			}

		case "rm":
			if len(args) > 0 {
				target := resolvePath(cwd, args[0])
				if err := sftpClient.Remove(target); err != nil {
					fmt.Printf("  Error: %v\n", err)
				}
			}

		case "rmdir":
			if len(args) > 0 {
				target := resolvePath(cwd, args[0])
				if err := sftpClient.RemoveDirectory(target); err != nil {
					fmt.Printf("  Error: %v\n", err)
				}
			}

		case "get":
			if len(args) > 0 {
				remote := resolvePath(cwd, args[0])
				local := args[0]
				if len(args) > 1 {
					local = args[1]
				}
				getFile(sftpClient, remote, local)
			}

		case "put":
			if len(args) > 0 {
				local := args[0]
				remote := resolvePath(cwd, args[0])
				if len(args) > 1 {
					remote = resolvePath(cwd, args[1])
				}
				putFile(sftpClient, local, remote)
			}

		default:
			fmt.Printf("  Unknown command: %s\n", cmd)
		}
	}

	fmt.Println("Batch complete.")
	return nil
}

func batchListRemote(client *sftp.Client, cwd string, args []string) {
	listRemote(client, cwd, args)
}

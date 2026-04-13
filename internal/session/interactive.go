package session

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmsperu/sftpgo/internal/config"
	"github.com/jmsperu/sftpgo/internal/progress"
	"github.com/jmsperu/sftpgo/internal/sshutil"
	"github.com/pkg/sftp"

	"github.com/chzyer/readline"
)

func Interactive(conn config.Connection) error {
	sshClient, sftpClient, err := sshutil.Dial(conn)
	if err != nil {
		return err
	}
	defer sshClient.Close()
	defer sftpClient.Close()

	cwd, _ := sftpClient.Getwd()
	fmt.Printf("Connected to %s@%s:%d\n", conn.User, conn.Host, conn.Port)
	fmt.Printf("Remote directory: %s\n", cwd)
	fmt.Println("Type 'help' for commands, 'quit' to exit.")

	completer := readline.NewPrefixCompleter(
		readline.PcItem("ls"),
		readline.PcItem("cd"),
		readline.PcItem("pwd"),
		readline.PcItem("get"),
		readline.PcItem("put"),
		readline.PcItem("mkdir"),
		readline.PcItem("rm"),
		readline.PcItem("rmdir"),
		readline.PcItem("stat"),
		readline.PcItem("help"),
		readline.PcItem("quit"),
		readline.PcItem("exit"),
		readline.PcItem("lcd"),
		readline.PcItem("lpwd"),
		readline.PcItem("lls"),
	)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       prompt(cwd),
		AutoComplete: completer,
	})
	if err != nil {
		return fmt.Errorf("readline: %w", err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt || err == io.EOF {
				fmt.Println("Bye.")
				return nil
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]
		args := parts[1:]

		switch cmd {
		case "quit", "exit":
			fmt.Println("Bye.")
			return nil

		case "help":
			printHelp()

		case "pwd":
			fmt.Println(cwd)

		case "lpwd":
			wd, _ := os.Getwd()
			fmt.Println(wd)

		case "cd":
			if len(args) == 0 {
				args = []string{"/"}
			}
			target := resolvePath(cwd, args[0])
			if _, err := sftpClient.Stat(target); err != nil {
				fmt.Fprintf(os.Stderr, "cd: %v\n", err)
			} else {
				cwd = target
				rl.SetPrompt(prompt(cwd))
			}

		case "lcd":
			if len(args) == 0 {
				home, _ := os.UserHomeDir()
				args = []string{home}
			}
			if err := os.Chdir(args[0]); err != nil {
				fmt.Fprintf(os.Stderr, "lcd: %v\n", err)
			} else {
				wd, _ := os.Getwd()
				fmt.Println(wd)
			}

		case "ls", "lls":
			if cmd == "lls" {
				listLocal(args)
			} else {
				listRemote(sftpClient, cwd, args)
			}

		case "get":
			if len(args) < 1 {
				fmt.Fprintln(os.Stderr, "Usage: get <remote-path> [local-path]")
				continue
			}
			remote := resolvePath(cwd, args[0])
			local := filepath.Base(remote)
			if len(args) > 1 {
				local = args[1]
			}
			getFile(sftpClient, remote, local)

		case "put":
			if len(args) < 1 {
				fmt.Fprintln(os.Stderr, "Usage: put <local-path> [remote-path]")
				continue
			}
			local := args[0]
			remote := resolvePath(cwd, filepath.Base(local))
			if len(args) > 1 {
				remote = resolvePath(cwd, args[1])
			}
			putFile(sftpClient, local, remote)

		case "mkdir":
			if len(args) < 1 {
				fmt.Fprintln(os.Stderr, "Usage: mkdir <dir>")
				continue
			}
			target := resolvePath(cwd, args[0])
			if err := sftpClient.MkdirAll(target); err != nil {
				fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
			}

		case "rm":
			if len(args) < 1 {
				fmt.Fprintln(os.Stderr, "Usage: rm <file>")
				continue
			}
			target := resolvePath(cwd, args[0])
			if err := sftpClient.Remove(target); err != nil {
				fmt.Fprintf(os.Stderr, "rm: %v\n", err)
			}

		case "rmdir":
			if len(args) < 1 {
				fmt.Fprintln(os.Stderr, "Usage: rmdir <dir>")
				continue
			}
			target := resolvePath(cwd, args[0])
			if err := sftpClient.RemoveDirectory(target); err != nil {
				fmt.Fprintf(os.Stderr, "rmdir: %v\n", err)
			}

		case "stat":
			if len(args) < 1 {
				fmt.Fprintln(os.Stderr, "Usage: stat <path>")
				continue
			}
			target := resolvePath(cwd, args[0])
			info, err := sftpClient.Stat(target)
			if err != nil {
				fmt.Fprintf(os.Stderr, "stat: %v\n", err)
			} else {
				fmt.Printf("  Name: %s\n  Size: %d\n  Mode: %s\n  Modified: %s\n",
					info.Name(), info.Size(), info.Mode(), info.ModTime().Format("2006-01-02 15:04:05"))
			}

		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s (type 'help')\n", cmd)
		}
	}
}

func prompt(cwd string) string {
	return fmt.Sprintf("sftp:%s> ", cwd)
}

func resolvePath(cwd, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(cwd, path))
}

func printHelp() {
	help := `Commands:
  ls [path]            List remote directory
  cd <path>            Change remote directory
  pwd                  Print remote working directory
  get <remote> [local] Download file
  put <local> [remote] Upload file
  mkdir <dir>          Create remote directory
  rm <file>            Remove remote file
  rmdir <dir>          Remove remote directory
  stat <path>          Show file info
  lcd <path>           Change local directory
  lpwd                 Print local working directory
  lls [path]           List local directory
  help                 Show this help
  quit / exit          Disconnect`
	fmt.Println(help)
}

func listRemote(client *sftp.Client, cwd string, args []string) {
	dir := cwd
	if len(args) > 0 {
		dir = resolvePath(cwd, args[0])
	}

	entries, err := client.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ls: %v\n", err)
		return
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, e := range entries {
		fmt.Printf("%s %10d %s %s\n",
			e.Mode(), e.Size(),
			e.ModTime().Format("Jan 02 15:04"),
			e.Name())
	}
}

func listLocal(args []string) {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lls: %v\n", err)
		return
	}

	for _, e := range entries {
		info, _ := e.Info()
		if info == nil {
			continue
		}
		fmt.Printf("%s %10d %s %s\n",
			info.Mode(), info.Size(),
			info.ModTime().Format("Jan 02 15:04"),
			info.Name())
	}
}

func getFile(client *sftp.Client, remotePath, localPath string) {
	info, err := client.Stat(remotePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get: %v\n", err)
		return
	}

	remoteFile, err := client.Open(remotePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get: %v\n", err)
		return
	}
	defer remoteFile.Close()

	localFile, err := os.Create(localPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get: %v\n", err)
		return
	}
	defer localFile.Close()

	bar := progress.NewBar(filepath.Base(remotePath), info.Size())
	pw := progress.NewWriter(bar, localFile)

	if _, err := io.Copy(pw, remoteFile); err != nil {
		fmt.Fprintf(os.Stderr, "\nget: %v\n", err)
		return
	}
	bar.Finish()
}

func putFile(client *sftp.Client, localPath, remotePath string) {
	info, err := os.Stat(localPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "put: %v\n", err)
		return
	}

	localFile, err := os.Open(localPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "put: %v\n", err)
		return
	}
	defer localFile.Close()

	remoteFile, err := client.Create(remotePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "put: %v\n", err)
		return
	}
	defer remoteFile.Close()

	bar := progress.NewBar(filepath.Base(localPath), info.Size())
	pr := progress.NewReader(bar, localFile)

	if _, err := io.Copy(remoteFile, pr); err != nil {
		fmt.Fprintf(os.Stderr, "\nput: %v\n", err)
		return
	}
	bar.Finish()
}

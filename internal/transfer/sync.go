package transfer

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/jmsperu/sftpgo/internal/config"
	"github.com/jmsperu/sftpgo/internal/sshutil"
	"github.com/pkg/sftp"
)

func Sync(conn config.Connection, localPath, remotePath string, deleteExtra bool) error {
	sshClient, sftpClient, err := sshutil.Dial(conn)
	if err != nil {
		return err
	}
	defer sshClient.Close()
	defer sftpClient.Close()

	// Track remote files for deletion
	remoteFiles := make(map[string]bool)

	// Walk local and upload changed files
	err = filepath.WalkDir(localPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(localPath, path)
		dest := remotePath + "/" + filepath.ToSlash(relPath)
		remoteFiles[dest] = true

		if d.IsDir() {
			sftpClient.MkdirAll(dest)
			return nil
		}

		localInfo, _ := d.Info()

		// Check if remote file exists and is up to date
		remoteInfo, err := sftpClient.Stat(dest)
		if err == nil {
			// File exists — skip if same size and not older
			if remoteInfo.Size() == localInfo.Size() && !remoteInfo.ModTime().Before(localInfo.ModTime()) {
				return nil
			}
		}

		fmt.Printf("Syncing: %s\n", relPath)
		return uploadFile(sftpClient, path, dest, localInfo.Size(), false)
	})
	if err != nil {
		return err
	}

	// Delete extra remote files
	if deleteExtra {
		deleteRemoteExtras(sftpClient, remotePath, remoteFiles)
	}

	fmt.Println("Sync complete.")
	return nil
}

func deleteRemoteExtras(client *sftp.Client, remotePath string, keep map[string]bool) {
	walker := client.Walk(remotePath)
	var toDelete []string

	for walker.Step() {
		if walker.Err() != nil {
			continue
		}
		if !keep[walker.Path()] {
			toDelete = append(toDelete, walker.Path())
		}
	}

	// Delete in reverse order (files before dirs)
	for i := len(toDelete) - 1; i >= 0; i-- {
		p := toDelete[i]
		info, err := client.Stat(p)
		if err != nil {
			continue
		}
		if info.IsDir() {
			client.RemoveDirectory(p)
		} else {
			client.Remove(p)
		}
		fmt.Printf("Deleted remote: %s\n", p)
	}
}

package transfer

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/jmsperu/sftpgo/internal/config"
	"github.com/jmsperu/sftpgo/internal/progress"
	"github.com/jmsperu/sftpgo/internal/sshutil"
	"github.com/pkg/sftp"
)

func Upload(conn config.Connection, localPath, remotePath string, resume bool) error {
	sshClient, sftpClient, err := sshutil.Dial(conn)
	if err != nil {
		return err
	}
	defer sshClient.Close()
	defer sftpClient.Close()

	info, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("stat local %s: %w", localPath, err)
	}

	if info.IsDir() {
		return uploadDir(sftpClient, localPath, remotePath, resume)
	}

	return uploadFile(sftpClient, localPath, remotePath, info.Size(), resume)
}

func uploadFile(client *sftp.Client, localPath, remotePath string, size int64, resume bool) error {
	// If remotePath is a directory, append the filename
	if info, err := client.Stat(remotePath); err == nil && info.IsDir() {
		remotePath = remotePath + "/" + filepath.Base(localPath)
	}

	var offset int64

	if resume {
		if info, err := client.Stat(remotePath); err == nil {
			offset = info.Size()
			if offset >= size {
				fmt.Printf("Already complete: %s\n", remotePath)
				return nil
			}
		}
	}

	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open local: %w", err)
	}
	defer localFile.Close()

	if offset > 0 {
		localFile.Seek(offset, io.SeekStart)
	}

	var remoteFile *sftp.File
	if offset > 0 {
		remoteFile, err = client.OpenFile(remotePath, os.O_WRONLY|os.O_APPEND)
	} else {
		remoteFile, err = client.Create(remotePath)
	}
	if err != nil {
		return fmt.Errorf("open remote: %w", err)
	}
	defer remoteFile.Close()

	bar := progress.NewBar(filepath.Base(localPath), size)
	if offset > 0 {
		bar.Add(offset)
	}

	pr := progress.NewReader(bar, localFile)
	if _, err := io.Copy(remoteFile, pr); err != nil {
		return fmt.Errorf("upload: %w", err)
	}

	bar.Finish()
	return nil
}

func uploadDir(client *sftp.Client, localPath, remotePath string, resume bool) error {
	return filepath.WalkDir(localPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(localPath, path)
		dest := remotePath + "/" + filepath.ToSlash(relPath)

		if d.IsDir() {
			client.MkdirAll(dest)
			return nil
		}

		info, _ := d.Info()
		if err := uploadFile(client, path, dest, info.Size(), resume); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s: %v\n", path, err)
		}
		return nil
	})
}

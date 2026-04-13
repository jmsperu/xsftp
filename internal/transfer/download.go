package transfer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jmsperu/sftpgo/internal/config"
	"github.com/jmsperu/sftpgo/internal/progress"
	"github.com/jmsperu/sftpgo/internal/sshutil"
	"github.com/pkg/sftp"
)

func Download(conn config.Connection, remotePath, localPath string, resume bool) error {
	sshClient, sftpClient, err := sshutil.Dial(conn)
	if err != nil {
		return err
	}
	defer sshClient.Close()
	defer sftpClient.Close()

	info, err := sftpClient.Stat(remotePath)
	if err != nil {
		return fmt.Errorf("stat remote %s: %w", remotePath, err)
	}

	if info.IsDir() {
		return downloadDir(sftpClient, remotePath, localPath, resume)
	}

	return downloadFile(sftpClient, remotePath, localPath, info.Size(), resume)
}

func downloadFile(client *sftp.Client, remotePath, localPath string, size int64, resume bool) error {
	// If localPath is a directory, use the remote filename
	if info, err := os.Stat(localPath); err == nil && info.IsDir() {
		localPath = filepath.Join(localPath, filepath.Base(remotePath))
	}

	var offset int64
	flags := os.O_WRONLY | os.O_CREATE

	if resume {
		if info, err := os.Stat(localPath); err == nil {
			offset = info.Size()
			if offset >= size {
				fmt.Printf("Already complete: %s\n", localPath)
				return nil
			}
			flags |= os.O_APPEND
		}
	} else {
		flags |= os.O_TRUNC
	}

	remoteFile, err := client.Open(remotePath)
	if err != nil {
		return fmt.Errorf("open remote: %w", err)
	}
	defer remoteFile.Close()

	if offset > 0 {
		if _, err := remoteFile.Seek(offset, io.SeekStart); err != nil {
			return fmt.Errorf("seek remote: %w", err)
		}
	}

	localFile, err := os.OpenFile(localPath, flags, 0644)
	if err != nil {
		return fmt.Errorf("open local: %w", err)
	}
	defer localFile.Close()

	bar := progress.NewBar(filepath.Base(remotePath), size)
	if offset > 0 {
		bar.Add(offset)
	}

	pw := progress.NewWriter(bar, localFile)
	if _, err := io.Copy(pw, remoteFile); err != nil {
		return fmt.Errorf("download: %w", err)
	}

	bar.Finish()
	return nil
}

func downloadDir(client *sftp.Client, remotePath, localPath string, resume bool) error {
	walker := client.Walk(remotePath)
	for walker.Step() {
		if walker.Err() != nil {
			continue
		}

		relPath, _ := filepath.Rel(remotePath, walker.Path())
		dest := filepath.Join(localPath, relPath)

		if walker.Stat().IsDir() {
			os.MkdirAll(dest, 0755)
			continue
		}

		os.MkdirAll(filepath.Dir(dest), 0755)
		if err := downloadFile(client, walker.Path(), dest, walker.Stat().Size(), resume); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s: %v\n", walker.Path(), err)
		}
	}
	return nil
}

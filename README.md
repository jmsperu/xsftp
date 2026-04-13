# sftpgo

Modern SFTP/SCP client with saved connections, interactive shell, batch commands, directory sync, and resumable transfers -- all in a single binary.

A clean CLI alternative to WinSCP and FileZilla.

## Install

### Binary download

Download the latest release for your platform from the [releases page](https://github.com/jmsperu/sftpgo/releases).

### go install

```sh
go install github.com/jmsperu/sftpgo@latest
```

### Build from source

```sh
git clone https://github.com/jmsperu/sftpgo.git
cd sftpgo
make build
```

## Quick start

```sh
# Save a connection
sftpgo add prod deploy@server.example.com -p 2222 -i ~/.ssh/id_ed25519

# Open interactive SFTP shell
sftpgo connect prod

# Upload / download
sftpgo put ./release.tar.gz prod:/opt/releases/
sftpgo get prod:/var/log/app.log ./logs/

# Sync a directory
sftpgo sync ./build prod:/var/www/html --delete
```

## Commands

### add

Save a named connection for later use.

```sh
sftpgo add <name> <user@host>
sftpgo add staging admin@staging.example.com
sftpgo add prod deploy@prod.example.com -p 2222 -i ~/.ssh/deploy_key
```

| Flag | Description |
|------|-------------|
| `-p, --port` | SSH port (default 22) |
| `-i, --key` | Path to private key file |

### connect

Open an interactive SFTP shell with tab completion. Accepts a saved connection name or `user@host`.

```sh
sftpgo connect prod
sftpgo connect admin@server.example.com
sftpgo connect admin@server.example.com -p 2222
```

| Flag | Description |
|------|-------------|
| `-p, --port` | SSH port override |
| `-i, --key` | Path to private key |

### get

Download a file or directory from a remote server.

```sh
sftpgo get prod:/etc/nginx/nginx.conf ./local/
sftpgo get prod:/var/backups/ ./backups/ --resume
```

| Flag | Description |
|------|-------------|
| `--resume` | Resume an interrupted transfer |

### put

Upload a file or directory to a remote server.

```sh
sftpgo put ./release.tar.gz prod:/opt/releases/
sftpgo put ./website/ prod:/var/www/html/ --resume
```

| Flag | Description |
|------|-------------|
| `--resume` | Resume an interrupted transfer |

### sync

Sync a local directory to a remote directory (upload only).

```sh
sftpgo sync ./build prod:/var/www/html
sftpgo sync ./build prod:/var/www/html --delete
```

| Flag | Description |
|------|-------------|
| `--delete` | Delete remote files not present locally |

### batch

Execute SFTP commands from a file, one per line. Lines starting with `#` are ignored.

```sh
sftpgo batch prod commands.txt
```

Supported batch commands: `ls`, `cd`, `get`, `put`, `mkdir`, `rm`, `pwd`.

### list (ls)

List all saved connections.

```sh
sftpgo list
sftpgo ls
```

## Global flags

| Flag | Description |
|------|-------------|
| `--config` | Config file path (default `~/.sftpgo.yml`) |

## Configuration

Saved connections are stored in `~/.sftpgo.yml`:

```yaml
connections:
  prod:
    host: server.example.com
    port: 22
    user: deploy
    key_file: ~/.ssh/deploy_key
  staging:
    host: staging.example.com
    port: 2222
    user: admin
```

## License

MIT

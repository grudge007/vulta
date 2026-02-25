# Vulta

**Vulta** is a high-performance, concurrent deployment and remote execution orchestrator. Built for engineers who need speed without the bloat of traditional configuration management, **Vulta** provides a streamlined way to synchronize code and execute commands across multiple nodes simultaneously using SSH and SFTP.

## Installation

**Vulta** is distributed via a custom Debian repository. You can install it on any compatible Linux system:

```bash
# Add the GPG key
curl -fsSL http://me.iamgrudge.online/gitz-repo.gpg \
| sudo gpg --dearmor -o /usr/share/keyrings/vulta.gpg

# Add the repository to your sources
echo "deb [signed-by=/usr/share/keyrings/vulta.gpg] http://me.iamgrudge.online stable main" \
| sudo tee /etc/apt/sources.list.d/vulta.list

# Update and install
sudo apt update && sudo apt install vulta

```

## Why Vulta?

Unlike heavy-duty automation suites, **Vulta** focuses on the "Developer Experience" (DX) for cloud engineers. It treats your infrastructure as a group of targets that you can hit with single-node precision (**Unicast**) or cluster-wide speed (**Broadcast**).

## Core Features

* **Native Concurrency:** Powered by Go goroutines to handle hundreds of node connections in parallel.
* **System-First Execution:** A built-in engine (`runz`) for executing raw shell commands directly on the target OS.
* **Minimalist Configuration:** Simple YAML-based inventory management. No complex DSL to learn.
* **Intelligent Sync:** Uses `.vultaignore` patterns to keep your deployments clean and relevant.
* **SSH Native:** Seamless integration with standard SSH private key authentication.

## Usage

### 1. Initialize

Set up a new project environment:

```bash
vulta init

```

### 2. Configure

Define your nodes in `.vulta/vulta.yaml`:

```yaml
project_name: MyCloudApp
nodes:
  - name: node1
    ip: 192.168.1.10
    user: root
    path: /var/www/app

```

### 3. Deploy and Execute

**Broadcast** to all nodes or **Unicast** to a specific IP:

```bash
# Push to everyone
vulta push

# Target a specific node
vulta push 192.168.1.10

# Run commands cluster-wide
vulta run "systemctl restart my-app"

```

## Contributing

The project is hosted at [github.com/grudge007/vulta](https://www.google.com/search?q=https://github.com/grudge007/vulta).

## License

This project is licensed under the **GNU General Public License v3.0 (GPL-3.0)**.

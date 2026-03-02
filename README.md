# Vulta: Fast & Simple Deployment

**Vulta** is a high-speed tool for engineers who need to ship code and run commands across multiple servers without the headache of massive, "enterprise" automation software. It uses SSH and SFTP to get the job done quickly and keeps things lightweight.

---

## Get Started

Vulta is currently available for Linux via a custom Debian repo. You can set it up in three steps:

```bash
# 1. Add the security key
curl -fsSL http://me.iamgrudge.online/vulta-repo.gpg | sudo gpg --dearmor -o /usr/share/keyrings/vulta.gpg

# 2. Add the repo to your system
echo "deb [signed-by=/usr/share/keyrings/vulta.gpg] http://me.iamgrudge.online:83 stable main" | sudo tee /etc/apt/sources.list.d/vulta.list

# 3. Install
sudo apt update && sudo apt install vulta

```

---

## What makes Vulta different?

Most automation tools are "bloated"—they require you to learn a whole new language just to move a file. Vulta focuses on **speed** and **simplicity**:

* **Fast as hell:** It uses Go’s native concurrency to talk to hundreds of servers at the same time.
* **No new languages:** You don't need to learn a complex DSL. If you know how to write a shell command, you know how to use Vulta.
* **Smart Syncing:** It tracks file hashes to avoid sending the same data twice, and respects your `.vultaignore` files.
* **Robust & Simple:** We've polished the error handling so it tells you exactly what went wrong (like missing SSH keys) without crashing.
* **Secure:** It works right on top of your existing SSH keys.

---

## How to use it

### 1. Start a project

Run this in your project folder to create the basic setup:

```bash
vulta init

```

### 2. Add your servers

List your servers in `.vulta/vulta.yaml`:

```yaml
project_name: MyCloudApp
nodes:
  - ip: 192.168.1.10
    user: root
    path: /var/www/app

```

### 3. Push and Run

You can send code to everyone at once or pick a single server.

* **Deploy to everyone:** `vulta push`
* **Deploy to one server:** `vulta -n 192.168.1.10 push`
* **Run a command everywhere:** `vulta run "systemctl restart my-app"`
* **Run a command on one server:** `vulta -n 192.168.1.10 run "systemctl restart my-app"`

---

## Links

* **Source Code:** [GitHub](https://www.google.com/search?q=https://github.com/grudge007/vulta)
* **License:** GNU General Public License v3.0
.

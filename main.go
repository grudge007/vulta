package main

import (
	"fmt"
	"gitz/initz"
	"gitz/pushz"
	"gitz/runz"

	"os"
	"strings"
)

const Version = "0.1.1"

func main() {
	var action string

	if len(os.Args) > 1 {
		action = os.Args[1]
	} else {
		fmt.Println("Usage:")
		fmt.Println("  gitz init [--force|-f]")
		fmt.Println("  gitz push")
		fmt.Println("  gitz run <command>")
		fmt.Println("  gitz --version")
		return
	}

	if action == "init" && len(os.Args) > 3 {
		fmt.Println("Error: too many arguments for 'gitz init'.")
		fmt.Println("Usage: gitz init [--force|-f]")
		return
	}

	switch action {
	case "init":
		force := false
		if len(os.Args) > 2 && (os.Args[2] == "--force" || os.Args[2] == "-f") {
			force = true
		}
		initz.InitGitz(force)

	case "push":
		nodeIp := "None"
		if len(os.Args) > 2 {
			nodeIp = os.Args[2]
		}
		loadedConfig := initz.NewInventory().LoadGitzConf()
		pushz.PushFilesToRemote(loadedConfig, nodeIp)

	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Error: no remote command specified.")
			fmt.Println("Usage: gitz run <command>")
			return
		}
		remoteCommand := strings.Join(os.Args[2:], " ")
		loadedConfig := initz.NewInventory().LoadGitzConf()
		runz.RunCommand(loadedConfig, remoteCommand)

	case "-v", "--version":
		fmt.Printf("gitz version %s\n", Version)
		return

	default:
		fmt.Printf("Error: unknown command '%s'.\n", action)
		fmt.Println("Usage:")
		fmt.Println("  gitz init [--force|-f]")
		fmt.Println("  gitz push")
		fmt.Println("  gitz run <command>")
		fmt.Println("  gitz --version")
	}
}

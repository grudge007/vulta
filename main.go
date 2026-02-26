package main

import (
	"flag"
	"fmt"
	"gitz/initz"
	"gitz/pushz"
	"gitz/runz"

	"strings"
)

const Version = "0.1.1"

func main() {
	var action string
	var forceInitPtr bool
	var nodePtr string
	var versionPtr bool

	flag.StringVar(&nodePtr, "node", "None", "Node IP")
	flag.StringVar(&nodePtr, "n", "None", "Node IP")

	flag.BoolVar(&versionPtr, "version", false, "Vesrion")
	flag.BoolVar(&versionPtr, "v", false, "Vesrion")

	flag.BoolVar(&forceInitPtr, "force", false, "Force Init")
	flag.BoolVar(&forceInitPtr, "f", false, "Force Init")

	flag.Parse()

	if versionPtr {
		fmt.Printf("gitz version %s\n", Version)
		return
	}

	args := flag.Args()

	if len(args) > 0 {
		action = args[0]
	} else {
		fmt.Println("Usage:")
		fmt.Println("  gitz init [--force|-f]")
		fmt.Println("  gitz push")
		fmt.Println("  gitz -node <node ip> push")
		fmt.Println("  gitz run <command>")
		fmt.Println("  gitz -node <node ip> run <command>")
		fmt.Println("  gitz --version")
		return
	}

	nodeIp := nodePtr

	if action == "init" && len(args) > 2 {
		fmt.Println("Error: too many arguments for 'gitz init'.")
		fmt.Println("Usage: gitz init [--force|-f]")
		return
	}

	switch action {
	case "init":
		initz.InitGitz(forceInitPtr)
	case "push":
		loadedConfig := initz.NewInventory().LoadGitzConf()
		pushz.PushFilesToRemote(loadedConfig, nodeIp)

	case "run":
		if len(args) < 2 {
			fmt.Println("Error: no remote command specified.")
			fmt.Println("Usage: gitz run <command>")
			return
		}
		remoteCommand := strings.Join(args[1:], " ")
		loadedConfig := initz.NewInventory().LoadGitzConf()
		runz.RunCommand(loadedConfig, remoteCommand, nodeIp)

	default:
		fmt.Printf("Error: unknown command '%s'.\n", action)
		fmt.Println("Usage:")
		fmt.Println("  gitz init [--force|-f]")
		fmt.Println("  gitz push")
		fmt.Println("  gitz run <command>")
		fmt.Println("  gitz --version")
	}
}

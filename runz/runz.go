package runz

import (
	"fmt"
	"os"
	"sync"
	"time"
	"vulta/initz"

	"golang.org/x/crypto/ssh"
)

type RunManager struct {
	Config initz.Inventory
	// RemoteCommand string
	Auth ssh.Signer
}

func RunCommand(loadedConfig *initz.Inventory, remoteCommand string, nodeIp string) {
	var wg sync.WaitGroup

	runCommand := CommandExec(*loadedConfig)
	if nodeIp != "None" {
		for i, node := range loadedConfig.Nodes {
			if node.IP == nodeIp {
				fmt.Println(runCommand.cmdRunner(remoteCommand, i))
			}
		}
	} else {
		for i := 0; i < len(loadedConfig.Nodes); i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				fmt.Println(runCommand.cmdRunner(remoteCommand, index))

			}(i)

		}
		wg.Wait()
	}
}

func CommandExec(inventory initz.Inventory) *RunManager {
	RunManager := &RunManager{
		Config: inventory,
	}
	// RemoteCommand =  getRemoteCommand(),
	RunManager.Auth = RunManager.getSshSigner()
	return RunManager
}

func (inventory *RunManager) getSshConnection(index int) *ssh.Client {
	config := &ssh.ClientConfig{
		User: inventory.Config.Nodes[index].User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(inventory.Auth),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	conn, err := ssh.Dial("tcp", inventory.Config.Nodes[index].IP+":22", config)
	if err != nil {
		fmt.Printf("[ERROR] SSH connection failed to %s: %v\n",
			inventory.Config.Nodes[index].IP, err)
		return nil
	}
	return conn
}

func (inventory *RunManager) getSshSigner() ssh.Signer {
	pvtKey, err := os.ReadFile(inventory.Config.PrivateKeyPath)
	if err != nil {
		fmt.Printf("[ERROR] Unable to read private key at %s: %v\n",
			inventory.Config.PrivateKeyPath, err)
		return nil
	}
	signer, err := ssh.ParsePrivateKey(pvtKey)
	if err != nil {
		fmt.Printf("[ERROR] Failed to parse private key at %s: %v\n",
			inventory.Config.PrivateKeyPath, err)
	}
	return signer
}

func (inventory *RunManager) cmdRunner(remoteCommand string, index int) string {
	nodeIP := inventory.Config.Nodes[index].IP

	connection := inventory.getSshConnection(index)
	if connection == nil {
		return fmt.Sprintf("[%s] Connection failed", nodeIP)
	}
	defer connection.Close()

	session, err := connection.NewSession()
	if err != nil {
		return fmt.Sprintf("[%s] Session failed: %v", nodeIP, err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(remoteCommand)
	if err != nil {
		return fmt.Sprintf("[%s] Error: %v\nOutput: %s", nodeIP, err, string(output))
	}

	return fmt.Sprintf("[%s] Success:\n%s", nodeIP, string(output))
}

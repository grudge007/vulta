package pushz

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"vulta/initz"
	"vulta/state"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type PushManager struct {
	Config      initz.Inventory
	FilesToPush []string
	Auth        ssh.Signer
	State       *state.DeploymentState
}

func PushFilesToRemote(loadedConfig *initz.Inventory, nodeIp string, quiet bool, files []string) {
	var wg sync.WaitGroup
	ds := state.LoadDeploymentState()
	myPushManager := NewPushManager(*loadedConfig, files, ds)
	if nodeIp != "None" {
		for i, node := range myPushManager.Config.Nodes {
			if node.IP == nodeIp {
				switch quiet {
				case false:
					fmt.Println(myPushManager.pushFiles(i))
				case true:
					myPushManager.pushFiles(i)

				}

			}
		}
	} else {

		sem := make(chan struct{}, 10)
		for i := 0; i < len(myPushManager.Config.Nodes); i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				sem <- struct{}{}        // Acquire
				defer func() { <-sem }() // Release

				output := myPushManager.pushFiles(index)
				if !quiet {
					fmt.Println(output)
				}
			}(i)

		}
		wg.Wait()

	}
}

func NewPushManager(inventory initz.Inventory, files []string, ds *state.DeploymentState) *PushManager {
	pushManager := &PushManager{
		Config: inventory,
		State:  ds,
	}
	pushManager.FilesToPush = pushManager.getLocalFiles(files)
	pushManager.Auth = pushManager.getSshSigner()
	return pushManager

}

// get ignore files

func (inventory *PushManager) getIgnoreFiles() []string {
	var ignoreFiles []string
	file := filepath.Join(inventory.Config.ProjectRoot, ".vulta", "vultaignore")
	data, err := os.Open(file)
	if err != nil {
		// vultaignore doesn't exist, no files to ignore
		return ignoreFiles
	}
	defer data.Close()
	scanner := bufio.NewScanner(data)
	for scanner.Scan() {
		line := scanner.Text()
		ignoreFiles = append(ignoreFiles, line)
	}
	return ignoreFiles

}

func (inventory *PushManager) getLocalFiles(files []string) []string {
	var filesToBeSent []string
	fmt.Println(files)
	if len(files) > 0 {
		for _, file := range files {
			absPath, err := filepath.Abs(file)
			if err != nil {
				fmt.Printf("Error in find absoulute path for %v, Err: %v", file, err)
			}
			err = filepath.WalkDir(absPath, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() {
					filesToBeSent = append(filesToBeSent, path)
					return nil
				}

				return nil

			})

		}
		// fmt.Println(filesToBeSent)
		return filesToBeSent

	} else {
		ignoreFiles := inventory.getIgnoreFiles()
		err := filepath.WalkDir(inventory.Config.ProjectRoot, func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			for _, fileName := range ignoreFiles {
				if info.Name() == fileName {
					if info.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
			}
			if info.IsDir() {
				return nil
			}

			filesToBeSent = append(filesToBeSent, path)
			return nil

		})
		if err != nil {
			fmt.Println(err)
		}
		return filesToBeSent

	}

}

func (inventory *PushManager) getSshSigner() ssh.Signer {
	pvtKey, err := os.ReadFile(inventory.Config.PrivateKeyPath)
	if err != nil {
		fmt.Printf("[ERROR] Unable to read private key at %s: %v\n", inventory.Config.PrivateKeyPath, err)
		return nil
	}
	signer, err := ssh.ParsePrivateKey(pvtKey)
	if err != nil {
		fmt.Printf("[ERROR] Failed to parse private key at %s: %v\n", inventory.Config.PrivateKeyPath, err)
		return nil
	}
	return signer
}

func (inventory *PushManager) getSshConnection(index int) *ssh.Client {
	pvtKey := inventory.Auth

	config := &ssh.ClientConfig{
		User: inventory.Config.Nodes[index].User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(pvtKey),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	conn, err := ssh.Dial("tcp", inventory.Config.Nodes[index].IP+":22", config)
	if err != nil {
		fmt.Printf("Failed to establish ssh connection to node %v, Err: %v\n", inventory.Config.Nodes[index].IP, err)
		return nil
	}
	return conn
}

func (inventory *PushManager) pushFiles(index int) string {
	node := inventory.Config.Nodes[index]
	var builder strings.Builder

	// 1. Use the shared state
	ds := inventory.State

	// 2. Batch check hashes
	filesToBeSent, hashes := ds.CompareHash(node.IP, inventory.FilesToPush)
	if len(filesToBeSent) == 0 {
		return fmt.Sprintf("Node %s is already up to date.\n", node.IP)
	}

	// 3. Setup Connection
	connection := inventory.getSshConnection(index)
	if connection == nil {
		return "Failed to Connect SSH"
	}
	defer connection.Close()

	client, err := sftp.NewClient(connection)
	if err != nil {
		return fmt.Sprintf("SFTP Client Error: %v\n", err)
	}
	defer client.Close()

	createdDirs := make(map[string]bool)
	var successfulFiles []string
	var successfulHashes []string

	for i, file := range filesToBeSent {
		relPath, err := filepath.Rel(inventory.Config.ProjectRoot, file)
		if err != nil {
			continue
		}

		remotePath := filepath.Join(node.Path, relPath)
		remoteDir := filepath.Dir(remotePath)

		// Only try to create the directory if we haven't in this session
		if !createdDirs[remoteDir] {
			_ = client.MkdirAll(remoteDir) // Ignore error if dir already exists
			createdDirs[remoteDir] = true
		}

		// Logic fix: Wrap file operations in a func or ensure they close immediately
		err = func() error {
			localFile, err := os.Open(file)
			if err != nil {
				return err
			}
			defer localFile.Close()

			remoteFile, err := client.Create(remotePath)
			if err != nil {
				return err
			}
			defer remoteFile.Close()

			_, err = io.Copy(remoteFile, localFile)
			return err
		}()

		if err == nil {
			builder.WriteString(fmt.Sprintf("Copied %v\n", file))
			successfulFiles = append(successfulFiles, file)
			successfulHashes = append(successfulHashes, hashes[i])
		} else {
			builder.WriteString(fmt.Sprintf("Failed %v: %v\n", file, err))
		}
	}

	// 4. Batch update state only for what worked
	if len(successfulFiles) > 0 {
		ds.UpdateHashTable(node.IP, successfulFiles, successfulHashes)
	}

	return builder.String()
}

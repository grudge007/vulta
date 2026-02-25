package pushz

import (
	"bufio"
	"fmt"
	"gitz/initz"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type PushManager struct {
	Config      initz.Inventory
	FilesToPush []string
	Auth        ssh.Signer
}

func PushFilesToRemote(loadedConfig *initz.Inventory, nodeIp string) {
	var wg sync.WaitGroup

	myPushManager := NewPushManager(*loadedConfig)
	if nodeIp != "None" {
		for i := 0; i < len(myPushManager.Config.Nodes); i++ {
			if nodeIp == myPushManager.Config.Nodes[i].IP {
				myPushManager.pushFiles(i)
			}

		}
	} else {

		for i := 0; i < len(myPushManager.Config.Nodes); i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				myPushManager.pushFiles(index)
			}(i)

		}
		wg.Wait()

	}
}

func NewPushManager(inventory initz.Inventory) *PushManager {
	pushManager := &PushManager{
		Config: inventory,
	}
	pushManager.FilesToPush = pushManager.getLocalFiles()
	pushManager.Auth = pushManager.getSshSigner()
	return pushManager

}

// get ignore files

func (inventory *PushManager) getIgnoreFiles() []string {
	var ignoreFiles []string
	file := filepath.Join(inventory.Config.ProjectRoot, ".gitz", "gitzignore")
	fmt.Println(file)
	data, _ := os.Open(file)
	scanner := bufio.NewScanner(data)
	for scanner.Scan() {
		line := scanner.Text()
		ignoreFiles = append(ignoreFiles, line)
	}
	data.Close()
	return ignoreFiles

}

func (inventory *PushManager) getLocalFiles() []string {
	ignoreFiles := inventory.getIgnoreFiles()
	var filesToBeSent []string
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

func (inventory *PushManager) getSshSigner() ssh.Signer {
	pvtKey, err := os.ReadFile(inventory.Config.PrivateKeyPath)
	if err != nil {
		fmt.Println(err)
	}
	signer, err := ssh.ParsePrivateKey(pvtKey)
	if err != nil {
		fmt.Println(err)
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
		fmt.Println(err)
		return nil
	}
	return conn
}

func (inventory *PushManager) pushFiles(index int) {
	createdDirs := make(map[string]bool)
	filesToBeSent := inventory.FilesToPush
	connection := inventory.getSshConnection(index)
	if connection == nil {
		return
	}
	defer connection.Close()

	client, err := sftp.NewClient(connection)
	if err != nil {
		fmt.Printf("this line caused %v\n", err)
	}
	defer client.Close()

	for _, file := range filesToBeSent {
		relPath, _ := filepath.Rel(inventory.Config.ProjectRoot, file)
		remotePath := filepath.Join(inventory.Config.Nodes[index].Path, relPath)
		remoteDir := filepath.Dir(remotePath)
		if !createdDirs[remoteDir] {
			err := client.MkdirAll(remoteDir)
			if err != nil {
				fmt.Println(err)
			}

			createdDirs[remoteDir] = true
		}
		localFile, err := os.Open(file)
		if err != nil {
			fmt.Println(err)
			continue
		}

		remoteFile, err := client.Create(remotePath)
		if err != nil {
			fmt.Println(err)
			localFile.Close()
			continue
		}

		_, err = io.Copy(remoteFile, localFile)
		if err != nil {
			fmt.Println(err)
			localFile.Close()
			remoteFile.Close()
			continue
		}
		fmt.Printf("Succesfully Copied %v to %v\n", localFile.Name(), remoteFile.Name())

		localFile.Close()
		remoteFile.Close()

	}

}

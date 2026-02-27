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

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type PushManager struct {
	Config      initz.Inventory
	FilesToPush []string
	Auth        ssh.Signer
}

func PushFilesToRemote(loadedConfig *initz.Inventory, nodeIp string, quite bool, files []string) {
	var wg sync.WaitGroup
	myPushManager := NewPushManager(*loadedConfig, files)
	if nodeIp != "None" {
		for i, node := range myPushManager.Config.Nodes {
			if node.IP == nodeIp {
				switch quite {
				case false:
					fmt.Println(myPushManager.pushFiles(i, files))
				case true:
					myPushManager.pushFiles(i, files)

				}

			}
		}
	} else {

		for i := 0; i < len(myPushManager.Config.Nodes); i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				myPushManager.pushFiles(index, files)
			}(i)

		}
		wg.Wait()

	}
}

func NewPushManager(inventory initz.Inventory, files []string) *PushManager {
	pushManager := &PushManager{
		Config: inventory,
	}
	pushManager.FilesToPush = pushManager.getLocalFiles(files)
	pushManager.Auth = pushManager.getSshSigner()
	return pushManager

}

// get ignore files

func (inventory *PushManager) getIgnoreFiles() []string {
	var ignoreFiles []string
	file := filepath.Join(inventory.Config.ProjectRoot, ".vulta", "vultaignore")
	// fmt.Println(file)
	data, _ := os.Open(file)
	scanner := bufio.NewScanner(data)
	for scanner.Scan() {
		line := scanner.Text()
		ignoreFiles = append(ignoreFiles, line)
	}
	data.Close()
	return ignoreFiles

}

func (inventory *PushManager) getLocalFiles(files []string) []string {
	var filesToBeSent []string
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
		fmt.Printf("Failed to establish ssh connection to node %v, Err: %v\n", inventory.Config.Nodes[index].IP, err)
		return nil
	}
	return conn
}

func (inventory *PushManager) pushFiles(index int, files []string) string {
	var builder strings.Builder
	var filesToBeSent []string

	createdDirs := make(map[string]bool)

	filesToBeSent = inventory.FilesToPush

	connection := inventory.getSshConnection(index)
	if connection == nil {
		return "Failed to Connect SSH"
	}
	defer connection.Close()

	client, err := sftp.NewClient(connection)
	if err != nil {
		fmt.Printf("this line caused %v\n", err)
	}
	defer client.Close()

	for _, file := range filesToBeSent {
		// fmt.Println("file: ", file)
		relPath, _ := filepath.Rel(inventory.Config.ProjectRoot, file)
		// fmt.Println("relPath: ", relPath)
		remotePath := filepath.Join(inventory.Config.Nodes[index].Path, relPath)
		// fmt.Println("remotePath: ", remotePath)
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
			fmt.Printf("Err: %v\n", err)
			continue
		}

		remoteFile, err := client.Create(remotePath)
		// fmt.Println(remotePath)
		if err != nil {
			fmt.Printf("Err: %v\n", err)
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

		builder.WriteString(fmt.Sprintf("Succesfully Copied %v to %v\n", localFile.Name(), remoteFile.Name()))

		localFile.Close()
		remoteFile.Close()

	}
	finalReport := builder.String()

	return finalReport

}

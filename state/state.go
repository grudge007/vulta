package state

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
	"vulta/initz"
	// "vulta/initz"
)

// Metadata for a single file on a specific node
type FileRecord struct {
	Hash     string `json:"hash"`
	LastPush string `json:"last_push"`
}

// The collection of files for a specific node
type NodeManifest struct {
	Files map[string]FileRecord `json:"files"`
}

// The top-level "Source of Truth"
type DeploymentState struct {
	mu    sync.RWMutex
	Nodes map[string]NodeManifest `json:"nodes"`
}

func TrackState() {

}

func LoadDeploymentState() *DeploymentState {
	data, err := os.ReadFile(".vulta/state.json")
	if err != nil {
		return &DeploymentState{
			Nodes: make(map[string]NodeManifest),
		}
	}
	var state DeploymentState
	err = json.Unmarshal(data, &state)
	if err != nil {
		return &DeploymentState{
			Nodes: make(map[string]NodeManifest),
		}
	}
	return &state
}

func (inventory *DeploymentState) calculateHash(file string) (string, error) {
	data, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer data.Close()

	hash := sha256.New()

	_, err = io.Copy(hash, data)

	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (inventory *DeploymentState) CompareHash(nodeIp string, files []string) ([]string, []string) {
	inventory.mu.RLock()
	defer inventory.mu.RUnlock()
	var dirtyFiles []string
	var dirtyHashes []string

	for _, file := range files {
		newHash, err := inventory.calculateHash(file)
		if err != nil {
			continue // Skip if file is unreadable
		}

		// Get the old hash safely
		var oldHash string
		if node, ok := inventory.Nodes[nodeIp]; ok {
			if record, exists := node.Files[file]; exists {
				oldHash = record.Hash
			}
		}
		if newHash != oldHash {
			dirtyFiles = append(dirtyFiles, file)
			dirtyHashes = append(dirtyHashes, newHash)
		}
	}
	return dirtyFiles, dirtyHashes
}

func (inventory *DeploymentState) UpdateHashTable(nodeip string, files []string, hashes []string) {
	inventory.mu.Lock()
	defer inventory.mu.Unlock()

	// Ensure the top‑level map exists
	if inventory.Nodes == nil {
		inventory.Nodes = make(map[string]NodeManifest)
	}

	// 1. Initialize node if it's new
	if _, ok := inventory.Nodes[nodeip]; !ok {
		inventory.Nodes[nodeip] = NodeManifest{Files: make(map[string]FileRecord)}
	}

	// 2. Map the batch to memory
	timestamp := time.Now().Format(time.RFC3339)
	for i, file := range files {
		inventory.Nodes[nodeip].Files[file] = FileRecord{Hash: hashes[i], LastPush: timestamp}
	}

	// 3. Save EVERYTHING to disk once
	data, _ := json.MarshalIndent(inventory, "", " ")
	os.WriteFile(".vulta/state.json", data, 0644)
}

func (inventory *DeploymentState) MakeStateFile(inv initz.Inventory) {
	inventory.mu.Lock()
	defer inventory.mu.Unlock()

	err := filepath.WalkDir(inv.ProjectRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if d.Name() == ".vulta" || d.Name() == ".git" {
				return filepath.SkipDir
			}
		}
		hash, err := inventory.calculateHash(path)

		if err != nil {
			return nil
		}

		for _, node := range inv.Nodes {
			if _, ok := inventory.Nodes[node.IP]; !ok {
				inventory.Nodes[node.IP] = NodeManifest{
					Files: map[string]FileRecord{},
				}
			}
			inventory.Nodes[node.IP].Files[path] = FileRecord{
				Hash:     hash,
				LastPush: "INITIALIZED",
			}
		}
		return nil
	})
	if err == nil {
		data, _ := json.MarshalIndent(inventory, "", " ")
		os.WriteFile(".vulta/state.json", data, 0644)
	}
}

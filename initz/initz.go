package initz

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Node struct {
	IP   string `yaml:"ip"`
	User string `yaml:"user"`
	Path string `yaml:"path"`
}

type Inventory struct {
	ProjectName    string `yaml:"project_name"`
	ProjectRoot    string `yaml:"project_root"`
	DefaultUser    string `yaml:"default_user"`
	DefaultPath    string `yaml:"default_path"`
	PrivateKeyPath string `yaml:"private_key_path"`
	Nodes          []Node `yaml:"nodes"`
}

// create constructor

func InitGitz(force bool) {
	project := NewInventory()
	if project.checkIsGitzExist(force) {
		project.convertToYaml()
	} else {
		os.Exit(1)
	}

}

func NewInventory() *Inventory {
	workingDir, _ := os.Getwd()
	return &Inventory{
		ProjectName:    "MyProject",
		ProjectRoot:    workingDir,
		DefaultUser:    "root",
		DefaultPath:    "/opt/MyProject",
		PrivateKeyPath: getSshPvtKeyPath(),
		Nodes: []Node{
			{
				IP:   "192.168.192.15",
				User: "root",
				Path: "/root/mypath",
			},
			{
				IP:   "192.168.192.30",
				User: "root",
				Path: "/opt/mypath",
			},
		},
	}

}

// yaml convertion helper
func (inventory *Inventory) convertToYaml() {

	dotGitz := inventory.getDotGitz()

	gitzConfigData, err := yaml.Marshal(&inventory)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// return gitzConfigFile
	err = os.WriteFile(dotGitz, gitzConfigData, 0644)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("successfully setup gitz, modify %v\n", dotGitz)
	}
}

// dotGitz creation helper
func (inventory *Inventory) getDotGitz() string {
	dotGitz := filepath.Join(inventory.ProjectRoot, ".gitz")
	err := os.MkdirAll(dotGitz, 0766)
	if err != nil {
		return ""
	}
	return filepath.Join(dotGitz, "gitz.yaml")

}

func getSshPvtKeyPath() string {
	homedir, _ := os.UserHomeDir()
	return filepath.Join(homedir, ".ssh/id_rsa")

}

func (inventory *Inventory) checkIsGitzExist(force bool) bool {
	_, err := os.Stat(inventory.getDotGitz())
	if err != nil || force {
		return true
	} else {
		fmt.Println("An existing configuration found")
		fmt.Println("use --force to override")
		os.Exit(1)
		return false
	}
}

func (inventory *Inventory) LoadGitzConf() *Inventory {
	inventoryFile, err := os.ReadFile(inventory.getDotGitz())
	if err != nil {
		fmt.Println("Failed To Load Config")
		os.Exit(1)
	}
	err = yaml.Unmarshal(inventoryFile, inventory)
	return inventory
}

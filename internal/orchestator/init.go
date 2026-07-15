package orchestator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yaml/go-yaml"
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

func (m *Manager) Init(force bool) error {
	if !m.isVultaExist(force) {
		return errors.New("")
	}

	err := m.createDotVulta()
	if err != nil {
		return err
	}

	err = m.convertToYaml()
	if err != nil {
		return err
	}

	fmt.Printf("successfully setup vulta, modify %v\n", m.Config.ConfFile)
	m.appendToGitIgnore()
	return nil
}

func (m *Manager) createDotVulta() error {
	_, err := os.Stat(m.Config.ConfDir)
	if err != nil {
		if err := os.Mkdir(m.Config.ConfDir, 0755); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) convertToYaml() error {
	inventory := Inventory{
		ProjectName:    "MyProject",
		ProjectRoot:    m.Config.WorkingDir,
		DefaultUser:    "root",
		DefaultPath:    "/opt/MyProject",
		PrivateKeyPath: getSSHPvtKeyPath(),
		Nodes: []Node{
			{
				IP:   "10.0.0.10",
				User: "root",
				Path: "/opt/MyProject",
			},

			{
				IP:   "10.0.0.11",
				User: "root",
				Path: "/opt/MyProject",
			},
		},
	}

	vultaConfig, err := yaml.Marshal(inventory)
	if err != nil {
		return err
	}

	err = os.WriteFile(m.Config.ConfFile, vultaConfig, 0644)
	if err != nil {
		return err
	}

	return nil
}

func getSSHPvtKeyPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/path/to/ssh/pvtkey"
	}
	return filepath.Join(home, ".ssh", "id_rsa.pub ")
}

func (m *Manager) appendToGitIgnore() {
	_, err := os.Stat(filepath.Join(m.Config.WorkingDir, ".git"))
	if err != nil {
		return
	}
	file, err := os.OpenFile(filepath.Join(m.Config.WorkingDir, ".gitignore"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}

	defer file.Close()

	if _, err := file.WriteString("/.vulta\n"); err != nil {
		return
	}
}

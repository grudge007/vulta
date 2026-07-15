package orchestator

import (
	"fmt"
	"os"
	"vulta/internal/config"
)

type Manager struct {
	Config *config.Config
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		Config: cfg,
	}
}

func (m *Manager) isVultaExist(force bool) bool {
	_, err := os.Stat(m.Config.ConfFile)
	if err != nil || force {
		return true
	}
	fmt.Println("An existing configuration found")
	fmt.Println("use --force to override")
	return false
}

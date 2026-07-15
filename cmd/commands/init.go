package commands

import (
	"vulta/internal/config"
	"vulta/internal/orchestator"

	"github.com/spf13/cobra"
)

var force bool

func init() {
	InitCmd.Flags().BoolVarP(&force, "force", "f", false, "reset existing configuration")
}

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Init vulta in the current working directory",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		cfg := config.InitConfig()
		m := orchestator.NewManager(cfg)

		err := m.Init(force)
		if err != nil {
			return err
		}

		return nil
	},
}

package apply

import (
	"github.com/frontierdigital/vertag/pkg/cmd/app/apply"
	"github.com/spf13/cobra"
)

// NewCmdApply creates a command to apply config
func NewCmdApply() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply config",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := apply.Apply(); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

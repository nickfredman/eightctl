package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/steipete/eightctl/internal/client"
	"github.com/steipete/eightctl/internal/tokencache"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear cached authentication token",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := client.New(
			viper.GetString("email"),
			viper.GetString("password"),
			viper.GetString("user_id"),
			viper.GetString("client_id"),
			viper.GetString("client_secret"),
		)
		if err := tokencache.Clear(c.Identity()); err != nil {
			return fmt.Errorf("clear token: %w", err)
		}
		fmt.Println("Logged out (token cache cleared)")
		return nil
	},
}

package cmd

import (
	"fmt"
	"namescore/asoc"
	"namescore/config"
	"namescore/utils"

	"os"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Shows status of namescore",
	Long: `This command return status of current namescore setup.
Following informations are provided:
- If configuration is correct.
- API key status.
- State of namescore process .`,
	Run: status,
}

func status(cmd *cobra.Command, args []string) {
	fmt.Println("namescore status")
	fmt.Println()
	cfg := config.Get()

	if cfg.ConfigFileExists() == false {
		fmt.Println("error: no config file present.")
		fmt.Println("Run \"namescore register\" first.")
		os.Exit(1)
	}

	if cfg.ReadFromFile() != nil {
		fmt.Println("error: failed to read configuration file")
		os.Exit(1)
	}

	if cfg.APIKey == "" {
		fmt.Println("error: no API key set.")
		fmt.Println("Create new with \"namescore register\"")
		os.Exit(1)
	} else if asoc.VerifyKey(cfg.APIKey) == false {
		fmt.Println("error: API key does not meet requirements.")
		os.Exit(1)
	} else {
		fmt.Println("API key present")
	}

	client := asoc.Client{Server: cfg.GetAlphaSocAddress()}
	client.SetKey(cfg.APIKey)

	status, err := client.AccountStatus()
	if err != nil {
		fmt.Println("error: Failed to check account status")
		os.Exit(1)
	}

	fmt.Println("Account registered:", status.Registered)
	fmt.Println("Account expired:", status.Expired)

	if utils.LockSocket() != nil {
		fmt.Println("namescore is running")
	} else {
		fmt.Println("namescore is not running")
	}

}

func init() {
	RootCmd.AddCommand(statusCmd)
}

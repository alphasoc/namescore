package cmd

import (
	"fmt"
	"os"

	"github.com/alphasoc/namescore/client"
	"github.com/alphasoc/namescore/config"
	"github.com/alphasoc/namescore/helpers"
	"github.com/spf13/cobra"
)

func newAccountRegisterCommand(configPath *string) *cobra.Command {
	var key string
	var cmd = &cobra.Command{
		Use:   "register",
		Short: "Acquire and register API key.",
		Long:  `This command provides interactive mode to retrieve API key and register it.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _, err := config.New(*configPath)
			if err != nil {
				return err
			}
			return register(cfg, *configPath, key)
		},
	}
	cmd.Flags().StringVar(&key, "key", "", "AlphaSOC api key")
	return cmd
}

func register(cfg *config.Config, configPath, key string) error {
	c, err := client.New(cfg.Alphasoc.Host, cfg.Alphasoc.APIVersion)
	if err != nil {
		return err
	}

	if key != "" {
		c.SetKey(key)
		fmt.Printf("Using key %s for registration.\n\n", key)
	} else if cfg.Alphasoc.APIKey != "" {
		c.SetKey(cfg.Alphasoc.APIKey)
		fmt.Printf("Using key %s for registration.\n\n", cfg.Alphasoc.APIKey)
	}

	if status, err := c.AccountStatus(); err == nil && status.Registered {
		fmt.Println("Account is already registered.")
		return nil
	}

	fmt.Println(`Provide your details to generate an API key and complete setup.
A valid email address is required to activate the key. 

By performing this request you agree to our Terms of Service and Privacy Policy
https://www.alphasoc.com/terms-of-service
`)
	name, email, organization, err := helpers.GetAccountRegisterDetails()
	if err != nil {
		return err
	}

	if key == "" && cfg.Alphasoc.APIKey == "" {
		keyReq, err := c.KeyRequest()
		if err != nil {
			return err
		}
		c.SetKey(keyReq.Key)
		cfg.Alphasoc.APIKey = keyReq.Key
	}

	var errSave error
	if configPath == "" {
		errSave = cfg.SaveDefault()
	} else {
		errSave = cfg.Save(configPath)
	}

	var req client.AccountRegisterRequest
	req.Details.Name = name
	req.Details.Email = email
	req.Details.Phone = "+48606690566"
	req.Details.Address = [3]string{"a", "b", "c"}
	req.Details.Organization = organization
	if err := c.AccountRegister(&req); err != nil {
		if errSave != nil {
			fmt.Fprintf(os.Stderr, `We were unable to register your account.
What's more there was problem with saving namescore config. In order to 
register account please run namescore again with following command
and follow the instructions:

$ namescore account register --key %s

Also put your config in /etc/namescore.yml for future usage. 
Config format below:

alphasoc:
  api_key: %s`, cfg.Alphasoc.APIKey, cfg.Alphasoc.APIKey)
			return errSave
		}

		fmt.Fprintf(os.Stderr, `We were unable to register your account.
Please run namescore again with following command and follow the instructions:

$ namescore`)
		return err
	}

	fmt.Println("Success! Check your email and click the verification link to activate your API key")
	return nil
}

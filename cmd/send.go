package cmd

import (
	"errors"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/alphasoc/namescore/client"
	"github.com/alphasoc/namescore/config"
	"github.com/alphasoc/namescore/executor"

	"github.com/spf13/cobra"
)

func newSendCommand() *cobra.Command {
	var configPath string
	var cmd = &cobra.Command{
		Use:   "send",
		Short: "send dns queries stored in pcap file",
		Long: `Read file in pcap fromat and send DNS queries to AlphaSOC for analyze
The queries could be save to file via tools like tcpdump or namescore in offline mode.
See namescore start --help for more informations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("at least 1 file required")
			}

			cfg, c, err := createConfigAndClient(configPath, true)
			if err != nil {
				return err
			}
			return send(cfg, c, args)
		},
	}
	cmd.Flags().StringVarP(&configPath, "config", "c", config.DefaultLocation, "Config path for namescore")
	return cmd
}

func send(cfg *config.Config, c client.Client, files []string) error {
	e, err := executor.New(c, cfg)
	if err != nil {
		return err
	}

	for i := range files {
		if err := e.Send(files[i]); err != nil {
			return err
		}
		if err := os.Rename(files[i], files[i]+"."+time.Now().Format(time.RFC3339)); err != nil {
			return err
		}
		log.Infof("file %s sent\n", files[i])
	}
	return nil
}

package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/syslog"
	"os"

	"github.com/alphasoc/namescore/asoc"
	"github.com/alphasoc/namescore/config"
	log "github.com/inconshreveable/log15"
	"github.com/spf13/cobra"
)

// registerCmd represents the register command
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Acquire and register API key.",
	Long:  `Acquire and register API key.`, //todo longer description, write what is needed
	Run:   register,
}

func init() {
	RootCmd.AddCommand(registerCmd)
}

const (
	noInput = "invalid user input"
)

func register(cmd *cobra.Command, args []string) {
	logger := log.New()
	if sysloghandler, err := log.SyslogHandler(syslog.LOG_USER|syslog.LOG_ERR, "namescore/register", log.TerminalFormat()); err != nil {
		logger.SetHandler(log.DiscardHandler())
	} else {
		logger.SetHandler(sysloghandler)
	}

	fmt.Println("namescore register")

	cfg := config.Get()
	if err := cfg.ReadFromFile(); err != nil {
		logger.Warn("Failed to read configuration file.", "err:", err)
		fmt.Println("Failed to read configuration file.")
		os.Exit(1)
	}

	client := asoc.Client{Server: cfg.AlphaSOCAddress}
	if cfg.APIKey != "" {
		client.SetKey(cfg.APIKey)
		status, err := client.AccountStatus()
		if err != nil {
			logger.Warn("Failed to check account status.", "err:", err)
			fmt.Println("Failed to check account status.")
			os.Exit(1)
		}
		if status.Registered {
			fmt.Println("Account is already registered.")
			os.Exit(0)
		}
	}

	if cfg.NetworkInterface == "" {
		fmt.Println("Provide network interface to be used by namescore:")
		cfg.NetworkInterface = readInterface()
	}

	data, err := readRegisterData(defaultUserInput())
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if cfg.APIKey == "" {
		key, err := client.KeyRequest()
		if err != nil {
			logger.Warn("Failed to get new API key from server.", "err:", err)
			fmt.Println("Failed to get new API key from server.")
			os.Exit(1)
		}
		logger.Info("New API key retrieved.")
		cfg.APIKey = key
	}
	client.SetKey(cfg.APIKey)

	if err := cfg.SaveToFile(); err != nil {
		logger.Warn("Failed to save config file.", "err:", err)
		fmt.Println("Failed to save config file.")
		os.Exit(1)
	}

	if err := client.Register(data); err != nil {
		logger.Warn("Failed to register account.", "err:", err)
		fmt.Println("Failed to register account.")
		os.Exit(1)
	}

	logger.Info("Account was successfully registered.")
	fmt.Println("Account was successfully registered.")
}

type userInput struct {
	reader  io.Reader
	writer  io.Writer
	scanner *bufio.Scanner
}

func defaultUserInput() *userInput {
	u := &userInput{reader: os.Stdin, writer: os.Stdout}
	u.scanner = bufio.NewScanner(u.reader)
	return u
}

func readInterface() string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return scanner.Text()
}

func (u *userInput) get(text string, mandatory bool) (string, error) {
	if _, err := fmt.Fprintf(u.writer, "%s", text); err != nil {
		return "", err
	}
	u.scanner.Scan()
	line := u.scanner.Text()

	if err := u.scanner.Err(); err != nil {
		return "", err
	}

	if mandatory && line == "" {
		return "", errors.New(noInput)
	}
	return line, nil
}

func readRegisterData(userIn *userInput) (rq *asoc.RegisterReq, err error) {

	fmt.Fprintln(userIn.writer, "Provide necessary data for API key registration.")

	rq = &asoc.RegisterReq{}

	if rq.Details.Name, err = userIn.get("Name: ", true); err != nil {
		return nil, err
	}

	if rq.Details.Organization, err = userIn.get("Organization: ", true); err != nil {
		return nil, err
	}

	if rq.Details.Email, err = userIn.get("email: ", true); err != nil {
		return nil, err
	}

	if rq.Details.Phone, err = userIn.get("phone: ", true); err != nil {
		return nil, err
	}

	if rq.Details.Address[0], err = userIn.get("Address (1/3): ", true); err != nil {
		return nil, err
	}

	if rq.Details.Address[1], err = userIn.get("Address (2/3): ", false); err != nil {
		return nil, err
	}

	if rq.Details.Address[2], err = userIn.get("Address (3/3): ", false); err != nil {
		return nil, err
	}

	return rq, nil
}

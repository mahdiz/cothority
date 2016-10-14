/*
* Prifi-app starts the cothority in either trustee, relay or client mode.
 */
package main

import (
	"os"

	"path"

	"fmt"
	"os/user"
	"runtime"

	"github.com/dedis/cothority/app/lib/config"
	"github.com/dedis/cothority/app/lib/server"
	"github.com/dedis/cothority/log"
	"github.com/dedis/cothority/sda"
	"github.com/dedis/cothority/services/prifi"
	"gopkg.in/urfave/cli.v1"
)

// DefaultName is the name of the binary we produce and is used to create a directory
// folder with this name
const DefaultName = "cothorityd"

// DefaultServerConfig is the default name of a server configuration file
const DefaultServerConfig = "config.toml"

func main() {
	app := cli.NewApp()
	app.Name = "Template"
	app.Usage = "Used for building other apps."
	app.Version = "0.1"
	app.Commands = []cli.Command{
		{
			Name:    "setup",
			Aliases: []string{"s"},
			Usage:   "setup the configuration for the server",
			Action:  setupCothorityd,
		},
		{
			Name:    "trustee",
			Usage:   "start in trustee mode",
			Aliases: []string{"t"},
			Action:  trustee,
		},
		{
			Name:      "relay",
			Usage:     "start in relay mode",
			ArgsUsage: "group [id-name]",
			Aliases:   []string{"r"},
			Action:    relay,
		},
		{
			Name:    "client",
			Usage:   "start in client mode",
			Aliases: []string{"c"},
			Action:  client,
		},
	}
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "debug, d",
			Value: 0,
			Usage: "debug-level: 1 for terse, 5 for maximal",
		},
		cli.StringFlag{
			Name:  "config, c",
			Value: getDefaultConfigFile(),
			Usage: "configuration-file",
		},
		cli.BoolFlag{
			Name:  "nowait",
			Usage: "Return immediately",
		},
	}
	app.Before = func(c *cli.Context) error {
		log.SetDebugVisible(c.Int("debug"))
		return nil
	}
	app.Run(os.Args)

}

// Start the cothority in trustee-mode using the already stored configuration.
func trustee(c *cli.Context) error {
	log.Info("Starting trustee")
	host, err := cothorityd(c)
	log.ErrFatal(err)
	prifi := host.GetService(prifi.ServiceName).(*prifi.Service)
	// Do other setups
	log.ErrFatal(prifi.StartTrustee())

	// Wait for the end of the world
	if !c.GlobalBool("nowait") {
		host.WaitForClose()
	}
	return nil
}

// Start the cothority in relay-mode using the already stored configuration.
func relay(c *cli.Context) error {
	log.Info("Starting relay")
	host, err := cothorityd(c)
	log.ErrFatal(err)
	prifi := host.GetService(prifi.ServiceName).(*prifi.Service)
	// Do other setups
	if c.NArg() == 0 {
		log.Fatal("Please give a group-definition")
	}
	group := getGroup(c)
	log.ErrFatal(prifi.StartRelay(group))

	// Wait for the end of the world
	if !c.GlobalBool("nowait") {
		host.WaitForClose()
	}
	return nil
}

// Start the cothority in client-mode using the already stored configuration.
func client(c *cli.Context) error {
	log.Info("Starting client")
	host, err := cothorityd(c)
	log.ErrFatal(err)
	prifi := host.GetService(prifi.ServiceName).(*prifi.Service)
	// Do other setups
	log.ErrFatal(prifi.StartClient())

	// Wait for the end of the world
	if !c.GlobalBool("nowait") {
		host.WaitForClose()
	}
	return nil
}

// Sets up a new cothorityd which is used for every type of prifi-mode.
func setupCothorityd(c *cli.Context) error {
	server.InteractiveConfig("cothorityd")
	return nil
}

// Starts the cothorityd to enable communication with the prifi-service.
// Returns the prifi-service.
func cothorityd(c *cli.Context) (*sda.Host, error) {
	// first check the options
	cfile := c.GlobalString("config")

	if _, err := os.Stat(cfile); os.IsNotExist(err) {
		return nil, err
	}
	// Let's read the config
	_, host, err := config.ParseCothorityd(cfile)
	if err != nil {
		return nil, err
	}
	host.ListenAndBind()
	host.StartProcessMessages()
	return host, nil
}

func getDefaultConfigFile() string {
	u, err := user.Current()
	// can't get the user dir, so fallback to current working dir
	if err != nil {
		fmt.Print("[-] Could not get your home's directory. Switching back to current dir.")
		if curr, err := os.Getwd(); err != nil {
			log.Fatalf("Impossible to get the current directory. %v", err)
		} else {
			return path.Join(curr, DefaultServerConfig)
		}
	}
	// let's try to stick to usual OS folders
	switch runtime.GOOS {
	case "darwin":
		return path.Join(u.HomeDir, "Library", DefaultName, DefaultServerConfig)
	default:
		return path.Join(u.HomeDir, ".config", DefaultName, DefaultServerConfig)
		// TODO WIndows ? FreeBSD ?
	}
}

// Reads the group-file and returns it
func getGroup(c *cli.Context) *config.Group {
	gfile := c.Args().Get(0)
	gr, err := os.Open(gfile)
	log.ErrFatal(err)
	defer gr.Close()
	groups, err := config.ReadGroupDescToml(gr)
	log.ErrFatal(err)
	if groups == nil || groups.Roster == nil || len(groups.Roster.List) == 0 {
		log.Fatal("No servers found in roster from", gfile)
	}
	return groups
}

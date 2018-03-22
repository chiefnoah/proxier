package main

import (
	"fmt"
	//"net/http/httputil"
	"github.com/urfave/cli"
	"os"
	"log"
)

func main() {
	app := cli.NewApp()
	app.Name = "proxier"
	app.Usage = "A commandline development path proxier"
	app.Action = func(c *cli.Context) error {
		fmt.Printf("Hello world!\n %+s", c.Args().Get(0))
		port := c.String("port")
		paths := c.StringSlice("path")
		dests := c.StringSlice("dest")
		fmt.Printf("\nPort: %s", port)
		fmt.Printf("\nPaths: %+s", paths)
		return nil
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "port, P"},
		cli.StringSliceFlag{Name: "path, p"},
		cli.StringSliceFlag{Name: "dest, d"},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
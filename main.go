package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dc0d/argify"
	"github.com/urfave/cli"
)

func main() {
	app()
}

func app() {
	// if err := cnf.LoadHCL(&conf); err != nil {
	// 	log.Println("warn:", err)
	// }

	app := cli.NewApp()

	{
		app.Version = "0.0.1"
		app.Author = "dc0d"
		app.Copyright = "dc0d"
		now := time.Now()
		app.Description = fmt.Sprintf(
			"Build Time:  %v %v\n   Go:          %v\n   Commit Hash: %v\n   Git Tag:     %v",
			now.Weekday(),
			BuildTime,
			GoVersion,
			CommitHash,
			GitTag)
		app.Name = "glint"
		app.Usage = ""
	}

	{
		helpers := cli.Command{
			Name:    "helpers",
			Action:  cmdHelpers,
			Aliases: []string{"hp"},
		}

		app.Commands = append(app.Commands, helpers)
	}

	argify.NewArgify().Build(app, &conf)

	if err := app.Run(os.Args); err != nil {
		log.Fatalln("error:", err)
	}
}

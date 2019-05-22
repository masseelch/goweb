package main

import (
	"github.com/masseelch/goweb/goweb"
	"github.com/urfave/cli"
	"log"
	"os"
)

func main() {
	//c, err := config.Load("config/config.yaml")
	//if err != nil {
	//	log.Fatal(err)
	//}

	app := cli.NewApp()

	app.Name = "goweb"
	app.Usage = "make the web go again!"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		{
			Name: "generate",
			Usage: "Generate code for repositories and more",
			Subcommands: []cli.Command{
				{
					Name: "repository",
					Usage: "Generate repository for a model",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: goweb.FlagSource,
							Usage: "Path to `FILE` containing the model declaration",
						},
						cli.StringFlag{
							Name: goweb.FlagTemplatePathRepositoryInterface,
							Usage: "Path to `FILE` containing the interface template",
							Value: goweb.TemplatePathRepositoryInterface,
						},
						cli.StringFlag{
							Name: goweb.FlagTemplatePathRepositoryImplementation,
							Usage: "Path to `FILE` containing the implementation template",
							Value: goweb.TemplatePathRepositoryImplementation,
						},
					},
					Action: goweb.GenerateRepository,
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

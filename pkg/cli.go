package kandie

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

type App struct {
	target string
	clictx *cli.Context
	kc     *KubeClient
}

// cli return an app
func (app *App) Construct() *cli.App {
	return &cli.App{
		Name:                 "kandie",
		Usage:                "Describe kubernetes resources",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "namespace",
				Aliases: []string{"n"},
				Usage:   "If present, the namespace scope for this CLI request",
			},

			&cli.StringFlag{
				Name:  "context",
				Usage: "The name of the kubeconfig context to use",
			},
			&cli.StringFlag{
				Name:    "kubeconfig",
				Usage:   "Path to the kubeconfig file to use for CLI requests.",
				EnvVars: []string{"KUBECONFIG"},
				Value:   "~/.kube/config",
			},
		},
		Action: func(c *cli.Context) error {
			ctx := context.Background()
			app.clictx = c
			res := app.clictx.Args().Get(0)
			app.target = app.clictx.Args().Get(1)
			resSplit := strings.Split(res, "/")

			if len(resSplit) > 1 {
				res = resSplit[0]
				app.target = resSplit[1]
			}
			app.kc = &KubeClient{
				kubeConfigPath: c.String("kubeconfig"),
				namespace:      c.String("namespace"),
				kubeContext:    c.String("context"),
			}
			if err := app.kc.create(); err != nil {
				return err
			}
			if res == "pod" {
				return app.doPod(ctx)
			}
			return fmt.Errorf("i don't know how to kandie: %s", res)
		},
	}
}

func Run(args []string) error {
	app := App{
		kc: &KubeClient{},
	}
	return app.Construct().Run(args)
}

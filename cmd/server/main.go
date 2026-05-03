package main

import (
	"fmt"
	"goph-profile/internal/app"
	"goph-profile/internal/config"
	"log"
	"os"
)

func main() {
	ctx, cancel := app.NewContext()
	defer cancel()

	cfg, err := config.New()
	if err != nil {
		log.Fatal(fmt.Errorf("failed to parse config: %w", err))
	}

	inst := app.NewInstance(cfg)
	err = inst.Bootstrap(ctx)
	if err != nil {
		os.Exit(1)
	}

	inst.Run(ctx)
}

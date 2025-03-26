package main

import (
	"fmt"
	"github.com/maevlava/Gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
	}
	err = cfg.SetUser("lane")
	if err != nil {
		return
	}
	cfg, err = config.Read()
	if err != nil {
		return
	}
	fmt.Println(cfg)
}

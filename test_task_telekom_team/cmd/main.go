package main

import (
	"bufio"
	"flag"
	"os"
	"strings"
	"time"

	"test.task.telekom.game/internal/config"
	"test.task.telekom.game/internal/executor"
	"test.task.telekom.game/internal/game"
	"test.task.telekom.game/internal/util"
)

const pathToCongig = "../config.json"

func main() {
	var path string

	flag.StringVar(&path, "path", pathToCongig, "path of config file")

	flag.Parse()

	cfg := config.MustLoad(path)

	scanner := bufio.NewScanner(os.Stdin)

	var content strings.Builder

	for scanner.Scan() {
		content.WriteString(scanner.Text())
		content.WriteString("\n")
	}

	openAt, err := util.ParseTimeDuration(cfg.OpenAt)
	if err != nil {
		return
	}

	game, err := game.New(cfg.Floors, cfg.Monsters, openAt, openAt+time.Duration(cfg.Duration)*time.Hour)
	if err != nil {
		return
	}

	executor.Execute(game, content.String())
}

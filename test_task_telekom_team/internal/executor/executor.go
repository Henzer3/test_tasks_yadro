package executor

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"test.task.telekom.game/internal/game"
	"test.task.telekom.game/internal/util"
)

func Execute(game *game.Game, content string) {
	lines := strings.Split(content, "\n")

	for i, v := range lines {
		if v == "" {
			continue
		}

		fields := strings.Fields(v)
		if len(fields) < 3 {
			fmt.Printf("bad content in line: %d\n", i+1)
			continue
		}
		s := fields[0]
		s = s[1 : len(s)-1]
		time, err := util.ParseTimeDuration(s)
		if err != nil {
			fmt.Printf("bad content in line: %d\n", i+1)
			continue
		}

		playerID, err := strconv.Atoi(fields[1])
		if err != nil {
			fmt.Printf("bad content in line: %d\n", i+1)
			continue
		}

		eventID, err := strconv.Atoi(fields[2])
		if err != nil {
			fmt.Printf("bad content in line: %d\n", i+1)
			continue
		}

		var extraField int
		if len(fields) > 3 {
			v, err := strconv.Atoi(fields[3])
			if err != nil {
				fmt.Printf("bad content in line: %d\n", i+1)
				continue
			}
			extraField = v
		}

		makeAction(game, playerID, time, eventID, extraField)
	}

	fmt.Println(game.GameLogs())
}

func makeAction(game *game.Game, id int, t time.Duration, eventID int, extraField int) {
	switch eventID {
	case 1:
		fmt.Println(game.RegisterPlayer(id, t).Error())
	case 2:
		fmt.Println(game.EnterDungeon(id, t).Error())
	case 3:
		fmt.Println(game.KillMonster(id, t).Error())
	case 4:
		fmt.Println(game.EnterNextFloor(id, t).Error())
	case 5:
		fmt.Println(game.EnterPreviosFloor(id, t).Error())
	case 6:
		fmt.Println(game.EnterBossFloor(id, t).Error())
	case 7:
		fmt.Println(game.KillBoss(id, t).Error())
	case 8:
		fmt.Println(game.LeftDungeon(id, t).Error())
	case 9:
		fmt.Println(game.PlayerDisconnect(id, t, extraField).Error())
	case 10:
		fmt.Println(game.RestoreHP(id, t, extraField).Error())
	case 11:
		fmt.Println(game.RecieveDamage(id, t, extraField).Error())
	}
}

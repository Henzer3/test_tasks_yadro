package game

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"test.task.telekom.game/internal/util"
)

type Game struct {
	Floors          int
	Monsters        int
	OpenAt          time.Duration
	DungeonDuration time.Duration
	players         map[int]*Player
}

func New(floors int, monsters int, openAt time.Duration, duration time.Duration) (*Game, error) {
	// floors >= 2 так как должен быть простой этаж и этаж с боссом
	if floors < 2 || monsters < 0 {
		return nil, fmt.Errorf("BadArguments")
	}
	return &Game{
		Floors:          floors - 1,
		Monsters:        monsters,
		OpenAt:          openAt,
		DungeonDuration: duration,
		players:         make(map[int]*Player),
	}, nil
}

func (g *Game) RegisterPlayer(playerID int, eventTime time.Duration) error {
	if g.isRegister(playerID) {
		return fmt.Errorf("player %d already exist", playerID)
	}

	g.players[playerID] = NewPlayer(playerID, g.Floors, g.Monsters)
	return fmt.Errorf("%s Player [%d] registered", util.FormatDuration(eventTime), playerID)
}

func (g *Game) isRegister(playerID int) bool {
	_, ok := g.players[playerID]
	return ok
}

func (g *Game) EnterDungeon(playerID int, eventTime time.Duration) error {
	if err := g.generalCheck(playerID, eventTime); err != nil {
		return err
	}

	if g.players[playerID].GetInDungeon() {
		return fmt.Errorf("%s Player [%d] makes impossible move [2]", util.FormatDuration(eventTime), playerID)
	}

	g.players[playerID].ChangeInDungeon(true)
	g.players[playerID].SetEnterDungeonTime(eventTime)
	g.players[playerID].SetFloorStarTime(0, eventTime)
	return fmt.Errorf("%s Player [%d] entered the dungeon", util.FormatDuration(eventTime), playerID)
}

func (g *Game) LeftDungeon(playerID int, eventTime time.Duration) error {
	if err := g.generalCheck(playerID, eventTime); err != nil {
		return err
	}

	if !g.players[playerID].GetInDungeon() {
		return fmt.Errorf("%s Player [%d] makes impossible move [8]", util.FormatDuration(eventTime), playerID)
	}

	g.players[playerID].ChangeInDungeon(false)
	g.players[playerID].SetLeftDungeonTime(eventTime)
	return fmt.Errorf("%s Player [%d] lefted the dungeon", util.FormatDuration(eventTime), playerID)
}

func (g *Game) isClosed(eventTime time.Duration) error {
	if eventTime < g.OpenAt || eventTime > g.OpenAt+g.DungeonDuration {
		return fmt.Errorf("Dungeon is closed")
	}
	return nil
}

func (g *Game) EnterNextFloor(playerID int, eventTime time.Duration) error {
	if err := g.generalCheck(playerID, eventTime); err != nil {
		return err
	}

	v := g.players[playerID].GetFloor()
	if v+1 >= g.Floors {
		return fmt.Errorf("%s Player [%d] makes impossible move [4]", util.FormatDuration(eventTime), playerID)
	}

	g.players[playerID].ChangeFloor(v + 1)
	g.players[playerID].SetFloorStarTime(v+1, eventTime)
	return fmt.Errorf("%s Player [%d] entered next floor", util.FormatDuration(eventTime), playerID)
}

func (g *Game) EnterPreviosFloor(playerID int, eventTime time.Duration) error {
	if err := g.generalCheck(playerID, eventTime); err != nil {
		return err
	}

	v := g.players[playerID].GetFloor()

	if v-1 < 0 {
		return fmt.Errorf("%s Player [%d] makes impossible move [5]", util.FormatDuration(eventTime), playerID)
	}

	g.players[playerID].ChangeFloor(v - 1)
	g.players[playerID].SetFloorStarTime(v-1, eventTime)
	return fmt.Errorf("%s Player [%d] entered previos floor", util.FormatDuration(eventTime), playerID)
}

func (g *Game) EnterBossFloor(playerID int, eventTime time.Duration) error {
	if err := g.generalCheck(playerID, eventTime); err != nil {
		return err
	}

	v := g.players[playerID].GetFloor()

	if v == g.Floors {
		return fmt.Errorf("%s Player [%d] makes impossible move [6]", util.FormatDuration(eventTime), playerID)
	}

	g.players[playerID].ChangeFloor(v)
	g.players[playerID].SetEnterBossRoom(eventTime)
	return fmt.Errorf("%s Player [%d] entered the boss's floor", util.FormatDuration(eventTime), playerID)
}

func (g *Game) KillBoss(playerID int, eventTime time.Duration) error {
	if err := g.generalCheck(playerID, eventTime); err != nil {
		return err
	}

	if g.players[playerID].GetFloor() != g.Floors {
		return fmt.Errorf("%s Player [%d] makes impossible move [7]", util.FormatDuration(eventTime), playerID)
	}

	if g.players[playerID].GetBossKilled() {
		return fmt.Errorf("%s Player [%d] makes impossible move [7]", util.FormatDuration(eventTime), playerID)
	}

	g.players[playerID].ChangeBossKilled(true)
	g.players[playerID].SetBossKillTime(eventTime)
	return fmt.Errorf("%s Player [%d] killed the boss", util.FormatDuration(eventTime), playerID)
}

func (g *Game) RecieveDamage(playerID int, eventTime time.Duration, hp int) error {
	if err := g.generalCheck(playerID, eventTime); err != nil {
		return err
	}

	if g.players[playerID].GetHP()-hp <= 0 {
		g.players[playerID].ChangeState("FAIL")
		g.players[playerID].ChangeHP(0)
		g.players[playerID].SetLeftDungeonTime(eventTime)
		return fmt.Errorf("%s Player [%d] is dead", util.FormatDuration(eventTime), playerID)
	}

	g.players[playerID].ChangeHP(g.players[playerID].GetHP() - hp)
	return fmt.Errorf("%s Player [%d] recieved [%d] of damage", util.FormatDuration(eventTime), playerID, hp)
}

func (g *Game) RestoreHP(playerID int, eventTime time.Duration, hp int) error {
	if err := g.generalCheck(playerID, eventTime); err != nil {
		return err
	}

	g.players[playerID].ChangeHP(min(g.players[playerID].GetHP()+hp, 100))
	return fmt.Errorf("%s Player [%d] restored [%d] of hp", util.FormatDuration(eventTime), playerID, hp)
}

func (g *Game) PlayerDisconnect(playerID int, eventTime time.Duration, reason int) error {
	if v := g.isRegister(playerID); !v {
		return fmt.Errorf("player %d is not registered", playerID)
	}

	g.players[playerID].ChangeState("DISQUAL")
	return fmt.Errorf("%s Player [%d] cant continue game because: %d", util.FormatDuration(eventTime), playerID, reason)
}

func (g *Game) KillMonster(playerID int, eventTime time.Duration) error {
	if err := g.generalCheck(playerID, eventTime); err != nil {
		return err
	}

	if g.players[playerID].GetCurrentMonstersOnFloor() <= 0 {
		return fmt.Errorf("%s Player [%d] makes impossible move [3]", util.FormatDuration(eventTime), playerID)
	}

	g.players[playerID].KillMonster()

	if g.players[playerID].GetCurrentMonstersOnFloor() == 0 {
		g.players[playerID].SetFloorClearTime(g.players[playerID].GetFloor(), eventTime)
	}

	return fmt.Errorf("%s Player [%d] killed the monster", util.FormatDuration(eventTime), playerID)
}

func (g *Game) GameLogs() string {
	var res strings.Builder
	res.WriteString("Final report:\n\n")

	for _, v := range g.players {
		id := strconv.Itoa(v.ID)

		state := v.GetState()

		timeInDungeon := v.GetTimeInDungeon()

		average := v.GetAverage()

		bossKilltime := v.GetBossKillTime()

		hp := strconv.Itoa(v.GetHP())

		res.WriteString(fmt.Sprintf("%s  %s [%s %s %s] HP:%s\n", state, id, timeInDungeon, average, bossKilltime, hp))
	}

	return res.String()
}

func (g *Game) generalCheck(playerID int, eventTime time.Duration) error {
	if !g.isRegister(playerID) {
		return fmt.Errorf("%s player [%d] is not registered", util.FormatDuration(eventTime), playerID)
	}

	if err := g.isClosed(eventTime); err != nil {
		return err
	}

	if g.players[playerID].GetState() == "DISQUAL" {
		return fmt.Errorf("%s Player [%d] is disqualified", util.FormatDuration(eventTime), playerID)
	}

	return nil
}

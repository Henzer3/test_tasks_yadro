package game

import (
	"time"

	"test.task.telekom.game/internal/util"
)

type Player struct {
	ID int

	InDungeon bool
	Finished  bool

	State string // SUCCESS, FAIL, DISQUAL

	HP int

	CurrentFloor  int
	MonstersAlive []int

	EnterDungeonTime time.Duration
	EndTime          time.Duration

	FloorStartTime  []time.Duration
	FloorClearTimes []time.Duration

	BossEnterTime time.Duration
	BossKillTime  time.Duration

	BossKilled bool
}

func NewPlayer(id int, floors int, monsters int) *Player {
	monstersAlive := make([]int, 0, floors)
	for range floors {
		monstersAlive = append(monstersAlive, monsters)
	}
	return &Player{
		ID:              id,
		HP:              100,
		MonstersAlive:   monstersAlive,
		FloorStartTime:  make([]time.Duration, floors),
		FloorClearTimes: make([]time.Duration, floors),
	}
}

func (p *Player) SetEnterDungeonTime(t time.Duration) {
	p.EnterDungeonTime = t
}

func (p *Player) SetLeftDungeonTime(t time.Duration) {
	p.EndTime = t
}

func (p *Player) SetEnterBossRoom(t time.Duration) {
	p.BossEnterTime = t
}

func (p *Player) SetBossKillTime(t time.Duration) {
	p.BossKillTime = t
}

func (p *Player) GetState() string {
	if p.State != "" {
		return p.State
	}

	if !p.BossKilled {
		return "FAIL"
	}

	for _, v := range p.MonstersAlive {
		if v != 0 {
			return "FAIL"
		}
	}

	return "SUCCESS"
}

func (p *Player) SetFloorStarTime(floor int, t time.Duration) {
	if p.FloorStartTime[floor] == 0 {
		p.FloorStartTime[floor] = t
	}
}

func (p *Player) SetFloorClearTime(floor int, t time.Duration) {
	if p.FloorClearTimes[floor] == 0 {
		p.FloorClearTimes[floor] = t
	}
}

func (p *Player) GetTimeInDungeon() string {
	return util.FormatDuration(p.EndTime - p.EnterDungeonTime)
}

func (p *Player) GetAverage() string {
	var sum time.Duration

	for i := range len(p.FloorStartTime) {
		if p.FloorClearTimes[i] == 0 {
			return "none"
		}
		sum += p.FloorClearTimes[i] - p.FloorStartTime[i]
	}

	return util.FormatDuration(sum / time.Duration(len(p.FloorStartTime)))
}

func (p *Player) GetBossKillTime() string {
	if p.BossKillTime == 0 {
		return "none"
	}
	return util.FormatDuration(p.BossKillTime - p.BossEnterTime)
}

func (p *Player) KillMonster() {
	p.MonstersAlive[p.CurrentFloor]--
}

func (p *Player) GetCurrentMonstersOnFloor() int {
	return p.MonstersAlive[p.CurrentFloor]
}

func (p *Player) ChangeState(value string) {
	p.State = value
}

func (p *Player) GetHP() int {
	return p.HP
}

func (p *Player) ChangeHP(value int) {
	p.HP = value
}

func (p *Player) GetBossKilled() bool {
	return p.BossKilled
}

func (p *Player) ChangeBossKilled(value bool) {
	p.BossKilled = value
}

func (p *Player) GetFloor() int {
	return p.CurrentFloor
}

func (p *Player) ChangeFloor(value int) {
	p.CurrentFloor = value
}

func (p *Player) GetInDungeon() bool {
	return p.InDungeon
}

func (p *Player) ChangeInDungeon(value bool) {
	p.InDungeon = value
}

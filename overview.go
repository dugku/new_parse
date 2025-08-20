package main

import (
	"time"

	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
)

type parsingState struct {
	round        int
	RoundOngoing bool
	TeamASide    common.Team
	TeamBSide    common.Team
	WarmupKills  []events.Kill
}

type MatchInfo struct {
	Map          string
	TeamA        string
	TeamB        string
	Winner       string
	Loser        string
	RoundsPlayed int
	Date         string
	MatchId      string
	TeamOneId    string
	TeamTwoId    string
	Rounds       []RoundInformation
	Players      map[uint64]playerStat
}

type RoundInformation struct {
	RoundNum         int
	TeamASide        int
	TeamBSide        int
	CTEquipmentValue int    // at freeze time end
	TEquipmentValue  int    // at freeze time end
	TypeofBuyCT      string // at freeze time end
	TypeofBuyT       string // at freeze time end
	ScoreCT          int
	ScoreT           int
	FirstKillCount   int      `json:"-"` //need to think about trade kills later
	SurvivorsA       []string `json:"-"`
	SurvivorsB       []string `json:"-"`
	SurvivorsAInt    int      //number of Survivors at the end of the round
	SurvivorsBInt    int      //number of Survivors at the end of the round
	BombPlanted      bool     // was bomb planted at some point in the round?
	PlayerPlanted    string
	RoundEndedReason string
	SideWon          string
	RoundClanWinner  string
	RoundClanLoser   string
	KillARound       map[int]RoundKill
	BombPlantedSite  string
	Duration         time.Duration
	PlayersAliveA    map[uint64]bool `json:"-"`
	PlayersAliveB    map[uint64]bool `json:"-"`
	OneVX            bool            `json:"-"`
	OneVXCount       int             `json:"-"`
}

type RoundKill struct {
	TimeOfKill       time.Duration
	Killer           string
	IsOpening        bool
	KillerId         uint64
	VictId           uint64
	Victim           string
	Assistor         string
	KillerTeamString string
	VictimTeamString string
	VictFlashDur     float32
	VictDmgTaken     int
	AttDmgTaken      int
	IsHeadshot       bool
	IsFlashed        bool
	Dist             float64
	KillerWeapon     int
	KillerTeam       int
	VictTeam         int
	AttackerHealth   int
	AttackerX        float64
	AttackerY        float64
	VictX            float64
	VictY            float64
	KillerClan       string
	VictClan         string
}

type playerStat struct {
	ImpactPerRnd       float64
	UserName           string
	SteamID            uint64
	Kills              int
	Deaths             int
	Assists            int
	HS                 int
	HeadPercent        float64
	ADR                float64
	KAST               float64
	KDRatio            float64
	Firstkill          int
	FirstDeath         int
	FKDiff             int
	Round2k            int
	Round3k            int
	Round4k            int
	Round5k            int
	Totaldmg           int
	TradeKills         int
	TradeDeath         int
	CTkills            int
	Tkills             int
	EffectiveFlashes   int
	AvgflshDuration    float64
	WeaponKill         map[int]int
	WeaponKillHS       map[int]int
	AvgDist            float64 `json:"-"`
	TotalDist          float64 `json:"-"`
	FlashesThrown      int     `json:"-"`
	ClanName           string
	TotalUtilDmg       int
	AvgKillsRnd        float64
	AvgDeathsRnd       float64
	AvgAssistsRnd      float64
	AvgNadeDmg         float64
	AvgInferDmg        float64
	RoundSurvived      int
	RoundTraded        int
	RoundContributed   []int `json:"-"`
	InfernoDmg         int
	NadeDmg            int
	OpeningKillsSuc    int
	OpeningKillFail    int
	TotalOpening       int
	OpeningPercent     float64
	OpeningAttpPrecent float64
	OpeningRoundsWon   int
	OpeningWinPercent  float64
	OneVsOne           int
	OneVsTwo           int
	OneVsThree         int
	OneVsFour          int
	OneVsFive          int
}

type BombStates struct {
	Tick       int
	secs       time.Duration
	ClockTime  string
	X, Y, Z    float64
	SteamID    uint64
	PlayerName string
	BombSite   string
}

type PlayerState struct {
}

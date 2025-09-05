package main

import (
	"fmt"
	"log"
	"time"

	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
)

var (
	player_map = make(map[string]int)
)

func player_get(c *common.Player) playerStat {
	return playerStat{
		UserName:        c.Name,
		SteamID:         c.SteamID64,
		Kills:           0,
		Deaths:          0,
		Assists:         0,
		HS:              0,
		HeadPercent:     0,
		ADR:             0,
		KAST:            0,
		KDRatio:         0,
		Firstkill:       0,
		FirstDeath:      0,
		FKDiff:          0,
		Round2k:         0,
		Round3k:         0,
		Round4k:         0,
		Round5k:         0,
		Totaldmg:        0,
		TradeKills:      0,
		TradeDeath:      0,
		CTkills:         0,
		Tkills:          0,
		AvgflshDuration: 0,
		WeaponKill:      all_weapons(),
		WeaponKillHS:    all_weapons(),
		ClanName:        c.TeamState.ClanName(),
		TotalUtilDmg:    0,
	}
}

func all_weapons() map[int]int { return make(map[int]int) }

func player_getter(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	p.RegisterEventHandler(func(e events.RoundEnd) {
		gs := p.GameState()

		teamAObj := gs.Team(s.TeamASide)
		teamBObj := gs.Team(s.TeamBSide)

		if teamAObj == nil || teamBObj == nil {
			log.Println("One or both team objects are nil; skipping player getter")
			return
		}

		TeamA := teamAObj.Members()
		TeamB := teamBObj.Members()

		if TeamA == nil || TeamB == nil {
			log.Println("One or both team member lists are nil; skipping player getter")
			return
		}

		stat_setter(TeamA, m, gs)
		stat_setter(TeamB, m, gs)
	})
}
func stat_setter(c []*common.Player, m *MatchInfo, gs demoinfocs.GameState) {
	for i := range c {
		steam_id := c[i].SteamID64
		player, exists := m.Players[steam_id]
		if !exists {
			continue
		}

		player.Kills = c[i].Kills()
		player.Deaths = c[i].Deaths()
		player.Assists = c[i].Assists()
		player.Totaldmg = c[i].TotalDamage()
		player.ADR = calc_adr(gs, player.Totaldmg)
		player.KDRatio = calc_kd(player.Kills, player.Deaths)
		player.HeadPercent = calc_hs_per(player.HS, player.Kills)
		player.AvgKillsRnd = calc_avg_per_round(player.Kills, gs.TotalRoundsPlayed())
		player.AvgDeathsRnd = calc_avg_per_round(player.Deaths, gs.TotalRoundsPlayed())
		player.AvgAssistsRnd = calc_avg_per_round(player.Assists, gs.TotalRoundsPlayed())
		player.ImpactPerRnd = calc_impatct(player.AvgKillsRnd, player.AvgAssistsRnd)
		player.KAST = calc_kast(len(player.RoundContributed), gs.TotalRoundsPlayed()) * 100
		player.AvgNadeDmg = calc_ndmg(player.NadeDmg, gs.TotalRoundsPlayed())
		player.AvgInferDmg = calc_infer_dmg(player.InfernoDmg, gs.TotalRoundsPlayed())
		player.TotalOpening = player.OpeningKillsSuc + player.OpeningKillFail
		player.OpeningPercent = calc_open_percent(player.TotalOpening, gs.TotalRoundsPlayed())
		player.OpeningAttpPrecent = calc_open_percent(player.OpeningKillsSuc, player.TotalOpening)
		player.OpeningWinPercent = calc_open_percent(player.OpeningRoundsWon, player.TotalOpening)

		//fmt.Println(player.OpeningPercent)

		player_name := c[i].Name

		multi_check := c[i].Kills() - player_map[player_name]

		switch {
		case multi_check == 2:
			player.Round2k++
		case multi_check == 3:
			player.Round3k++
		case multi_check == 4:
			player.Round4k++
		case multi_check == 5:
			player.Round5k++
		}

		m.Players[steam_id] = player
	}
}

func get_pres_round_kill(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	p.RegisterEventHandler(func(e events.RoundFreezetimeEnd) {
		gs := p.GameState()

		teamA := gs.Team(s.TeamASide)
		teamB := gs.Team(s.TeamBSide)

		if teamA != nil {
			print_this(teamA.Members())
		} else {
			log.Println("Warning: TeamA is nil; cannot retrieve members.")
		}

		if teamB != nil {
			print_this(teamB.Members())
		} else {
			log.Println("Warning: TeamB is nil; cannot retrieve members.")
		}
	})
}

// still don't know how this works. even after rewriting this twice
func print_this(c []*common.Player) {
	for _, v := range c {
		player_map[v.Name] = v.Kills()
	}
}

func kill_handler(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	p.RegisterEventHandler(func(e events.Kill) {
		gs := p.GameState()
		round_info := &m.Rounds[s.round-1]
		if e.Killer == nil || e.Victim == nil {
			return
		}

		if p.GameState().IsWarmupPeriod() {
			s.WarmupKills = append(s.WarmupKills, e)
			return
		}

		if e.IsHeadshot && e.Weapon != nil {
			add_headshot(e.Killer, e.Weapon.Type, m)
		}

		var assistor string
		if e.Assister != nil {
			assistor = e.Assister.Name
			//assistor_id = e.Assister.SteamID64
		}

		open_kill := false

		if s.round > 0 && s.round <= len(m.Rounds) {
			if m.Rounds[s.round-1].KillARound == nil {
				m.Rounds[s.round-1].KillARound = make(map[int]RoundKill)
			}

			count := len(m.Rounds[s.round-1].KillARound) + 1

			if count == 1 {
				open_kill = true
			} else {
				open_kill = false
			}

			if _, exists := m.Rounds[s.round-1].KillARound[count]; exists {
				return
			} else {
				delta := p.CurrentTime() - round_info.TimeRoundStart
				m.Rounds[s.round-1].KillARound[count] = RoundKill{
					Tick:              p.GameState().IngameTick(),
					TimeOfKill:        int64(delta / time.Second),
					AttackerName:      e.Killer.Name,
					AttackerId:        e.Killer.SteamID64,
					AttackerHealth:    e.Killer.Health(),
					AttackerTeam:      int(e.Killer.TeamState.Team()),
					AttackerX:         e.Killer.Position().X,
					AttackerY:         e.Killer.Position().Y,
					AttackerClan:      gs.Team(e.Killer.TeamState.Team()).ClanName(),
					AttackerViewX:     e.Killer.ViewDirectionX(),
					AttackerViewY:     e.Killer.ViewDirectionY(),
					AttackerIsFlashed: e.Killer.IsBlinded(),
					Assistor:          assistor,
					VictimName:        e.Victim.Name,
					VictimID:          e.Victim.SteamID64,
					VictimFlashed:     e.Victim.IsBlinded(),
					VictFlashDur:      e.Victim.FlashDuration,
					VictTeam:          int(e.Victim.TeamState.Team()),
					VictimX:           e.Victim.Position().X,
					VictimY:           e.Victim.Position().Y,
					VictimViewX:       e.Victim.ViewDirectionX(),
					VictimViewY:       e.Victim.ViewDirectionY(),
					VictClan:          gs.Team(e.Victim.TeamState.Team()).ClanName(),
					IsHeadshot:        e.IsHeadshot,
					IsOpening:         open_kill,
					IsWallbang:        e.IsWallBang(),
					IsNoscope:         e.NoScope,
					IsThroughSmoke:    e.ThroughSmoke,
					IsAssistFlash:     e.AssistedFlash,
					Weapon:            int(e.Killer.ActiveWeapon().Type),
				}
				//count++
			}

			if m.Rounds[s.round-1].FirstKillCount == 0 {
				m.Rounds[s.round-1].FirstKillCount = 1

				player_killer := e.Killer.SteamID64
				player_vict := e.Victim.SteamID64

				player_stat_kill, exists := m.Players[player_killer]
				if !exists {
					return
				}

				player_stat_kill.Firstkill++
				m.Players[player_killer] = player_stat_kill

				player_stat_vict, exists2 := m.Players[player_vict]
				if !exists2 {
					return
				}

				player_stat_vict.FirstDeath++
				m.Players[player_vict] = player_stat_vict

			}

			if s.round > 0 && s.round <= len(m.Rounds) {
				kill_id := e.Killer.SteamID64

				player, exists := m.Players[kill_id]
				if !exists {
					return
				}
				player.RoundContributed = append(player.RoundContributed, s.round)
				m.Players[kill_id] = player

				if e.Assister != nil {
					if playerAssist, exists2 := m.Players[e.Assister.SteamID64]; exists2 {
						playerAssist.RoundContributed = append(playerAssist.RoundContributed, s.round)
						m.Players[e.Assister.SteamID64] = playerAssist
					}
				}

			}

			if e.Weapon != nil {
				update_weapon_kill(e.Killer, e.Weapon.Type, m)
			}

			//1vX logic this is going to be fun
			if s.round > 0 && s.round <= len(m.Rounds) {
				round_info := &m.Rounds[s.round-1]

				OneVsX := map[int]string{
					1: "OneVsOne",
					2: "OneVsTwo",
					3: "OneVsThree",
					4: "OneVsFour",
					5: "OneVsFive",
				}

				//checksides?
				if e.Victim.Team == s.TeamASide {
					delete(round_info.PlayersAliveA, e.Victim.SteamID64)
				}
				if e.Victim.Team == s.TeamBSide {
					delete(round_info.PlayersAliveB, e.Victim.SteamID64)
				}

				if !round_info.OneVX {

					if len(round_info.PlayersAliveA) == 1 && len(round_info.PlayersAliveB) >= 1 {
						if _, exists := OneVsX[len(round_info.PlayersAliveB)]; exists {
							round_info.OneVXCount = len(round_info.PlayersAliveB)

							round_info.OneVX = true
						}
					} else if len(round_info.PlayersAliveB) == 1 && len(round_info.PlayersAliveA) >= 1 {
						if _, exists := OneVsX[len(round_info.PlayersAliveA)]; exists {
							round_info.OneVXCount = len(round_info.PlayersAliveA)
							round_info.OneVX = true
						}
					}
				}

				if len(round_info.PlayersAliveA) == 1 && len(round_info.PlayersAliveB) == 0 {
					killer_id := e.Killer.SteamID64
					player, exists := m.Players[killer_id]
					if !exists {
						return
					}

					if round_info.OneVXCount > 0 && round_info.OneVXCount <= 5 {
						clutchStats := []*int{
							&player.OneVsOne, &player.OneVsTwo, &player.OneVsThree, &player.OneVsFour, &player.OneVsFive,
						}
						*clutchStats[round_info.OneVXCount-1]++
					}

					m.Players[killer_id] = player
				}

				if len(round_info.PlayersAliveA) == 0 && len(round_info.PlayersAliveB) == 1 {
					killer_id := e.Killer.SteamID64

					player, exists := m.Players[killer_id]
					if !exists {
						return
					}

					if round_info.OneVXCount > 0 && round_info.OneVXCount <= 5 {
						clutchStats := []*int{
							&player.OneVsOne, &player.OneVsTwo, &player.OneVsThree, &player.OneVsFour, &player.OneVsFive,
						}
						*clutchStats[round_info.OneVXCount-1]++
					}

					m.Players[killer_id] = player
				}

			}

		}
	})
}
func add_headshot(c *common.Player, w common.EquipmentType, m *MatchInfo) {
	player_id := c.SteamID64
	player, exists := m.Players[player_id]
	if !exists {
		return
	}
	player.HS++
	fmt.Println(player.HS, player.UserName, player.Kills)
	player.WeaponKillHS[int(w)]++
	m.Players[player_id] = player
}

func update_weapon_kill(c *common.Player, weapon_type common.EquipmentType, m *MatchInfo) {
	if weapon_type == 407 {
		return
	}

	player_id := c.SteamID64
	player, exists := m.Players[player_id]
	if !exists {
		return
	}

	player.WeaponKill[int(weapon_type)]++
	m.Players[player_id] = player
}

func players_hurt(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	p.RegisterEventHandler(func(e events.PlayerHurt) {
		// Safety net: never let a single bad event kill the whole parse
		defer func() {
			if r := recover(); r != nil {
				log.Printf("recovered in PlayerHurt handler: %v", r)
			}
		}()

		gs := p.GameState()

		// Skip warmup completely to avoid contaminating stats
		if gs.IsWarmupPeriod() {
			return
		}

		// Victim is required for a meaningful PlayerHurt
		if e.Player == nil {
			log.Printf("PlayerHurt with nil victim @tick=%d; skipping", gs.IngameTick())
			return
		}

		// Validate round bounds BEFORE indexing
		if s.round <= 0 || s.round > len(m.Rounds) {
			return
		}
		round := &m.Rounds[s.round-1]

		var dmg PlayerDamages
		dmg.Tick = gs.IngameTick()
		dmg.Secs = int64((p.CurrentTime() - round.TimeRoundStart) / time.Second)

		if a := e.Attacker; a != nil {
			dmg.AttackerId = a.SteamID64
			dmg.AttackerName = a.Name
			dmg.AttackerTeam = a.ClanTag()
			dmg.AttackerSide = int(a.Team)
			apos := a.Position()
			dmg.AttackerPosX = apos.X
			dmg.AttackerPosY = apos.Y
			dmg.AttackerViewX = a.ViewDirectionX()
			dmg.AttackerViewY = a.ViewDirectionY()
			dmg.AttckerHealth = a.Health()
		} else {
			// World / env damage or unknown attacker
			dmg.AttackerId = 0
			dmg.AttackerName = "world"
			dmg.AttackerTeam = ""
			dmg.AttackerSide = int(common.TeamUnassigned)
			dmg.AttckerHealth = 0
		}

		v := e.Player
		dmg.VictimID = v.SteamID64
		dmg.VictimName = v.Name
		dmg.VictimTeam = v.ClanTag()
		dmg.VictimSide = int(v.Team) // avoid TeamState
		vpos := v.Position()
		dmg.VictimPosX = vpos.X
		dmg.VictimPosY = vpos.Y
		dmg.VictimViewX = v.ViewDirectionX()
		dmg.VictimViewY = v.ViewDirectionY()
		dmg.VictimHealth = v.Health()

		if e.Weapon != nil {
			dmg.Weapon = int(e.Weapon.Type)
			if e.WeaponString != "" {
				dmg.WeaponClass = e.WeaponString
			} else {
				dmg.WeaponClass = e.Weapon.String()
			}
		} else {
			dmg.Weapon = 0
			if e.WeaponString != "" {
				dmg.WeaponClass = e.WeaponString
			} else {
				dmg.WeaponClass = "env"
			}
		}

		dmg.HPDmg = e.HealthDamage
		dmg.HPDmgTaken = e.HealthDamageTaken
		dmg.ArmorDmg = e.ArmorDamage
		dmg.ArmourDmgTaken = e.ArmorDamageTaken
		dmg.HitGroup = int(e.HitGroup)

		round.Damages = append(round.Damages, dmg)

		if e.Attacker == nil || e.Weapon == nil {
			log.Printf("PlayerHurt @tick=%d attacker_nil=%t weapon_nil=%t dmg=%d hit=%d",
				gs.IngameTick(), e.Attacker == nil, e.Weapon == nil, int(e.HealthDamage), e.HitGroup)
		}
	})
}

func player_fired(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	p.RegisterEventHandler(func(e events.WeaponFire) {
		// Safety net: never let one bad event kill the parse
		defer func() {
			if r := recover(); r != nil {
				log.Printf("recovered in WeaponFire handler: %v", r)
			}
		}()

		gs := p.GameState()

		// Shooter required for your stats
		if e.Shooter == nil {
			return
		}

		// Ignore warmup entirely
		if gs.IsWarmupPeriod() {
			return
		}

		// Validate round index BEFORE indexing
		if s.round <= 0 || s.round > len(m.Rounds) {
			return
		}
		round := &m.Rounds[s.round-1]

		var f PlayerFired
		f.Tick = gs.IngameTick()
		f.Secs = int64((p.CurrentTime() - round.TimeRoundStart) / time.Second)

		sh := e.Shooter
		f.PlayerSteamID = sh.SteamID64

		f.PlayerName = sh.Name

		f.PlayerSide = int(sh.Team)

		spos := sh.Position()
		f.PlayerPosX = spos.X
		f.PlayerPosY = spos.Y
		f.PlayerViewX = sh.ViewDirectionX()
		f.PlayerViewY = sh.ViewDirectionY()

		var wType int
		var ammoMag, ammoRes int
		if e.Weapon != nil {
			wType = int(e.Weapon.Type)
			ammoMag = e.Weapon.AmmoInMagazine()
			ammoRes = e.Weapon.AmmoReserve()
		} else if aw := sh.ActiveWeapon(); aw != nil {
			wType = int(aw.Type)
			ammoMag = aw.AmmoInMagazine()
			ammoRes = aw.AmmoReserve()
		} else {
			// Unknown / env / timing race
			wType = 0
			ammoMag = 0
			ammoRes = 0
			log.Printf("WeaponFire @tick=%d shooter=%s(%d) with nil weapon and no active weapon",
				gs.IngameTick(), sh.Name, sh.SteamID64)
		}
		f.Weapon = wType
		f.AmmoInMag = ammoMag
		f.AmmoInReserve = ammoRes

		round.ShotsFired = append(round.ShotsFired, f)
	})
}

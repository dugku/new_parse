package main

import (
	"log"

	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
)

func state_control(p demoinfocs.Parser, pstate *parsingState, m *MatchInfo) {
	p.RegisterEventHandler(func(e events.RoundStart) {
		gs := p.GameState()

		pstate.round++

		pstate.RoundOngoing = true
		round := RoundInformation{
			BombStuff:      make([]BombStates, 0),
			TimeRoundStart: p.CurrentTime(),
		}

		teamA := gs.Team(pstate.TeamASide)
		teamB := gs.Team(pstate.TeamBSide)
		if teamA == nil || teamB == nil {
			log.Printf("Nil team state encountered: teamA=%v, teamB=%v", teamA, teamB)
			return // Skip further processing to avoid panic.
		}

		m.Rounds = append(m.Rounds, round)

		m.RoundsPlayed = pstate.round

		alive_players_logic(pstate, m, round, gs)
	})

	p.RegisterEventHandler(func(e events.RoundEnd) {
		pstate.RoundOngoing = false

	})
}

func alive_players_logic(s *parsingState, m *MatchInfo, round_ah RoundInformation, gs demoinfocs.GameState) {

	if s.round <= 0 || s.round > len(m.Rounds) {
		return // Prevent out-of-bounds access
	}

	round := &m.Rounds[s.round-1] // Get pointer to the actual round data

	round.OneVX = false

	// Reset alive players map
	round.PlayersAliveA = make(map[uint64]bool)
	round.PlayersAliveB = make(map[uint64]bool)

	// Populate alive players
	for _, player := range gs.Participants().Playing() {
		if player.TeamState == nil {
			continue // Skip invalid players but continue the loop
		}
		if player.Team == s.TeamASide {
			round.PlayersAliveA[player.SteamID64] = true
		} else if player.Team == s.TeamBSide {
			round.PlayersAliveB[player.SteamID64] = true
		}
	}
}

func match_started(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	p.RegisterEventHandler(func(e events.MatchStartedChanged) {
		gs := p.GameState()

		players := gs.Participants().Playing()

		//This creates the player so that we don't do it everytime.
		for _, player := range players {
			steamId := player.SteamID64

			if m.Players == nil {
				m.Players = make(map[uint64]playerStat)
			}

			if _, exists := m.Players[uint64(steamId)]; exists {
				return
			} else {
				m.Players[uint64(steamId)] = player_get(player)
			}
		}

		team_a := gs.TeamCounterTerrorists().ClanName()
		team_b := gs.TeamTerrorists().ClanName()

		m.TeamA = team_a
		m.TeamB = team_b

		//need to switch when half
		s.TeamASide = common.TeamCounterTerrorists
		s.TeamBSide = common.TeamTerrorists

		//map name
		m.Map = p.Header().MapName
	})
}

func team_switch(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	p.RegisterEventHandler(func(e events.TeamSideSwitch) {
		TempSide := s.TeamASide
		s.TeamASide = s.TeamBSide
		s.TeamBSide = TempSide

	})
}

func score_updater(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	p.RegisterEventHandler(func(e events.ScoreUpdated) {
		gs := p.GameState()
		ctScr := gs.TeamCounterTerrorists().Score()
		tScr := gs.TeamTerrorists().Score()

		if s.round > 0 && s.round <= len(m.Rounds) {
			m.Rounds[s.round-1].RoundNum = s.round
			m.Rounds[s.round-1].ScoreCT = ctScr
			m.Rounds[s.round-1].ScoreT = tScr
			m.Rounds[s.round-1].TeamASide = int(s.TeamASide)
			m.Rounds[s.round-1].TeamBSide = int(s.TeamBSide)
		}
	})
}

func check_side(team common.Team, p demoinfocs.Parser) (TeamScore int) {
	gs := p.GameState()

	if team == common.TeamCounterTerrorists {
		return gs.TeamCounterTerrorists().Score()
	}

	if team == common.TeamTerrorists {
		return gs.TeamTerrorists().Score()
	}

	return 0
}

// Going to reassessing Round Econ, need to thing about it a little
// but for now this is going to have to work I guess.
func round_econ_logic(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	p.RegisterEventHandler(func(e events.RoundFreezetimeEnd) {
		gs := p.GameState()

		round_info := &m.Rounds[s.round-1]

		ct := gs.TeamCounterTerrorists()
		t := gs.TeamTerrorists()

		// Check if the team states are nil
		if ct == nil || t == nil {
			// Optionally log or handle the nil case
			log.Println("One or both team states are nil; skipping econ calculation.")
			return
		}

		ctval := ct.FreezeTimeEndEquipmentValue()
		tval := t.FreezeTimeEndEquipmentValue()

		ctBuy := assess_econ(ctval)
		tBuy := assess_econ(tval)

		round_info.CTEquipmentValue = ctval
		round_info.TEquipmentValue = tval
		round_info.TypeofBuyCT = ctBuy
		round_info.TypeofBuyT = tBuy

	})
}

func assess_econ(team_econ int) string {
	FullBuy := 20000
	HalfBuy := 10000
	SemiEco := 5000

	switch {
	case team_econ >= FullBuy:
		return "Full Buy"
	case team_econ >= HalfBuy:
		return "Half Buy"
	case team_econ >= SemiEco:
		return "Force Buy"
	default:
		return "Eco"
	}
}

// start other stuff
func round_end_logic(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	p.RegisterEventHandler(func(e events.RoundEnd) {
		gs := p.GameState()

		WinnerMap := map[int]string{
			2: "t",
			3: "ct",
		}

		round_info := &m.Rounds[s.round-1]

		round_info.TimeRoundEnd = p.CurrentTime()

		players_a := gs.TeamCounterTerrorists().Members()
		players_b := gs.TeamTerrorists().Members()

		round_Enwinner(e, m, s)
		round_contrib(e, m, s, gs)
		is_tade(m, s)

		remove_dups(p, m, players_a)
		remove_dups(p, m, players_b)
		open_kill(m, s)
		open_kill_win(m, s, WinnerMap)
	})
}

func is_tade(m *MatchInfo, s *parsingState) {
	if s.round > 0 && s.round <= len(m.Rounds) {
		round_info := &m.Rounds[s.round-1]

		for key, _ := range round_info.KillARound {
			if key+1 < len(round_info.KillARound) {
				next_val := round_info.KillARound[key+1]

				if round_info.KillARound[key].AttackerId == next_val.VictimID {
					diffS := next_val.TimeOfKill - round_info.KillARound[key].TimeOfKill
					if diffS >= 0 && diffS < 5 {
						trade_killer_id := next_val.AttackerId
						trade_vict_id := next_val.VictimID

						player_kill, exists := m.Players[trade_killer_id]
						if !exists {
							return
						}
						player_kill.TradeKills++
						player_kill.RoundContributed = append(player_kill.RoundContributed, s.round)
						m.Players[trade_killer_id] = player_kill

						player_vict, exists := m.Players[trade_vict_id]
						if !exists {
							return
						}

						player_vict.RoundTraded++
						m.Players[trade_vict_id] = player_vict
					}
				}
			}
		}

		for key, _ := range round_info.KillARound {
			killer_id := round_info.KillARound[key].AttackerId
			player, exists := m.Players[killer_id]
			if !exists {
				return
			}

			if round_info.KillARound[key].AttackerName == round_info.KillARound[key].AttackerName {
				continue
			}
			if round_info.KillARound[key].AttackerTeam == 3 {
				player.CTkills++
			}

			if round_info.KillARound[key].AttackerTeam == 2 {
				player.Tkills++
			}
			m.Players[killer_id] = player
		}
	}
}

func round_contrib(e events.RoundEnd, m *MatchInfo, s *parsingState, gs demoinfocs.GameState) {
	players_a := gs.TeamCounterTerrorists().Members()
	players_b := gs.TeamTerrorists().Members()

	if s.round > 0 && s.round <= len(m.Rounds) {
		round_info := &m.Rounds[s.round-1]

		for _, v := range players_a {
			if v.IsAlive() {
				player_id := v.SteamID64

				round_info.SurvivorsA = append(round_info.SurvivorsA, v.String())

				player, exists := m.Players[player_id]
				if !exists {
					return
				}

				player.RoundSurvived++
				player.RoundContributed = append(player.RoundContributed, s.round)
				m.Players[player_id] = player
			}
		}

		for _, v := range players_b {
			if v.IsAlive() {
				player_id := v.SteamID64

				round_info.SurvivorsB = append(round_info.SurvivorsB, v.String())
				player, exists := m.Players[player_id]

				if !exists {
					return
				}

				player.RoundSurvived++
				player.RoundContributed = append(player.RoundContributed, s.round)
				m.Players[player_id] = player
			}
		}

		round_info.SurvivorsAInt = len(round_info.SurvivorsA)
		round_info.SurvivorsBInt = len(round_info.SurvivorsB)
	}
}

func round_Enwinner(e events.RoundEnd, m *MatchInfo, s *parsingState) {
	ReasonsMap := map[int]string{
		1:  "TargetBombed",
		7:  "BombDefused",
		8:  "CTWin",
		9:  "TWin",
		12: "TargetSaved",
	}

	WinnerMap := map[int]string{
		2: "t",
		3: "ct",
	}

	reason := e.Reason
	side_won := e.WinnerState
	won := e.Winner
	side_lost := e.LoserState

	if s.round > 0 && s.round <= len(m.Rounds) {
		round_info := &m.Rounds[s.round-1]

		round_info.RoundEndedReason = ReasonsMap[int(reason)]
		round_info.SideWon = WinnerMap[int(won)]

		if side_won != nil {
			round_info.RoundClanWinner = side_won.ClanName()
		} else {
			round_info.RoundClanWinner = ""
		}

		if side_lost != nil {
			round_info.RoundClanLoser = side_lost.ClanName()
		} else {
			round_info.RoundClanLoser = ""
		}
	}
}

func open_kill(m *MatchInfo, s *parsingState) {
	if s.round > 0 && s.round <= len(m.Rounds) {
		round_info := &m.Rounds[s.round-1]

		if round_info.KillARound[1].IsOpening {
			killer_id := round_info.KillARound[1].AttackerId
			player, exists := m.Players[killer_id]
			if !exists {
				return
			}

			player.OpeningKillsSuc++

			m.Players[killer_id] = player
		}

		if round_info.KillARound[1].IsOpening {
			victim_id := round_info.KillARound[1].VictimID
			player, exists := m.Players[victim_id]
			if !exists {
				return
			}
			player.OpeningKillFail++
			m.Players[victim_id] = player
		}
	}
}

func open_kill_win(m *MatchInfo, s *parsingState, win map[int]string) {

	if s.round > 0 && s.round <= len(m.Rounds) {
		round_info := &m.Rounds[s.round-1]

		if round_info.KillARound[1].IsOpening {
			killer_T := round_info.KillARound[1].AttackerTeam
			side_won := round_info.SideWon

			val := win[killer_T]

			if val == side_won {
				kill_id := round_info.KillARound[1].AttackerId
				player, exists := m.Players[kill_id]
				if !exists {
					return
				}

				player.OpeningRoundsWon++

				m.Players[kill_id] = player
			}
		}
	}
}

func bom_planted(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	p.RegisterEventHandler(func(e events.BombPlanted) {
		gs := p.GameState()

		if m == nil || s == nil {
			log.Printf("bomb_planted: nil state m=%v s=%v", m, s)
			return
		}
		if e.Player == nil {
			log.Printf("bomb_planted: nil planter at tick %d", p.GameState().IngameTick())
			return
		}

		if len(m.Rounds) == 0 {
			log.Printf("bomb_planted: no rounds exist at tick %d; ignoring", p.GameState().IngameTick())
			return
		}

		// Validate/repair s.round (1-based)
		if s.round < 1 || s.round > len(m.Rounds) {
			log.Printf("bomb_planted: s.round=%d out of [1,%d]; recovering to last",
				s.round, len(m.Rounds))
			s.round = len(m.Rounds)
		}

		idx := s.round - 1
		round_info := &m.Rounds[idx]

		sites := map[int]rune{
			65: 'A',
			66: 'B',
		}

		var b BombStates

		round_info.BombPlanted = true
		round_info.PlayerPlanted = e.Player.Name
		round_info.BombPlantedSite = string(e.Site)

		b.Tick = gs.IngameTick()
		diff := p.CurrentTime() - round_info.TimeRoundStart
		b.Secs = int(diff.Seconds())
		b.X = e.Player.Position().X
		b.Y = e.Player.Position().Y
		b.Z = e.Player.Position().Z
		b.SteamID = e.Player.SteamID64
		b.PlayerName = e.Player.Name

		site := sites[int(e.Site)]

		b.BombSite = site
		b.EventType = "BombPlanted"

		round_info.BombStuff = append(round_info.BombStuff, b)

	})
}

func bomb_plantBegin(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	p.RegisterEventHandler(func(e events.BombPlantBegin) {
		gs := p.GameState()

		if m == nil || s == nil {
			log.Printf("bomb_planted: nil state m=%v s=%v", m, s)
			return
		}
		if e.Player == nil {
			log.Printf("bomb_planted: nil planter at tick %d", p.GameState().IngameTick())
			return
		}

		if len(m.Rounds) == 0 {
			log.Printf("bomb_planted: no rounds exist at tick %d; ignoring", p.GameState().IngameTick())
			return
		}

		// Validate/repair s.round (1-based)
		if s.round < 1 || s.round > len(m.Rounds) {
			log.Printf("bomb_planted: s.round=%d out of [1,%d]; recovering to last",
				s.round, len(m.Rounds))
			s.round = len(m.Rounds)
		}

		idx := s.round - 1
		round_info := &m.Rounds[idx]

		sites := map[int]rune{
			65: 'A',
			66: 'B',
		}

		var b BombStates

		b.Tick = gs.IngameTick()
		diff := p.CurrentTime() - round_info.TimeRoundStart
		b.Secs = int(diff.Seconds())
		b.X = e.Player.Position().X
		b.Y = e.Player.Position().Y
		b.Z = e.Player.Position().Z
		b.SteamID = e.Player.SteamID64
		b.PlayerName = e.Player.Name

		site := sites[int(e.Site)]

		b.BombSite = site
		b.EventType = "BombPlantBegin"

		round_info.BombStuff = append(round_info.BombStuff, b)
	})
}

func bomb_plantAbort(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	p.RegisterEventHandler(func(e events.BombPlantAborted) {
		gs := p.GameState()

		if m == nil || s == nil {
			log.Printf("bomb_planted: nil state m=%v s=%v", m, s)
			return
		}
		if e.Player == nil {
			log.Printf("bomb_planted: nil planter at tick %d", p.GameState().IngameTick())
			return
		}

		if len(m.Rounds) == 0 {
			log.Printf("bomb_planted: no rounds exist at tick %d; ignoring", p.GameState().IngameTick())
			return
		}

		// Validate/repair s.round (1-based)
		if s.round < 1 || s.round > len(m.Rounds) {
			log.Printf("bomb_planted: s.round=%d out of [1,%d]; recovering to last",
				s.round, len(m.Rounds))
			s.round = len(m.Rounds)
		}

		idx := s.round - 1
		round_info := &m.Rounds[idx]

		var b BombStates

		b.Tick = gs.IngameTick()
		diff := p.CurrentTime() - round_info.TimeRoundStart
		b.Secs = int(diff.Seconds())
		b.X = e.Player.Position().X
		b.Y = e.Player.Position().Y
		b.Z = e.Player.Position().Z
		b.SteamID = e.Player.SteamID64
		b.PlayerName = e.Player.Name

		b.EventType = "BombPlantAbort"

		round_info.BombStuff = append(round_info.BombStuff, b)
	})
}

func bomb_defuseStart(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	p.RegisterEventHandler(func(e events.BombDefuseStart) {
		gs := p.GameState()

		if m == nil || s == nil {
			log.Printf("bomb_planted: nil state m=%v s=%v", m, s)
			return
		}
		if e.Player == nil {
			log.Printf("bomb_planted: nil planter at tick %d", p.GameState().IngameTick())
			return
		}

		if len(m.Rounds) == 0 {
			log.Printf("bomb_planted: no rounds exist at tick %d; ignoring", p.GameState().IngameTick())
			return
		}

		// Validate/repair s.round (1-based)
		if s.round < 1 || s.round > len(m.Rounds) {
			log.Printf("bomb_planted: s.round=%d out of [1,%d]; recovering to last",
				s.round, len(m.Rounds))
			s.round = len(m.Rounds)
		}

		idx := s.round - 1
		round_info := &m.Rounds[idx]

		var b BombStates

		b.Tick = gs.IngameTick()
		diff := p.CurrentTime() - round_info.TimeRoundStart
		b.Secs = int(diff.Seconds())
		b.X = e.Player.Position().X
		b.Y = e.Player.Position().Y
		b.Z = e.Player.Position().Z
		b.SteamID = e.Player.SteamID64
		b.PlayerName = e.Player.Name
		b.HasKit = e.HasKit
		b.EventType = "BombDefuseStart"

		round_info.BombStuff = append(round_info.BombStuff, b)
	})
}

func bomb_defuseAborted(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	p.RegisterEventHandler(func(e events.BombDefuseAborted) {
		gs := p.GameState()

		if m == nil || s == nil {
			log.Printf("bomb_planted: nil state m=%v s=%v", m, s)
			return
		}
		if e.Player == nil {
			log.Printf("bomb_planted: nil planter at tick %d", p.GameState().IngameTick())
			return
		}

		if len(m.Rounds) == 0 {
			log.Printf("bomb_planted: no rounds exist at tick %d; ignoring", p.GameState().IngameTick())
			return
		}

		// Validate/repair s.round (1-based)
		if s.round < 1 || s.round > len(m.Rounds) {
			log.Printf("bomb_planted: s.round=%d out of [1,%d]; recovering to last",
				s.round, len(m.Rounds))
			s.round = len(m.Rounds)
		}

		idx := s.round - 1
		round_info := &m.Rounds[idx]

		var b BombStates

		b.Tick = gs.IngameTick()
		diff := p.CurrentTime() - round_info.TimeRoundStart
		b.Secs = int(diff.Seconds())
		b.X = e.Player.Position().X
		b.Y = e.Player.Position().Y
		b.Z = e.Player.Position().Z
		b.SteamID = e.Player.SteamID64
		b.PlayerName = e.Player.Name
		b.EventType = "BombDefuseAborted"

		round_info.BombStuff = append(round_info.BombStuff, b)
	})
}

func bomb_defused(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	p.RegisterEventHandler(func(e events.BombDefused) {
		gs := p.GameState()

		if m == nil || s == nil {
			log.Printf("bomb_planted: nil state m=%v s=%v", m, s)
			return
		}
		if e.Player == nil {
			log.Printf("bomb_planted: nil planter at tick %d", p.GameState().IngameTick())
			return
		}

		if len(m.Rounds) == 0 {
			log.Printf("bomb_planted: no rounds exist at tick %d; ignoring", p.GameState().IngameTick())
			return
		}

		// Validate/repair s.round (1-based)
		if s.round < 1 || s.round > len(m.Rounds) {
			log.Printf("bomb_planted: s.round=%d out of [1,%d]; recovering to last",
				s.round, len(m.Rounds))
			s.round = len(m.Rounds)
		}

		idx := s.round - 1
		round_info := &m.Rounds[idx]

		sites := map[int]rune{
			65: 'A',
			66: 'B',
		}

		var b BombStates

		b.Tick = gs.IngameTick()
		diff := p.CurrentTime() - round_info.TimeRoundStart
		b.Secs = int(diff.Seconds())
		b.X = e.Player.Position().X
		b.Y = e.Player.Position().Y
		b.Z = e.Player.Position().Z
		b.SteamID = e.Player.SteamID64
		b.PlayerName = e.Player.Name

		site := sites[int(e.Site)]

		b.HasKit = e.Player.HasDefuseKit()
		b.BombSite = site
		b.EventType = "BombDefused"

		round_info.BombStuff = append(round_info.BombStuff, b)
	})
}

func remove_dups(p demoinfocs.Parser, m *MatchInfo, c []*common.Player) {

	for _, pl := range c {
		player_id := pl.SteamID64

		player, exists := m.Players[player_id]
		if !exists {
			return
		}

		seen := make(map[int]bool)
		var result []int

		for _, val := range player.RoundContributed {
			if !seen[val] {
				// If we haven't seen this value yet, append it to result
				seen[val] = true
				result = append(result, val)
			}
		}

		player.RoundContributed = result
		m.Players[player_id] = player
	}
}

func get_match_winner(m *MatchInfo) {
	if len(m.Rounds) == 0 {
		log.Println("This is whats wrong", m)
		return
	}
	last := m.Rounds[len(m.Rounds)-1]

	m.Winner = last.RoundClanWinner
	m.Loser = last.RoundClanLoser
}

package main

import (
	"time"

	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
)

func flash_logic(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	p.RegisterEventHandler(func(e events.FlashExplode) {
		if e.Thrower == nil {
			return
		}

		gs := p.GameState()

		players_a := gs.TeamCounterTerrorists().Members()
		players_b := gs.TeamTerrorists().Members()

		for _, player := range players_a {
			if player.FlashDurationTime() >= 2000*time.Millisecond {
				throw_id := e.Thrower.SteamID64

				pl, exists := m.Players[throw_id]
				if !exists {
					return
				}

				pl.EffectiveFlashes++

				m.Players[throw_id] = pl
			}
		}

		for _, player := range players_b {
			if player.IsBlinded() && time.Duration(player.FlashDuration) >= time.Duration(2*time.Second) {
				throw_id := e.Thrower.SteamID64

				pl, exists := m.Players[throw_id]
				if !exists {
					return
				}

				pl.EffectiveFlashes++

				m.Players[throw_id] = pl
			}
		}
	})
}

func nade_dmg(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	p.RegisterEventHandler(func(e events.PlayerHurt) {
		if e.Player == nil {
			return
		}

		if e.Weapon == nil {
			return
		}

		if e.Weapon.Type == 506 {
			att_id := e.Attacker.SteamID64

			player, exists := m.Players[att_id]
			if !exists {
				return
			}

			player.NadeDmg += e.HealthDamageTaken
			m.Players[att_id] = player
		}

		if e.Weapon.Type == 503 {
			att_id := e.Attacker.SteamID64

			player, exists := m.Players[att_id]

			if !exists {
				return
			}

			player.InfernoDmg += e.HealthDamageTaken
			m.Players[att_id] = player
		}
	})
}

func nades(p demoinfocs.Parser, m *MatchInfo, s *parsingState) {
	if p.GameState().IsWarmupPeriod() {
		return
	}

	if s.SeenNadeIDs == nil {
		s.SeenNadeIDs = make(map[int]struct{})
	}

	// Reset per round; entity IDs can be reused
	p.RegisterEventHandler(func(events.RoundStart) {
		s.SeenNadeIDs = make(map[int]struct{})
	})

	p.RegisterEventHandler(func(e events.GrenadeEventIf) {
		id := e.Base().GrenadeEntityID
		if id == 0 {
			return
		}

		if e.Base().Thrower == nil {
			//log.Printf("InfernoStartBurn with nil thrower @tick=%d; skipping", p.GameState().IngameTick())
			return
		}

		if e.Base().Grenade == nil {
			//log.Printf("Grenade is Nil @tick=%d, type=%s", p.GameState().IngameTick(), e.Base().GrenadeType)
			return
		}

		gs := p.GameState()
		ri := &m.Rounds[s.round-1]

		// seconds from round start (float64 is better, but keep your int64 for now)
		secFromRoundStart := int64((p.CurrentTime() - ri.TimeRoundStart) / time.Second)

		// FIRST sighting -> START (dedup)
		if _, seen := s.SeenNadeIDs[id]; !seen {
			s.SeenNadeIDs[id] = struct{}{}

			var sn NadeEventStart
			sn.Tick = gs.IngameTick()
			sn.Secs = secFromRoundStart // ABS: seconds from round start
			sn.EntityId = id
			sn.NadeType = int(e.Base().GrenadeType)
			sn.NadePosX = e.Base().Position.X
			sn.NadePosY = e.Base().Position.Y
			if thr := e.Base().Thrower; thr != nil {
				sn.ThrowerID = thr.SteamID64
				sn.ThrowerName = thr.Name
				sn.ThrowerPosX = thr.Position().X
				sn.ThrowerPosY = thr.Position().Y
				sn.ThrowerSide = int(thr.Team)
				// If you want clan, add a separate field (don’t overwrite name)
				// sn.ThrowerClan = thr.ClanTag()
			}
			ri.NadeStarts = append(ri.NadeStarts, sn)
			return // IMPORTANT: don't also treat this same callback as an end
		}

		// Already seen -> candidate END (only once per entity)
		if !hasEnd(ri.NadeEnds, id) {
			var en NadeEventEnd
			en.DestroyTick = gs.IngameTick()
			en.DestroySecs = secFromRoundStart // ABS: seconds from round start
			en.EntityId = id
			en.NadeType = int(e.Base().GrenadeType)
			en.NadePosX = e.Base().Position.X
			en.NadePosY = e.Base().Position.Y
			if thr := e.Base().Thrower; thr != nil {
				en.ThrowerID = thr.SteamID64
				en.ThrowerName = thr.Name
				en.ThrowerPosX = thr.Position().X
				en.ThrowerPosY = thr.Position().Y
				en.ThrowerSide = int(thr.Team)
			}
			ri.NadeEnds = append(ri.NadeEnds, en)
		}
	})
}

func hasEnd(ends []NadeEventEnd, id int) bool {
	for i := range ends {
		if ends[i].EntityId == id {
			return true
		}
	}
	return false
}

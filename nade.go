package main

import (
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/events"
	"time"
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

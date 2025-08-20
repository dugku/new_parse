package main

import (
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	"math"
)

func calc_adr(gs demoinfocs.GameState, dmg int) float64 {
	rounds_played := gs.TotalRoundsPlayed()

	adr := float64(dmg) / float64(rounds_played)

	if math.IsNaN(adr) || math.IsInf(adr, 0) {
		return 0
	}

	return math.Round(adr*100) / 100
}

func calc_kd(k int, d int) float64 {
	ratio := float64(k) / float64(d)

	if math.IsNaN(ratio) || math.IsInf(ratio, 0) {
		return 0
	}

	return math.Round(ratio*100) / 100
}

func calc_hs_per(k int, d int) float64 {
	hs_per := (float64(k) / float64(d)) * 100

	if math.IsNaN(hs_per) || math.IsInf(hs_per, 0) {
		return 0
	}

	return math.Round(hs_per*100) / 100
}

func calc_avg_per_round(i int, r int) float64 {
	avg_per_round := float64(i) / float64(r)

	if math.IsNaN(avg_per_round) || math.IsInf(avg_per_round, 0) {
		return 0
	}

	return math.Round(avg_per_round*100) / 100
}

func calc_impatct(ak float64, ava float64) float64 {
	impact := (2.13 * ak) + (0.42 * ava) - 0.41

	if math.IsNaN(impact) || math.IsInf(impact, 0) {
		return 0
	}

	return math.Round(impact*100) / 100
}

// crude but gets the job done.
func calc_kast(l int, r int) float64 {
	kast := float64(l) / float64(r)

	if math.IsNaN(kast) || math.IsInf(kast, 0) {
		return 0
	}

	return math.Round(kast*10000) / 10000
}

func calc_ndmg(d int, r int) float64 {
	avg_n_dmg := float64(d) / float64(r)

	if math.IsNaN(avg_n_dmg) || math.IsInf(avg_n_dmg, 0) {
		return 0
	}

	return math.Round(avg_n_dmg*100) / 100
}

func calc_infer_dmg(d int, r int) float64 {
	avg_infer_dmg := float64(d) / float64(r)

	if math.IsNaN(avg_infer_dmg) || math.IsInf(avg_infer_dmg, 0) {
		return 0
	}

	return math.Round(avg_infer_dmg*100) / 100
}

func calc_open_percent(o int, r int) float64 {
	opening_percent := (float64(o) / float64(r)) * 100

	if math.IsNaN(opening_percent) || math.IsInf(opening_percent, 0) {
		return 0
	}

	return math.Round(opening_percent*100) / 100
}

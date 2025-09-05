package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	ex "github.com/markus-wa/demoinfocs-golang/v4/examples"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v4/pkg/demoinfocs/msg"
)

var (
	mapMetadata ex.Map
)

func main() {
	paths := get_files()

	count := 0
	for _, path := range paths {
		state := parsingState{}
		match := MatchInfo{
			Map:       "",
			TeamA:     "",
			TeamB:     "",
			Date:      "",
			MatchId:   "",
			TeamOneId: "",
			TeamTwoId: "",
			Rounds:    []RoundInformation{},
			Players:   map[uint64]playerStat{},
		}

		err4 := start_parsing(path, &match, &state)
		if err4 != nil {
			log.Printf("Error parsing demo %s: %v\n", path, err4)
			continue
		}

		get_match_winner(&match)
		fmt.Println(path)
		temp := strings.Split(path, string(filepath.Separator))[4]
		fmt.Println(temp)
		temp3 := strings.Split(temp, "_")
		match_id_thing := strings.Split(temp3[0], "-")[0]
		team_one := strings.Split(temp3[0], "-")[2]
		team_two := strings.Split(temp3[0], "-")[3]

		date := strings.Split(path, string(os.PathSeparator))[1]

		match.MatchId = match_id_thing
		match.TeamOneId = team_one
		match.TeamTwoId = team_two
		match.Date = date
		output_file_n := fmt.Sprintf("%s-%s-%s-%d-k.json", match_id_thing, team_one, team_two, count)
		targetDir := `/mnt/c/Users/Mike/Desktop/Parse/json_matches`

		// Ensure the directory is created (won’t fail if it already exists)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			log.Fatal(err)
		}

		// Build the full path to the JSON file
		outputPath := filepath.Join(targetDir, output_file_n)
		// Marshal your data

		jsonData, err := json.MarshalIndent(match, "", "  ")
		check(err)
		// Write the file to the specified directory

		err = os.WriteFile(outputPath, jsonData, 0644)
		check(err)
		fmt.Println("JSON written to:", outputPath)

		count++
	}

}

func start_parsing(path string, m *MatchInfo, s *parsingState) error {
	fmt.Printf("Parsing %s\n", path)
	f, err := os.Open(path)
	if err != nil {
		log.Panic("failed to open demo file: ", err)
	}
	defer f.Close()
	p := demoinfocs.NewParser(f)
	defer p.Close()

	p.RegisterNetMessageHandler(func(msg *msg.CSVCMsg_ServerInfo) {
		// Get metadata for the map that the game was played on for coordinate translations
		fmt.Println(msg.MapName)

	})
	fmt.Println(p.TickRate())
	state_control(p, s, m)
	match_started(p, m, s)
	team_switch(p, m, s)
	score_updater(p, m, s)
	player_getter(p, m, s)
	kill_handler(p, m, s)
	get_pres_round_kill(p, m, s)
	round_end_logic(p, m, s)
	flash_logic(p, m, s)
	nade_dmg(p, m, s)
	round_econ_logic(p, m, s)
	bom_planted(p, m, s)
	bomb_plantBegin(p, m, s)
	bomb_plantAbort(p, m, s)
	bomb_defuseStart(p, m, s)
	bomb_defuseAborted(p, m, s)
	bomb_defused(p, m, s)
	players_hurt(p, m, s)
	player_fired(p, m, s)
	nades(p, m, s)
	open_kill(m, s)
	// Parse to end
	err = p.ParseToEnd()
	if err != nil {
		// If the error is an unexpected EOF, log it as a warning instead of panicking.
		if errors.Is(err, io.EOF) || errors.Is(err, demoinfocs.ErrUnexpectedEndOfDemo) || strings.Contains(err.Error(), "index out of range") {
			log.Printf("demo parsing finished with warning: %v", err)
		} else {
			log.Panic("failed to parse demo: ", err)
		}
	}
	return nil
}

func deleteElement(slice []string, index int) []string {
	return append(slice[:index], slice[index+1:]...)
}

func get_files() []string {
	filepaths := make([]string, 0)
	path := `/mnt/d`

	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("failed to access %q: %v\n", p, err)
			return nil // Continue walking instead of stopping.
		}
		if d.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(p)) != ".dem" {
			return nil
		}
		filepaths = append(filepaths, p)
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(filepaths))
	return filepaths
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

type Hole19Data struct {
	Hole19Rounds `json:"rounds"`
}
type Hole19Rounds []struct {
	ShareID     string `json:"share_id"`
	StartedAt   string `json:"started_at"`
	EndedAt     string `json:"ended_at"`
	InputMode   string `json:"input_mode"`
	ScoringMode string `json:"scoring_mode"`
	Course      struct {
		Name string `json:"name"`
	} `json:"course"`
	Tee struct {
		Name string `json:"name"`
	} `json:"tee"`
	Handicap        float64 `json:"handicap"`
	PlayingHandicap int     `json:"playing_handicap"`
	DistanceWalked  int     `json:"distance_walked"`
	Steps           any     `json:"steps"`
	HcpRound        any     `json:"hcp_round"`
	Scores          []struct {
		Hole struct {
			Sequence int `json:"sequence"`
			Si       int `json:"si"`
			Par      int `json:"par"`
		} `json:"hole"`
		TotalOfStrokes    int    `json:"total_of_strokes"`
		TotalOfPutts      int    `json:"total_of_putts"`
		TotalOfSandShots  int    `json:"total_of_sand_shots"`
		TotalOfPenalties  int    `json:"total_of_penalties"`
		FairwayHit        string `json:"fairway_hit"`
		GreenInRegulation bool   `json:"green_in_regulation"`
		Scrambling        bool   `json:"scrambling"`
		SandSaves         bool   `json:"sand_saves"`
		Scratched         bool   `json:"scratched"`
		UpAndDown         bool   `json:"up_and_down"`
		PossibleUpAndDown bool   `json:"possible_up_and_down"`
	} `json:"scores"`
	Multiplayers []any `json:"multiplayers"`
}

type MostBirdiesTracker struct {
	Course       string
	Date         time.Time
	HolesBirdied []int
}

func main() {
	rawData, err := os.ReadFile("golf-rounds.json")
	if err != nil {
		log.Fatal(err)
	}
	var data Hole19Data
	err = json.Unmarshal(rawData, &data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v rounds tracked\n", len(data.Hole19Rounds))

	// put the data into a map by years
	annualRoundMap := make(map[int]Hole19Rounds)
	for _, r := range data.Hole19Rounds {
		t, err := time.Parse(time.DateTime+" MST", r.StartedAt)
		if err != nil {
			log.Fatal(err)
		}
		annualRoundMap[t.Local().Year()] = append(annualRoundMap[t.Local().Year()], r)
	}
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{
		"Year",
		"Rounds",
		"Tracked\nHoles",
		"-3",
		"-2",
		"-1",
		"0",
		"+1",
		"+2",
		"+3",
		"GIR %",
		"BM %",
		"BPR",
		"Most Birdies\nIn 1 Round",
	})
	for year, rounds := range annualRoundMap {
		t.AppendRow(yearAggScores(year, rounds))
	}
	t.SortBy([]table.SortBy{
		{Name: "Year", Mode: table.Asc},
	})
	t.Render()
}

func yearAggScores(year int, rounds Hole19Rounds) table.Row {
	var tHoles, tMinus3, tMinus2, tMinus1, tPars, tPlus1, tPlus2, tPlus3, gir, bM int
	var mostBirdsPerRound []MostBirdiesTracker
	for _, r := range rounds {
		var holesBirdiedPerRoundCounter []int
		for _, hole := range r.Scores {
			tHoles++
			diff := hole.TotalOfStrokes - hole.Hole.Par
			if diff < 0 {
				holesBirdiedPerRoundCounter = append(holesBirdiedPerRoundCounter, hole.Hole.Sequence)
			}
			if hole.GreenInRegulation {
				gir++
				if diff < 0 {
					bM++
				}
			}
			switch diff {
			case -3:
				tMinus3++
			case -2:
				tMinus2++
			case -1:
				tMinus1++
			case 0:
				tPars++
			case 1:
				tPlus1++
			case 2:
				tPlus2++
			case 3:
				tPlus3++
			default:
			}
		}
		t, err := time.Parse(time.DateTime+" MST", r.StartedAt)
		if err != nil {
			log.Fatal(err)
		}
		if len(mostBirdsPerRound) == 0 && len(holesBirdiedPerRoundCounter) > 0 {
			mostBirdsPerRound = append(mostBirdsPerRound, MostBirdiesTracker{
				Course:       r.Course.Name,
				Date:         t,
				HolesBirdied: holesBirdiedPerRoundCounter,
			})
		} else if len(mostBirdsPerRound) > 0 && len(holesBirdiedPerRoundCounter) == len(mostBirdsPerRound[0].HolesBirdied) {
			mostBirdsPerRound = append(mostBirdsPerRound, MostBirdiesTracker{
				Course:       r.Course.Name,
				Date:         t,
				HolesBirdied: holesBirdiedPerRoundCounter,
			})
		} else if len(mostBirdsPerRound) > 0 && len(holesBirdiedPerRoundCounter) > len(mostBirdsPerRound[0].HolesBirdied) {
			mostBirdsPerRound = nil
			mostBirdsPerRound = append(mostBirdsPerRound, MostBirdiesTracker{
				Course:       r.Course.Name,
				Date:         t,
				HolesBirdied: holesBirdiedPerRoundCounter,
			})
		}
	}
	girPercent := (float64(gir) / float64(tHoles)) * 100
	birdiesMadeWithGIR := (float64(bM) / float64(gir)) * 100
	//fmt.Printf("%v GIR and birdies made on those %v in %v\n", gir, bM, year)
	return table.Row{
		year,
		len(rounds),
		tHoles,
		tMinus3,
		tMinus2,
		tMinus1,
		tPars,
		tPlus1,
		tPlus2,
		tPlus3,
		fmt.Sprintf("%v%%", math.Round(girPercent*100)/100),
		fmt.Sprintf("%v%%", math.Round(birdiesMadeWithGIR*100)/100),
		fmt.Sprintf("%v", math.Round(float64(bM)/(float64(tHoles)/18)*100)/100),
		newLineBirds(mostBirdsPerRound),
	}
}

func newLineBirds(birdsList []MostBirdiesTracker) string {
	var s string
	for _, bd := range birdsList {
		s += fmt.Sprintf("- %v at %v birdied %v\n", bd.Date.Local().Format(time.DateOnly), bd.Course, bd.HolesBirdied)
	}
	return strings.TrimSuffix(s, "\n")
}

package types

import (
	"encoding/json"
	"strconv"
)

type TimingAppDataResponse struct {
	TimingAppData TimingAppData `json:"TimingAppData"`
}

type TimingAppData struct {
	Lines map[string]TimingAppDataLine `json:"Lines"`
}

type TimingAppDataLine struct {
	RacingNumber string `json:"RacingNumber"`
	Line         int    `json:"Line"`
	GridPos      string `json:"GridPos"`
	Stints       Stints `json:"Stints"`
}

type Stint struct {
	LapTime         string `json:"LapTime"`
	LapNumber       int    `json:"LapNumber"`
	LapFlags        *int   `json:"LapFlags"`
	Compound        string `json:"Compound"`
	New             string `json:"New"`
	TyresNotChanged string `json:"TyresNotChanged"`
	TotalLaps       *int   `json:"TotalLaps"`
	StartLaps       *int   `json:"StartLaps"`
}

type Stints map[string]Stint

func (s *Stints) UnmarshalJSON(data []byte) error {
	var stintMap = make(map[string]Stint)

	if len(data) > 0 && data[0] == '[' {

		var slice []Stint
		if err := json.Unmarshal(data, &slice); err != nil {
			return err
		}

		for i, stint := range slice {
			stintMap[strconv.Itoa(i)] = stint
		}

	} else if err := json.Unmarshal(data, &stintMap); err != nil {
		return err
	}

	*s = stintMap
	return nil
}

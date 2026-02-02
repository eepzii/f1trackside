package types

import (
	"encoding/json"
	"strconv"
)

type TimingDataResponse struct {
	TimingData TimingData `json:"TimingData"`
}

type TimingData struct {
	Lines map[string]TimingDataLine `json:"Lines"`
}

type TimingDataLine struct {
	GapToLeader             string                  `json:"GapToLeader"`
	IntervalToPositionAhead IntervalToPositionAhead `json:"IntervalToPositionAhead"`
	Line                    int                     `json:"Line"`
	Position                string                  `json:"Position"`
	ShowPosition            *bool                   `json:"ShowPosition"`
	RacingNumber            string                  `json:"RacingNumber"`
	Retired                 *bool                   `json:"Retired"`
	InPit                   *bool                   `json:"InPit"`
	PitOut                  *bool                   `json:"PitOut"`
	Stopped                 *bool                   `json:"Stopped"`
	Status                  *int                    `json:"Status"`
	NumberOfLaps            int                     `json:"NumberOfLaps"`
	NumberOfPitStops        *int                    `json:"NumberOfPitStops"`
	Sectors                 Sectors                 `json:"Sectors"`
	Speeds                  map[string]Speed        `json:"Speeds"`
	BestLapTime             BestLapTime             `json:"BestLapTime"`
	LastLapTime             LastLapTime             `json:"LastLapTime"`
}

type IntervalToPositionAhead struct {
	Value    string `json:"Value"`
	Catching *bool  `json:"Catching"`
}

type Sectors map[string]Sector

type Sector struct {
	Stopped         *bool    `json:"Stopped"`
	PreviousValue   string   `json:"PreviousValue"`
	Segments        Segments `json:"Segments"`
	Value           string   `json:"Value"`
	Status          *int     `json:"Status"`
	OverallFastest  *bool    `json:"OverallFastest"`
	PersonalFastest *bool    `json:"PersonalFastest"`
}

type Segments map[string]Segment

type Segment struct {
	Status *int `json:"Status"`
}

type Speed struct {
	Value           string `json:"Value"`
	Status          *int   `json:"Status"`
	OverallFastest  *bool  `json:"OverallFastest"`
	PersonalFastest *bool  `json:"PersonalFastest"`
}

type BestLapTime struct {
	Value string `json:"Value"`
	Lap   int    `json:"Lap"`
}

type LastLapTime struct {
	Value           string `json:"Value"`
	Status          *int   `json:"Status"`
	OverallFastest  *bool  `json:"OverallFastest"`
	PersonalFastest *bool  `json:"PersonalFastest"`
}

func (s *Sectors) UnmarshalJSON(data []byte) error {
	var sectorMap = make(map[string]Sector)

	if len(data) > 0 && data[0] == '[' {

		var slice []Sector
		if err := json.Unmarshal(data, &slice); err != nil {
			return err
		}

		for i, sector := range slice {
			sectorMap[strconv.Itoa(i)] = sector
		}

	} else if err := json.Unmarshal(data, &sectorMap); err != nil {
		return err
	}

	*s = sectorMap
	return nil
}

func (s *Segments) UnmarshalJSON(data []byte) error {
	var segmentMap = make(map[string]Segment)

	if len(data) > 0 && data[0] == '[' {

		var slice []Segment
		if err := json.Unmarshal(data, &slice); err != nil {
			return err
		}

		for i, segment := range slice {
			segmentMap[strconv.Itoa(i)] = segment
		}

	} else if err := json.Unmarshal(data, &segmentMap); err != nil {
		return err
	}

	*s = segmentMap
	return nil
}

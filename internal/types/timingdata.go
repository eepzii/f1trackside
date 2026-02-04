package types

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
	m, err := unmarshalDynamicJSON[Sector](data)
	if err != nil {
		return err
	}
	*s = m
	return nil
}

func (s *Segments) UnmarshalJSON(data []byte) error {
	m, err := unmarshalDynamicJSON[Segment](data)
	if err != nil {
		return err
	}
	*s = m
	return nil
}

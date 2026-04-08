package types

type TimingDataResponse struct {
	TimingData TimingData `json:"TimingData"`
}

type TimingData struct {
	CutOffPercentage string                    `json:"CutOffPercentage"`
	CutOffTime       *string                   `json:"CutOffTime"`
	Lines            map[string]TimingDataLine `json:"Lines"`
	NoEntries        DynamicJSON[*int]         `json:"NoEntries"`
	SessionPart      *int                      `json:"SessionPart"`
	Withheld         *bool                     `json:"Withheld"`
	KeyFrame         bool                      `json:"_kf"`
}

type TimingDataLine struct {
	BestLapTime             BestLapTime               `json:"BestLapTime"`
	BestLapTimes            DynamicJSON[BestLapTimes] `json:"BestLapTimes"`
	CutOff                  *bool                     `json:"Cutoff"`
	GapToLeader             *string                   `json:"GapToLeader"`
	InPit                   *bool                     `json:"InPit"`
	IntervalToPositionAhead IntervalToPositionAhead   `json:"IntervalToPositionAhead"`
	KnockedOut              *bool                     `json:"KnockedOut"`
	LastLapTime             LastLapTime               `json:"LastLapTime"`
	Line                    int                       `json:"Line"`
	NumberOfLaps            int                       `json:"NumberOfLaps"`
	NumberOfPitStops        int                       `json:"NumberOfPitStops"`
	PitOut                  *bool                     `json:"PitOut"`
	Position                string                    `json:"Position"`
	RacingNumber            string                    `json:"RacingNumber"`
	Retired                 *bool                     `json:"Retired"`
	Sectors                 DynamicJSON[Sector]       `json:"Sectors"`
	ShowPosition            *bool                     `json:"ShowPosition"`
	Speeds                  map[string]Speed          `json:"Speeds"`
	Stats                   DynamicJSON[Stat]         `json:"Stats"`
	Status                  *int                      `json:"Status"`
	Stopped                 *bool                     `json:"Stopped"`
	TimeDiffToFastest       *string                   `json:"TimeDiffToFastest"`
	TimeDiffToPositionAhead *string                   `json:"TimeDiffToPositionAhead"`
	Deleted                 []string                  `json:"_deleted"`
}

type BestLapTime struct {
	Lap     int      `json:"Lap"`
	Value   *string  `json:"Value"`
	Deleted []string `json:"_deleted"`
}

type BestLapTimes struct {
	Lap   int     `json:"Lap"`
	Value *string `json:"Value"`
}

type IntervalToPositionAhead struct {
	Catching *bool   `json:"Catching"`
	Value    *string `json:"Value"`
}

type LastLapTime struct {
	OverallFastest  *bool   `json:"OverallFastest"`
	PersonalFastest *bool   `json:"PersonalFastest"`
	Status          *int    `json:"Status"`
	Value           *string `json:"Value"`
}

type Sector struct {
	OverallFastest  *bool                `json:"OverallFastest"`
	PersonalFastest *bool                `json:"PersonalFastest"`
	PreviousValue   string               `json:"PreviousValue"`
	Segments        DynamicJSON[Segment] `json:"Segments"`
	Status          *int                 `json:"Status"`
	Stopped         *bool                `json:"Stopped"`
	Value           *string              `json:"Value"`
}

type Segment struct {
	Status *int `json:"Status"`
}

type Speed struct {
	OverallFastest  *bool   `json:"OverallFastest"`
	PersonalFastest *bool   `json:"PersonalFastest"`
	Status          *int    `json:"Status"`
	Value           *string `json:"Value"`
}

type Stat struct {
	TimeDiffToFastest       *string `json:"TimeDiffToFastest"`
	TimeDiffToPositionAhead *string `json:"TimeDifftoPositionAhead"`
}

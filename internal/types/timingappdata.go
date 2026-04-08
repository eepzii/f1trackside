package types

type TimingAppDataResponse struct {
	TimingAppData TimingAppData `json:"TimingAppData"`
}

type TimingAppData struct {
	Lines    map[string]TimingAppDataLine `json:"Lines"`
	KeyFrame bool                         `json:"_kf"`
}

type TimingAppDataLine struct {
	GridPos      string             `json:"GridPos"`
	Line         int                `json:"Line"`
	RacingNumber string             `json:"RacingNumber"`
	Stints       DynamicJSON[Stint] `json:"Stints"`
}

type Stint struct {
	Compound        string `json:"Compound"`
	LapFlags        *int   `json:"LapFlags"`
	LapNumber       int    `json:"LapNumber"`
	LapTime         string `json:"LapTime"`
	New             string `json:"New"`
	StartLaps       *int   `json:"StartLaps"`
	TotalLaps       *int   `json:"TotalLaps"`
	TyresNotChanged string `json:"TyresNotChanged"`
}

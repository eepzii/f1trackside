package types

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
	m, err := unmarshalDynamicJSON[Stint](data)
	if err != nil {
		return err
	}
	*s = m
	return nil
}

package types

type LapCountResponse struct {
	LapCount LapCount `json:"LapCount"`
}

type LapCount struct {
	CurrentLap *int `json:"CurrentLap"`
	TotalLaps  *int `json:"TotalLaps"`
	KeyFrame   bool `json:"_kf"`
}

package types

type TrackStatusResponse struct {
	TrackStatus TrackStatus `json:"TrackStatus"`
}

type TrackStatus struct {
	Message  string `json:"Message"`
	Status   string `json:"Status"`
	KeyFrame bool   `json:"_kf"`
}

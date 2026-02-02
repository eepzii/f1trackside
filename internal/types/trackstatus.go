package types

type TrackStatusResponse struct {
	TrackStatus TrackStatus `json:"TrackStatus"`
}

type TrackStatus struct {
	Status     string `json:"Status"`
	Message    string `json:"Message"`
	IsKeyFrame *bool  `json:"_kf"`
}

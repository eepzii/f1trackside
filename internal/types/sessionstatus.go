package types

type SessionStatusResponse struct {
	SessionStatus SessionStatus `json:"SessionStatus"`
}

type SessionStatus struct {
	Status     string `json:"Status"`
	Started    string `json:"Started"`
	IsKeyFrame *bool  `json:"_kf"`
}

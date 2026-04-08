package types

type SessionStatusResponse struct {
	SessionStatus SessionStatus `json:"SessionStatus"`
}

type SessionStatus struct {
	Started  string `json:"Started"`
	Status   string `json:"Status"`
	KeyFrame bool   `json:"_kf"`
}

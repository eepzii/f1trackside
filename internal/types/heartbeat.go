package types

type HeartbeatResponse struct {
	Heartbeat Heartbeat `json:"Heartbeat"`
}

type Heartbeat struct {
	UTC      string `json:"Utc"`
	KeyFrame bool   `json:"_kf"`
}

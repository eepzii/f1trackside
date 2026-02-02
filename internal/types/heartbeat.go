package types

type HeartbeatResponse struct {
	Heartbeat Heartbeat `json:"Heartbeat"`
}

type Heartbeat struct {
	Utc        string `json:"Utc"`
	IsKeyFrame *bool  `json:"_kf"`
}

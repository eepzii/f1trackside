package types

type SessionInfoResponse struct {
	SessionInfo SessionInfo `json:"SessionInfo"`
}

type SessionInfo struct {
	Meeting       Meeting       `json:"Meeting"`
	SessionStatus string        `json:"SessionStatus"`
	ArchiveStatus ArchiveStatus `json:"ArchiveStatus"`
	Key           int           `json:"Key"`
	Type          string        `json:"Type"`
	Name          string        `json:"Name"`
	StartDate     string        `json:"StartDate"`
	EndDate       string        `json:"EndDate"`
	GmtOffset     string        `json:"GmtOffset"`
	Path          string        `json:"Path"`
	IsKeyFrame    *bool         `json:"_kf"`
}

type Meeting struct {
	Key          int     `json:"Key"`
	Name         string  `json:"Name"`
	OfficialName string  `json:"OfficialName"`
	Location     string  `json:"Location"`
	Number       int     `json:"Number"`
	Country      Country `json:"Country"`
	Circuit      Circuit `json:"Circuit"`
}

type Country struct {
	Key  int    `json:"Key"`
	Code string `json:"Code"`
	Name string `json:"Name"`
}

type Circuit struct {
	Key       int    `json:"Key"`
	ShortName string `json:"ShortName"`
}

type ArchiveStatus struct {
	Status string `json:"Status"`
}

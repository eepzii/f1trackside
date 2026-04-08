package types

type SessionInfoResponse struct {
	SessionInfo SessionInfo `json:"SessionInfo"`
}

type SessionInfo struct {
	ArchiveStatus ArchiveStatus `json:"ArchiveStatus"`
	EndDate       string        `json:"EndDate"`
	GMTOffset     string        `json:"GmtOffset"`
	Key           int           `json:"Key"`
	Meeting       Meeting       `json:"Meeting"`
	Name          string        `json:"Name"`
	Number        int           `json:"Number"`
	Path          string        `json:"Path"`
	SessionStatus string        `json:"SessionStatus"`
	StartDate     string        `json:"StartDate"`
	Type          string        `json:"Type"`
	KeyFrame      bool          `json:"_kf"`
}

type ArchiveStatus struct {
	Status string `json:"Status"`
}

type Meeting struct {
	Circuit      Circuit `json:"Circuit"`
	Country      Country `json:"Country"`
	Key          int     `json:"Key"`
	Location     string  `json:"Location"`
	Name         string  `json:"Name"`
	Number       int     `json:"Number"`
	OfficialName string  `json:"OfficialName"`
}

type Circuit struct {
	Key       int    `json:"Key"`
	ShortName string `json:"ShortName"`
}

type Country struct {
	Code string `json:"Code"`
	Key  int    `json:"Key"`
	Name string `json:"Name"`
}

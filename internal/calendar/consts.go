package calendar

const (
	calendarURL = "https://ics.ecal.com/ecal-sub/6a329f7fab23640002745fb7/Formula%201.ics"

	UnknownLocation = "Unknown Location"
	UnknownSession  = "Unknown Session"

	cacheControl = "Cache-Control"
	maxAge       = "max-age="
)

type SessionType int

const (
	Unknown SessionType = iota
	Practice1
	Practice2
	Practice3
	SprintQualification
	Qualifying
	SprintRace
	Race
)

var sessionTypes = map[string]SessionType{
	"Practice 1":           Practice1,
	"Practice 2":           Practice2,
	"Practice 3":           Practice3,
	"Sprint Qualification": SprintQualification,
	"Qualifying":           Qualifying,
	"Sprint Race":          SprintRace,
	"Race":                 Race,
}

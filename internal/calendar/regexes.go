package calendar

import "regexp"

var (
	sessionRegex = regexp.MustCompile(`(Practice 1|Practice 2|Practice 3|Sprint Qualification|Qualifying|Sprint Race|Race)`)
)

package calendar

import "errors"

var (
	errCalendarFetch = errors.New("calendar fetch failed")
	errInvalidInput  = errors.New("invalid input parameter")
	errParseICS      = errors.New("failed to parse ics calendar")
)

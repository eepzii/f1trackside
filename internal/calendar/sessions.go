package calendar

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
)

func newSessions(icsCalendar *ics.Calendar, logger *slog.Logger) []Session {
	var f1Sessions []Session

	for _, event := range icsCalendar.Events() {
		session, err := eventToSession(event)
		if err != nil {
			logger.Error("session dropped",
				"event_id", event.Id(),
				"reason", err,
			)
			continue
		}

		f1Sessions = append(f1Sessions, session)
	}

	slices.SortFunc(f1Sessions, func(a, b Session) int {
		return a.StartTime.Compare(b.StartTime)
	})

	return f1Sessions
}

func eventToSession(event *ics.VEvent) (Session, error) {
	summary := UnknownSession
	summaryProperty := event.GetProperty(ics.ComponentPropertySummary)
	if summaryProperty != nil && summaryProperty.Value != "" {
		summary = summaryProperty.Value
	}

	name := sessionRegex.FindString(summary)
	sessionType, ok := sessionTypes[name]
	if !ok {
		sessionType = Unknown
	}

	start, err := event.GetStartAt()
	if err != nil {
		return Session{}, fmt.Errorf("failed to parse DTSTART: %w", err)
	}

	end, err := event.GetEndAt()
	if err != nil {
		var duration time.Duration

		switch sessionType {
		case Practice1, Practice2, Practice3, Qualifying, SprintRace:
			duration = time.Hour
		case SprintQualification:
			duration = 44 * time.Minute
		case Race:
			duration = 2 * time.Hour
		default:
			duration = time.Hour
		}

		end = start.Add(duration)
	}

	location := UnknownLocation
	locationProperty := event.GetProperty(ics.ComponentPropertyLocation)
	if locationProperty != nil && locationProperty.Value != "" {
		location = strings.TrimSpace(locationProperty.Value)
	}

	return Session{
		UID:       event.Id(),
		Title:     summary,
		Type:      sessionType,
		Location:  location,
		StartTime: start,
		EndTime:   end,
	}, nil
}

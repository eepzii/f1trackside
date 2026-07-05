package calendar

import (
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/google/go-cmp/cmp"
)

func TestNewSessions(t *testing.T) {

	tests := []struct {
		name  string
		input func() (*ics.Calendar, *slog.Logger)
		want  []Session
	}{
		{
			name: "good new sessions",
			input: func() (*ics.Calendar, *slog.Logger) {
				cal := ics.NewCalendar()

				schedule := []struct {
					key       string
					startTime time.Time
					endTime   time.Time
				}{
					{
						key:       "Race",
						startTime: time.Date(2026, time.July, 3, 12, 0, 0, 0, time.UTC),
						endTime:   time.Date(2026, time.July, 3, 14, 0, 0, 0, time.UTC),
					},
					{
						key:       "Sprint Qualification",
						startTime: time.Date(2026, time.July, 1, 14, 30, 0, 0, time.UTC),
						endTime:   time.Date(2026, time.July, 1, 15, 14, 0, 0, time.UTC),
					},
					{
						key:       "Practice 1",
						startTime: time.Date(2026, time.July, 1, 10, 30, 0, 0, time.UTC),
						endTime:   time.Date(2026, time.July, 1, 11, 30, 0, 0, time.UTC),
					},
					{
						key:       "Sprint Race",
						startTime: time.Date(2026, time.July, 2, 10, 0, 0, 0, time.UTC),
						endTime:   time.Date(2026, time.July, 2, 11, 0, 0, 0, time.UTC),
					},
					{
						key:       "Qualifying",
						startTime: time.Date(2026, time.July, 2, 14, 0, 0, 0, time.UTC),
						endTime:   time.Date(2026, time.July, 2, 15, 0, 0, 0, time.UTC),
					},
				}

				for _, session := range schedule {
					event := cal.AddEvent(session.key)
					event.SetSummary(fmt.Sprintf("Home Grand Prix - %s", session.key))
					event.SetLocation("Home")

					event.SetStartAt(session.startTime)
					event.SetEndAt(session.endTime)
				}

				return cal, nil
			},
			want: []Session{
				{
					UID:       "Practice 1",
					Title:     "Home Grand Prix - Practice 1",
					Type:      Practice1,
					Location:  "Home",
					StartTime: time.Date(2026, time.July, 1, 10, 30, 0, 0, time.UTC),
					EndTime:   time.Date(2026, time.July, 1, 11, 30, 0, 0, time.UTC),
				},
				{
					UID:       "Sprint Qualification",
					Title:     "Home Grand Prix - Sprint Qualification",
					Type:      SprintQualification,
					Location:  "Home",
					StartTime: time.Date(2026, time.July, 1, 14, 30, 0, 0, time.UTC),
					EndTime:   time.Date(2026, time.July, 1, 15, 14, 0, 0, time.UTC),
				},
				{
					UID:       "Sprint Race",
					Title:     "Home Grand Prix - Sprint Race",
					Type:      SprintRace,
					Location:  "Home",
					StartTime: time.Date(2026, time.July, 2, 10, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2026, time.July, 2, 11, 0, 0, 0, time.UTC),
				},
				{
					UID:       "Qualifying",
					Title:     "Home Grand Prix - Qualifying",
					Type:      Qualifying,
					Location:  "Home",
					StartTime: time.Date(2026, time.July, 2, 14, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2026, time.July, 2, 15, 0, 0, 0, time.UTC),
				},
				{
					UID:       "Race",
					Title:     "Home Grand Prix - Race",
					Type:      Race,
					Location:  "Home",
					StartTime: time.Date(2026, time.July, 3, 12, 00, 0, 0, time.UTC),
					EndTime:   time.Date(2026, time.July, 3, 14, 00, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "empty calendar",
			input: func() (*ics.Calendar, *slog.Logger) {
				cal := ics.NewCalendar()
				return cal, nil
			},
			want: []Session{},
		},
		{
			name: "valid and invalid events",
			input: func() (*ics.Calendar, *slog.Logger) {
				cal := ics.NewCalendar()

				validEvent := cal.AddEvent("valid")
				validEvent.SetStartAt(time.Date(2026, time.July, 1, 12, 00, 0, 0, time.UTC))
				validEvent.SetEndAt(time.Date(2026, time.July, 1, 13, 00, 0, 0, time.UTC))

				invalidEvent := cal.AddEvent("invalid")
				invalidEvent.SetSummary("invalid Event")

				return cal, slog.New(slog.DiscardHandler)
			},
			want: []Session{
				{
					UID:       "valid",
					Title:     UnknownSession,
					Type:      Unknown,
					Location:  UnknownLocation,
					StartTime: time.Date(2026, time.July, 1, 12, 00, 0, 0, time.UTC),
					EndTime:   time.Date(2026, time.July, 1, 13, 00, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "nil calendar",
			input: func() (*ics.Calendar, *slog.Logger) {
				return nil, slog.Default()
			},
			want: []Session{},
		},
		{
			name: "nil logger",
			input: func() (*ics.Calendar, *slog.Logger) {
				cal := ics.NewCalendar()

				event := cal.AddEvent("Race")
				event.SetSummary("Home Grand Prix - Race")
				event.SetLocation("Home")
				event.SetStartAt(time.Date(2026, time.July, 3, 12, 00, 0, 0, time.UTC))
				event.SetEndAt(time.Date(2026, time.July, 3, 14, 00, 0, 0, time.UTC))

				return cal, nil
			},
			want: []Session{
				{
					UID:       "Race",
					Title:     "Home Grand Prix - Race",
					Type:      Race,
					Location:  "Home",
					StartTime: time.Date(2026, time.July, 3, 12, 00, 0, 0, time.UTC),
					EndTime:   time.Date(2026, time.July, 3, 14, 00, 0, 0, time.UTC),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cal, logger := test.input()

			sessions := newSessions(cal, logger)

			if diff := cmp.Diff(test.want, sessions); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}

}

func TestEventToSession(t *testing.T) {

	tests := []struct {
		name     string
		input    func() *ics.VEvent
		checkErr func(t *testing.T, err error)
		want     Session
	}{
		{
			name: "good event to session",
			input: func() *ics.VEvent {
				event := ics.NewEvent("abc-123")
				event.SetSummary("Home Grand Prix - Practice 1")
				event.SetLocation("Home")

				startTime := time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC)
				event.SetStartAt(startTime)
				event.SetEndAt(startTime.Add(2 * time.Hour))

				return event
			},
			want: Session{
				UID:       "abc-123",
				Title:     "Home Grand Prix - Practice 1",
				Type:      Practice1,
				Location:  "Home",
				StartTime: time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2026, time.July, 1, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "trims whitespace from location",
			input: func() *ics.VEvent {
				event := ics.NewEvent("abc-123")
				event.SetStartAt(time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC))
				event.SetLocation("   At Home   ")
				return event
			},
			want: Session{
				UID:       "abc-123",
				Title:     UnknownSession,
				Type:      Unknown,
				Location:  "At Home",
				StartTime: time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2026, time.July, 1, 13, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "empty summary and location",
			input: func() *ics.VEvent {
				event := ics.NewEvent("abc-123")
				event.SetSummary("")
				event.SetLocation("")

				startTime := time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC)
				event.SetStartAt(startTime)

				return event
			},
			want: Session{
				UID:       "abc-123",
				Title:     UnknownSession,
				Type:      Unknown,
				Location:  UnknownLocation,
				StartTime: time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2026, time.July, 1, 13, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "missing optional properties",
			input: func() *ics.VEvent {
				event := ics.NewEvent("abc-123")

				startTime := time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC)
				event.SetStartAt(startTime)

				return event
			},
			want: Session{
				UID:       "abc-123",
				Title:     UnknownSession,
				Type:      Unknown,
				Location:  UnknownLocation,
				StartTime: time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2026, time.July, 1, 13, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "missing DTSTART",
			input: func() *ics.VEvent {
				return ics.NewEvent("")
			},
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", ics.ErrorPropertyNotFound)
				}

				if !errors.Is(err, ics.ErrorPropertyNotFound) {
					t.Fatalf("expected %v, got %v", ics.ErrorPropertyNotFound, err)
				}
			},
		},
		{
			name: "nil event",
			input: func() *ics.VEvent {
				return nil
			},
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", errInvalidInput)
				}

				if !errors.Is(err, errInvalidInput) {
					t.Fatalf("expected %v, got %v", errInvalidInput, err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			session, err := eventToSession(test.input())

			if test.checkErr != nil {
				test.checkErr(t, err)
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(test.want, session); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestEventToSession_MissingDTEND(t *testing.T) {

	baseTime := time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		summary  string
		wantType SessionType
		wantTime time.Time
	}{
		{
			name:     "Unknown",
			summary:  "unknown",
			wantType: Unknown,
			wantTime: baseTime.Add(time.Hour),
		},
		{
			name:     "Practice 1",
			summary:  "Practice 1",
			wantType: Practice1,
			wantTime: baseTime.Add(time.Hour),
		},
		{
			name:     "Practice 2",
			summary:  "Practice 2",
			wantType: Practice2,
			wantTime: baseTime.Add(time.Hour),
		},
		{
			name:     "Practice 3",
			summary:  "Practice 3",
			wantType: Practice3,
			wantTime: baseTime.Add(time.Hour),
		},
		{
			name:     "Qualifying",
			summary:  "Qualifying",
			wantType: Qualifying,
			wantTime: baseTime.Add(time.Hour),
		},
		{
			name:     "Sprint Race",
			summary:  "Sprint Race",
			wantType: SprintRace,
			wantTime: baseTime.Add(time.Hour),
		},
		{
			name:     "Sprint Qualification",
			summary:  "Sprint Qualification",
			wantType: SprintQualification,
			wantTime: baseTime.Add(44 * time.Minute),
		},
		{
			name:     "Race",
			summary:  "Race",
			wantType: Race,
			wantTime: baseTime.Add(2 * time.Hour),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			event := ics.NewEvent("abc-123")
			event.SetSummary(test.summary)
			event.SetStartAt(baseTime)

			session, err := eventToSession(event)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if session.Type != test.wantType {
				t.Errorf("expected session type %v, got %v", test.wantType, session.Type)
			}

			if diff := cmp.Diff(test.wantTime, session.EndTime); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

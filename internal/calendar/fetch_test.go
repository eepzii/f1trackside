package calendar

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

type mockRoundTripper struct {
	roundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestFetchCalendar(t *testing.T) {

	tests := []struct {
		name      string
		transport *mockRoundTripper
		ctxSetup  func() (context.Context, context.CancelFunc)
		checkErr  func(t *testing.T, err error)
		want      []Session
	}{
		{
			name: "good fetch calendar",
			transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					calendar := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nBEGIN:VEVENT\r\nDTSTART:20260701T120000Z\r\nSUMMARY:Practice 1\r\nEND:VEVENT\r\nBEGIN:VEVENT\r\nDTSTART:20260701T150000Z\r\nSUMMARY:Practice 2\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"

					header := make(http.Header)
					header.Set("Cache-Control", "max-age=3600")

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(calendar)),
						Header:     header,
					}, nil
				},
			},
			ctxSetup: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Second)
			},
			want: []Session{
				{
					Title:     "Practice 1",
					Type:      Practice1,
					Location:  "Unknown Location",
					StartTime: time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2026, time.July, 1, 13, 0, 0, 0, time.UTC),
				},
				{
					Title:     "Practice 2",
					Type:      Practice2,
					Location:  "Unknown Location",
					StartTime: time.Date(2026, time.July, 1, 15, 0, 0, 0, time.UTC),
					EndTime:   time.Date(2026, time.July, 1, 16, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "empty calendar",
			transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					calendar := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nEND:VCALENDAR\r\n"
					header := make(http.Header)
					header.Set("Cache-Control", "max-age=3600")

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(calendar)),
						Header:     header,
					}, nil
				},
			},
			ctxSetup: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Second)
			},
			want: []Session{},
		},
		{
			name: "nil context",
			ctxSetup: func() (context.Context, context.CancelFunc) {
				_, cancel := context.WithCancel(context.Background())
				return nil, cancel
			},
			transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					calendar := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nBEGIN:VEVENT\r\nDTSTART:20260701T120000Z\r\nSUMMARY:Practice 1\r\nEND:VEVENT\r\nBEGIN:VEVENT\r\nDTSTART:20260701T150000Z\r\nSUMMARY:Practice 2\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"

					header := make(http.Header)
					header.Set("Cache-Control", "max-age=3600")

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(calendar)),
						Header:     header,
					}, nil
				},
			},
			checkErr: func(t *testing.T, err error) {
				errString := "net/http: nil Context"

				if err == nil {
					t.Fatalf("expected %s, got nil", errString)
				}

				if !strings.Contains(err.Error(), errString) {
					t.Fatalf("expected %s, got %v", errString, err)
				}
			},
		},
		{
			name: "context canceled",
			transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					if err := req.Context().Err(); err != nil {
						return nil, err
					}

					calendar := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nBEGIN:VEVENT\r\nDTSTART:20260701T120000Z\r\nSUMMARY:Practice 1\r\nEND:VEVENT\r\nBEGIN:VEVENT\r\nDTSTART:20260701T150000Z\r\nSUMMARY:Practice 2\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"

					header := make(http.Header)
					header.Set("Cache-Control", "max-age=3600")

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(calendar)),
						Header:     header,
					}, nil
				},
			},
			ctxSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, cancel
			},
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", context.Canceled)
				}

				if !errors.Is(err, context.Canceled) {
					t.Fatalf("expected %v, got %v", context.Canceled, err)
				}
			},
		},
		{
			name: "server returns 500 status",
			ctxSetup: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Second)
			},
			transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					calendar := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nBEGIN:VEVENT\r\nDTSTART:20260701T120000Z\r\nSUMMARY:Practice 1\r\nEND:VEVENT\r\nBEGIN:VEVENT\r\nDTSTART:20260701T150000Z\r\nSUMMARY:Practice 2\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"

					header := make(http.Header)
					header.Set("Cache-Control", "max-age=3600")

					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(strings.NewReader(calendar)),
						Header:     header,
					}, nil
				},
			},
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", errCalendarFetch)
				}

				if !errors.Is(err, errCalendarFetch) {
					t.Fatalf("expected %v, got %v", errCalendarFetch, err)
				}
			},
		},
		{
			name: "ics parser error",
			transport: &mockRoundTripper{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					calendar := "BE:VCALENDARENDVCALENDAR\r\n"

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(calendar)),
						Header:     make(http.Header),
					}, nil
				},
			},
			ctxSetup: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Second)
			},
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", errParseICS)
				}

				if !errors.Is(err, errParseICS) {
					t.Fatalf("expected %v, got %v", errParseICS, err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.transport == nil {
				t.Fatal("transport not provided")
			}

			client := &Client{
				httpClient: &http.Client{
					Transport: test.transport,
				},
				logger: slog.Default(),
			}

			ctx, cancel := test.ctxSetup()
			defer cancel()

			sessions, err := client.FetchCalendar(ctx)

			if test.checkErr != nil {
				test.checkErr(t, err)
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(test.want, sessions); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFetchCalendar_Cache(t *testing.T) {
	var httpCalls atomic.Int32

	mockTransport := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			httpCalls.Add(1)
			calendar := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nEND:VCALENDAR\r\n"
			header := make(http.Header)
			header.Set("Cache-Control", "max-age=3600")

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(calendar)),
				Header:     header,
			}, nil
		},
	}

	client := &Client{
		httpClient: &http.Client{Transport: mockTransport},
		logger:     slog.Default(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.FetchCalendar(ctx)
	if err != nil {
		t.Fatalf("first fetch: unexpected error: %v", err)
	}
	if calls := httpCalls.Load(); calls != 1 {
		t.Fatalf("first call: expected HTTP call, got %d", calls)
	}

	_, err = client.FetchCalendar(ctx)
	if err != nil {
		t.Fatalf("second fetch: unexpected error: %v", err)
	}
	if calls := httpCalls.Load(); calls != 1 {
		t.Fatalf("second call: expected cache hit, but got %d calls", calls)
	}

	client.mu.Lock()
	client.expiresAt = time.Now().Add(-1 * time.Minute)
	client.mu.Unlock()

	_, err = client.FetchCalendar(ctx)
	if err != nil {
		t.Fatalf("third fetch: unexpected error: %v", err)
	}
	if calls := httpCalls.Load(); calls != 2 {
		t.Fatalf("third call: expected cache miss, but got %d calls", calls)
	}
}

func TestFetchCalendar_Concurrency(t *testing.T) {
	var httpCalls atomic.Int32

	mockTransport := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			httpCalls.Add(1)

			time.Sleep(50 * time.Millisecond)

			calendar := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nBEGIN:VEVENT\r\nDTSTART:20260701T120000Z\r\nSUMMARY:Practice 1\r\nEND:VEVENT\r\nEND:VCALENDAR\r\n"

			header := make(http.Header)
			header.Set("Cache-Control", "max-age=3600")

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(calendar)),
				Header:     header,
			}, nil
		},
	}

	client := &Client{
		httpClient: &http.Client{
			Transport: mockTransport,
		},
		logger: slog.Default(),
	}

	routines := 1000
	start := make(chan struct{})
	var wg sync.WaitGroup

	for range routines {
		wg.Go(func() {
			<-start

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			_, err := client.FetchCalendar(ctx)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}

	close(start)
	wg.Wait()

	if calls := httpCalls.Load(); calls != 1 {
		t.Errorf("double-checked lock failed: expected 1 HTTP call, got %d", calls)
	}
}

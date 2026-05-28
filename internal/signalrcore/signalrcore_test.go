package signalrcore

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/gorilla/websocket"
)

func TestNew(t *testing.T) {

	httpURL, err := url.Parse("http://hello.world")
	if err != nil {
		t.Fatalf("test aborted: could not parse url %v", err)
	}

	wsURL, err := url.Parse("ws://hello.world")
	if err != nil {
		t.Fatalf("test aborted: could not parse url %v", err)
	}

	httpsURL, err := url.Parse("https://hello.world")
	if err != nil {
		t.Fatalf("test aborted: could not parse url %v", err)
	}

	wssURL, err := url.Parse("wss://hello.world")
	if err != nil {
		t.Fatalf("test aborted: could not parse url %v", err)
	}

	tests := []struct {
		name     string
		baseURL  string
		config   *Config
		want     *Client
		checkErr func(t *testing.T, err error)
	}{
		{
			name:    "good new",
			baseURL: "http://hello.world",
			config: &Config{
				Client:      http.DefaultClient,
				Dialer:      websocket.DefaultDialer,
				Logger:      nil,
				IdleTimeout: 0,
				Token:       "",
			},
			want: &Client{
				baseURL:      httpURL,
				websocketURL: wsURL,

				eventChan:   make(map[string]chan Event),
				pendingChan: make(map[string]chan Message),
				doneChan:    make(chan struct{}),

				client:      http.DefaultClient,
				dialer:      websocket.DefaultDialer,
				idleTimeout: 45 * time.Second,
				token:       "",
			},
		},
		{
			name:    "custom config overrides",
			baseURL: "https://hello.world",
			config: &Config{
				Client:      http.DefaultClient,
				Dialer:      websocket.DefaultDialer,
				Logger:      slog.Default(),
				IdleTimeout: 10 * time.Second,
				Token:       "jwt",
			},
			want: &Client{
				baseURL:      httpsURL,
				websocketURL: wssURL,

				eventChan:   make(map[string]chan Event),
				pendingChan: make(map[string]chan Message),
				doneChan:    make(chan struct{}),

				client:      http.DefaultClient,
				dialer:      websocket.DefaultDialer,
				idleTimeout: 10 * time.Second,
				token:       "jwt",
			},
		},
		{
			name:    "nil config",
			baseURL: "http://hello.world",
			config:  nil,
			want:    nil,
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", errInvalidInput)
				}

				if !errors.Is(err, errInvalidInput) {
					t.Fatalf("expected %v, got %v", errInvalidInput, err)
				}
			},
		},
		{
			name:    "parse url",
			baseURL: "http://hello.world/%zz",
			config: &Config{
				Client: http.DefaultClient,
				Dialer: websocket.DefaultDialer,
			},
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatal(`expected invalid URL escape, got nil`)
				}

				var urlErr *url.Error
				if !errors.As(err, &urlErr) || urlErr.URL != "http://hello.world/%zz" {
					t.Fatalf(`expected invalid URL escape, got %v`, err)
				}
			},
		},
		{
			name:    "invalid url scheme",
			baseURL: "tcp://hello.world",
			config: &Config{
				Client: http.DefaultClient,
				Dialer: websocket.DefaultDialer,
			},
			want: nil,
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", errInvalidInput)
				}

				if !errors.Is(err, errInvalidInput) {
					t.Fatalf("expected %v, got %v", errInvalidInput, err)
				}
			},
		},
		{
			name:    "nil http client",
			baseURL: "http://hello.world",
			config: &Config{
				Client: nil,
				Dialer: websocket.DefaultDialer,
			},
			want: nil,
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", errInvalidConfig)
				}

				if !errors.Is(err, errInvalidConfig) {
					t.Fatalf("expected %v, got %v", errInvalidConfig, err)
				}
			},
		},
		{
			name:    "nil websocket dialer",
			baseURL: "http://hello.world",
			config: &Config{
				Client: http.DefaultClient,
				Dialer: nil,
			},
			want: nil,
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", errInvalidConfig)
				}

				if !errors.Is(err, errInvalidConfig) {
					t.Fatalf("expected %v, got %v", errInvalidConfig, err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, err := New(test.baseURL, test.config)

			if test.checkErr != nil {
				test.checkErr(t, err)
			} else if err != nil {
				t.Fatalf("unexpected error %v", err)
			}

			if test.want != nil {
				if test.config == nil {
					t.Fatalf("config cannot be nil")
				}

				if test.config.Logger != nil {
					if client.logger != test.config.Logger {
						t.Fatalf("expected custom logger %p, got %p", test.config.Logger, client.logger)
					}
				} else {
					if client.logger == nil {
						t.Fatal("expected default logger to be initialized, got nil")
					}
				}

				if client.eventChan == nil {
					t.Fatalf("event map cannot be nil")
				}

				if client.pendingChan == nil {
					t.Fatalf("pending map cannot be nil")
				}

				if client.doneChan == nil {
					t.Fatalf("done channel cannot be nil")
				}
			}

			opts := []cmp.Option{
				cmp.AllowUnexported(Client{}, atomic.Uint64{}, atomic.Uint32{}),
				cmpopts.IgnoreTypes(sync.Mutex{}),
				cmpopts.IgnoreFields(websocket.Dialer{}, "Proxy"),
				cmpopts.IgnoreFields(Client{}, "logger"),
				cmp.Comparer(func(x, y chan struct{}) bool {
					if x == nil && y == nil {
						return true
					}

					if x == nil || y == nil {
						return false
					}

					return cap(x) == cap(y)
				}),
			}

			if diff := cmp.Diff(test.want, client, opts...); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

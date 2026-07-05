package calendar

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestNewClient(t *testing.T) {

	tests := []struct {
		name     string
		input    func() (*http.Client, *slog.Logger)
		checkErr func(t *testing.T, err error)
		want     *Client
	}{
		{
			name: "good new client",
			input: func() (*http.Client, *slog.Logger) {
				return http.DefaultClient, slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
					Level: slog.LevelError,
				}))
			},
			want: &Client{
				httpClient: http.DefaultClient,
			},
		},
		{
			name: "nil http client",
			input: func() (*http.Client, *slog.Logger) {
				return nil, slog.Default()
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
		{
			name: "nil logger",
			input: func() (*http.Client, *slog.Logger) {
				return http.DefaultClient, nil
			},
			want: &Client{
				httpClient: http.DefaultClient,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpClient, logger := test.input()

			client, err := NewClient(httpClient, logger)

			if test.checkErr != nil {
				test.checkErr(t, err)
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if test.want != nil {
				if logger != nil {
					if client.logger != logger {
						t.Fatalf("expected custom logger %p, got %p", logger, client.logger)
					}
				} else {
					if client.logger == nil {
						t.Fatal("expected default logger to be initialized, got nil")
					}
				}

				client.logger.Info("test log to ensure logger initialization")
			}

			opts := []cmp.Option{
				cmp.AllowUnexported(Client{}),
				cmpopts.IgnoreTypes(sync.RWMutex{}),
				cmpopts.IgnoreFields(Client{}, "logger"),
			}
			if diff := cmp.Diff(test.want, client, opts...); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

package signalrcore

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestNegotiate(t *testing.T) {

	tests := []struct {
		name        string
		ctxSetup    func() (context.Context, context.CancelFunc)
		serverSetup func(t *testing.T) (*httptest.Server, url.URL)
		checkErr    func(t *testing.T, err error)
		want        negotiation
	}{
		{
			name: "good negotiation",
			ctxSetup: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Second)
			},
			serverSetup: func(t *testing.T) (*httptest.Server, url.URL) {
				response := []byte(`{"negotiateVersion":0,"connectionId":"abc1234","availableTransports":[{"transport":"WebSockets","transferFormats":["Text","Binary"]},{"transport":"ServerSentEvents","transferFormats":["Text"]},{"transport":"LongPolling","transferFormats":["Text","Binary"]}]}`)
				cookies := []*http.Cookie{
					{
						Name:  "Cookie1",
						Value: "coo/kie/1",
					},
					{
						Name:  "Cookie2",
						Value: "coo/kie/2",
					},
				}

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/negotiate" {
						t.Fatalf("negotiate needs to go to url with path: /negotiate")
					}
					for _, cookie := range cookies {
						http.SetCookie(w, cookie)
					}
					w.Write(response)
				}))

				httpURL, err := url.Parse(server.URL)
				if err != nil {
					t.Fatalf("could not parse server URL: %v", err)
				}

				return server, *httpURL
			},
			want: negotiation{
				body: negotiationBody{
					ConnectionID: "abc1234",
					AvailableTransports: []struct {
						Transport       string   `json:"transport"`
						TransferFormats []string `json:"transferFormats"`
					}{
						{
							Transport:       "WebSockets",
							TransferFormats: []string{"Text", "Binary"},
						},
						{
							Transport:       "ServerSentEvents",
							TransferFormats: []string{"Text"},
						},
						{
							Transport:       "LongPolling",
							TransferFormats: []string{"Text", "Binary"},
						},
					},
					NegotiateVersion: 0,
				},
				cookies: []*http.Cookie{
					{
						Name:  "Cookie1",
						Value: "coo/kie/1",
						Raw:   "Cookie1=coo/kie/1",
					},
					{
						Name:  "Cookie2",
						Value: "coo/kie/2",
						Raw:   "Cookie2=coo/kie/2",
					},
				},
			},
		},
		{
			name: "nil context",
			ctxSetup: func() (context.Context, context.CancelFunc) {
				_, cancel := context.WithCancel(context.Background())
				return nil, cancel
			},
			serverSetup: func(t *testing.T) (*httptest.Server, url.URL) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				return server, url.URL{}
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
			name: "canceled context",
			ctxSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, cancel
			},
			serverSetup: func(t *testing.T) (*httptest.Server, url.URL) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

				httpURL, err := url.Parse(server.URL)
				if err != nil {
					t.Fatalf("could not parse server URL: %v", err)
				}

				return server, *httpURL
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
			serverSetup: func(t *testing.T) (*httptest.Server, url.URL) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))

				httpURL, err := url.Parse(server.URL)
				if err != nil {
					t.Fatalf("could not parse server URL: %v", err)
				}

				return server, *httpURL
			},
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatalf("expected %v, got nil", errNegotiation)
				}

				if !errors.Is(err, errNegotiation) {
					t.Fatalf("expected %v, got %v", errNegotiation, err)
				}
			},
		},
		{
			name: "malformed JSON response",
			ctxSetup: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 3*time.Second)
			},
			serverSetup: func(t *testing.T) (*httptest.Server, url.URL) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte(`{something:123}`))
				}))

				httpURL, err := url.Parse(server.URL)
				if err != nil {
					t.Fatalf("could not parse server URL: %v", err)
				}

				return server, *httpURL
			},
			checkErr: func(t *testing.T, err error) {
				if err == nil {
					t.Fatal("expected JSON syntax error, got nil")
				}

				var jsonErr *json.SyntaxError
				if !errors.As(err, &jsonErr) {
					t.Fatalf("expected JSON syntax error, got type %T: %v", err, err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, httpURL := test.serverSetup(t)
			defer server.Close()

			client := Client{
				client: server.Client(),
			}

			ctx, cancel := test.ctxSetup()
			defer cancel()

			negotiationRes, err := client.negotiate(ctx, httpURL)

			if test.checkErr != nil {
				test.checkErr(t, err)
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			opts := cmp.AllowUnexported(negotiation{}, negotiationBody{})
			if diff := cmp.Diff(test.want, negotiationRes, opts); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}

		})
	}
}

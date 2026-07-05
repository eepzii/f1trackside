package calendar

import (
	"testing"
	"time"
)

func TestParseCacheControl(t *testing.T) {

	tests := []struct {
		name  string
		input string
		want  time.Duration
	}{
		{
			name:  "good max-age",
			input: "max-age=3600, private, must-revalidate",
			want:  time.Hour,
		},
		{
			name:  "empty header",
			input: "",
			want:  2 * time.Hour,
		},
		{
			name:  "max-age among other directives",
			input: "public, max-age=1800",
			want:  30 * time.Minute,
		},
		{
			name:  "missing max-age",
			input: "private, must-revalidate",
			want:  2 * time.Hour,
		},
		{
			name:  "missing max-age value",
			input: "max-age=",
			want:  2 * time.Hour,
		},
		{
			name:  "malformed max-age value",
			input: "max-age=;thea,max-gae=120",
			want:  2 * time.Hour,
		},
		{
			name:  "preceding invalid max-age values",
			input: "max-age=;thea,max-age=abc,max-age=60",
			want:  time.Minute,
		},
		{
			name:  "missing commas between max-age directives",
			input: "max-age=180max-age=120max-age=60",
			want:  2 * time.Hour,
		},
		{
			name:  "ignores case and whitespace",
			input: " Max-Age=3600 ",
			want:  time.Hour,
		},
		{
			name:  "zero max-age value",
			input: "max-age=0",
			want:  0,
		},
		{
			name:  "negative max-age value",
			input: "max-age=-1",
			want:  -1 * time.Second,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			duration := parseCacheDuration(test.input)

			if duration != test.want {
				t.Fatalf("expected %s, got %s", test.want, duration)
			}
		})
	}
}

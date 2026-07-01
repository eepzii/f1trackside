package calendar

import (
	"context"
	"fmt"
	"net/http"
	"time"

	ics "github.com/arran4/golang-ical"
)

func (c *Client) FetchCalendar(ctx context.Context) ([]Session, error) {
	c.mu.RLock()
	expiresAt := c.expiresAt
	cachedSessions := c.sessions
	c.mu.RUnlock()

	if time.Now().Before(expiresAt) && cachedSessions != nil {
		return cachedSessions, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().Before(c.expiresAt) && c.sessions != nil {
		return c.sessions, nil
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		calendarURL,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar fetch request: %w", err)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calendar fetch request failed: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server rejected calendar fetch request with status %s: %w", res.Status, errCalendarFetch)
	}

	ccHeader := res.Header.Get(cacheControl)
	maxAge := parseCacheDuration(ccHeader)

	cal, err := ics.ParseCalendar(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errParseICS, err)
	}

	c.expiresAt = time.Now().Add(maxAge)
	c.sessions = newSessions(cal, c.logger)

	return c.sessions, nil
}

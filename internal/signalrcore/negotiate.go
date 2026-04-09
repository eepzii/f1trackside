package signalrcore

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

func negotiate(url url.URL) (negotiationRes, error) {
	var res = negotiationRes{}

	postReq, err := http.NewRequest(
		http.MethodPost,
		url.String()+"/negotiate",
		nil,
	)
	if err != nil {
		return res, err
	}

	negotiationRes, err := http.DefaultClient.Do(postReq)
	if err != nil {
		return res, err
	}

	body, err := io.ReadAll(negotiationRes.Body)
	if err != nil {
		return res, err
	}

	if err := json.Unmarshal(body, &res.body); err != nil {
		return res, err
	}
	negotiationRes.Body.Close()
	res.cookies = append(res.cookies, negotiationRes.Cookies()...)

	return res, nil
}

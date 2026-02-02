package types

import (
	"encoding/json"
	"testing"
)

// TODO: apply some refactor changes of TestParseTimingData here as well.
func TestParseSessionInfo(t *testing.T) {

	jsonInput := []byte(`{
        "SessionInfo":{
            "Meeting":{
                "Key":1274,
                "Name":"Las Vegas Grand Prix",
                "OfficialName":"FORMULA 1 HEINEKEN LAS VEGAS GRAND PRIX 2025",
                "Location":"Las Vegas",
                "Number":22,
                "Country":{
                    "Key":19,
                    "Code":"USA",
                    "Name":"United States"
                },
                "Circuit":{
                    "Key":152,
                    "ShortName":"Las Vegas"
                }
            },
            "SessionStatus":"Inactive",
            "ArchiveStatus":{
                "Status":"Generating"
            },
            "Key":9858,
            "Type":"Race",
            "Name":"Race",
            "StartDate":"2025-11-22T20:00:00",
            "EndDate":"2025-11-22T22:00:00",
            "GmtOffset":"-08:00:00",
            "Path":"2025/2025-11-22_Las_Vegas_Grand_Prix/2025-11-22_Race/"
        }
    }`)

	var response SessionInfoResponse
	if err := json.Unmarshal(jsonInput, &response); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	test := struct {
		ExpectedSessionKey        int
		ExpectedMeetingKey        int
		ExpectedMeetingCircuitKey int
		ExpectedSessionType       string
		ExpectedMeetingName       string
		ExpectedArchiveStatus     string
	}{
		ExpectedSessionKey:        9858,
		ExpectedMeetingKey:        1274,
		ExpectedMeetingCircuitKey: 152,
		ExpectedSessionType:       "Race",
		ExpectedMeetingName:       "Las Vegas Grand Prix",
		ExpectedArchiveStatus:     "Generating",
	}

	if response.SessionInfo.Key != test.ExpectedSessionKey {
		t.Errorf(`expected .Key "%d", got: "%d"`,
			test.ExpectedSessionKey, response.SessionInfo.Key)
	}

	if response.SessionInfo.Meeting.Key != test.ExpectedMeetingKey {
		t.Errorf(`expected .Meeting.Key "%d", got: "%d"`,
			test.ExpectedMeetingKey, response.SessionInfo.Meeting.Key)
	}

	if response.SessionInfo.Meeting.Circuit.Key != test.ExpectedMeetingCircuitKey {
		t.Errorf(`expected .Meeting.Circuit.Key "%d", got "%d"`,
			test.ExpectedMeetingCircuitKey, response.SessionInfo.Meeting.Circuit.Key)
	}

	if response.SessionInfo.Type != test.ExpectedSessionType {
		t.Errorf(`expected .Type "%s", got: "%s"`,
			test.ExpectedSessionType, response.SessionInfo.Type)
	}

	if response.SessionInfo.Meeting.Name != test.ExpectedMeetingName {
		t.Errorf(`expected .Meeting.Name "%s", got: "%s"`,
			test.ExpectedMeetingName, response.SessionInfo.Meeting.Name)
	}

	if response.SessionInfo.ArchiveStatus.Status != test.ExpectedArchiveStatus {
		t.Errorf(`expected .ArchiveStatus.Status "%s", got: "%s"`,
			test.ExpectedArchiveStatus, response.SessionInfo.ArchiveStatus.Status)
	}

}

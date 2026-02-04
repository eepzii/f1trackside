package types

import (
	"encoding/json"
	"testing"
)

// TODO: refactor test table
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

	err := validateValues(response.SessionInfo.Key, test.ExpectedSessionKey)
	if err != nil {
		t.Errorf(".SessionInfo.Key -> %v", err)
	}

	err = validateValues(response.SessionInfo.Meeting.Key, test.ExpectedMeetingKey)
	if err != nil {
		t.Errorf(".SessionInfo.Meeting.Key -> %v", err)
	}

	err = validateValues(response.SessionInfo.Meeting.Circuit.Key, test.ExpectedMeetingCircuitKey)
	if err != nil {
		t.Errorf(".SessionInfo.Meeting.Circuit.Key -> %v", err)
	}

	err = validateValues(response.SessionInfo.Type, test.ExpectedSessionType)
	if err != nil {
		t.Errorf(".SessionInfo.Type -> %v", err)
	}

	err = validateValues(response.SessionInfo.Meeting.Name, test.ExpectedMeetingName)
	if err != nil {
		t.Errorf(".SessionInfo.Meeting.Name -> %v", err)
	}

	err = validateValues(response.SessionInfo.ArchiveStatus.Status, test.ExpectedArchiveStatus)
	if err != nil {
		t.Errorf(".SessionInfo.ArchiveStatus.Status -> %v", err)
	}

}

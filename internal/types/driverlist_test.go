package types

import (
	"encoding/json"
	"testing"
)

// TODO: apply some refactor changes of TestParseTimingData here as well.
func TestParseDriverList(t *testing.T) {

	jsonInput := []byte(`{
		"DriverList":{
		    "4":{
		        "RacingNumber":"4",
		        "BroadcastName":"L NORRIS",
		        "FullName":"Lando NORRIS",
		        "Tla":"NOR",
		        "Line":1,
		        "TeamName":"McLaren",
		        "TeamColour":"F47600",
		        "FirstName":"Lando",
		        "LastName":"Norris",
		        "Reference":"LANNOR01",
		        "HeadshotUrl":"https://media.formula1.com/d_driver_fallback_image.png/content/dam/fom-website/drivers/L/LANNOR01_Lando_Norris/lannor01.png.transform/1col/image.png",
		        "PublicIdRight":"common/f1/2025/mclaren/lannor01/2025mclarenlannor01right"
			},
		    "1":{
		        "Line":2
		    }
		}
	}`)

	var response DriverListResponse
	if err := json.Unmarshal(jsonInput, &response); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	if response.DriverList == nil {
		t.Fatalf(".DriverList is nil")
	}

	tests := []struct {
		DriverNumber          string
		ExpectedBroadcastName string
		ExpectedTeamColor     string
		ExpectedPosition      int
	}{
		{"4", "L NORRIS", "F47600", 1},
		{"1", "", "", 2},
	}

	for _, test := range tests {

		driver, exists := response.DriverList[test.DriverNumber]
		if !exists {
			t.Errorf("missing %s driver from map", test.DriverNumber)
			continue
		}

		if driver.BroadcastName != test.ExpectedBroadcastName {
			t.Errorf(`expected .BroadcastName "%s", got: "%s"`,
				test.ExpectedBroadcastName, driver.BroadcastName)
		}

		if driver.TeamColor != test.ExpectedTeamColor {
			t.Errorf(`expected .TeamColor "%s", get: "%s"`,
				test.ExpectedTeamColor, driver.TeamColor)
		}

		if driver.Line != test.ExpectedPosition {
			t.Errorf(`expected .Line "%d", got: "%d"`,
				test.ExpectedPosition, driver.Line)
		}
	}

}

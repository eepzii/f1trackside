package types

import (
	"encoding/json"
	"testing"
)

// TODO: refactor test table
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
			t.Errorf("missing key on .DriverList: %s", test.DriverNumber)
			continue
		}

		err := validateValues(driver.BroadcastName, test.ExpectedBroadcastName)
		if err != nil {
			t.Errorf(".DriverList[%s].BroadcastName -> %v",
				test.DriverNumber, err)
		}

		err = validateValues(driver.TeamColor, test.ExpectedTeamColor)
		if err != nil {
			t.Errorf(".DriverList[%s].TeamColor -> %v", test.DriverNumber, err)
		}

		err = validateValues(driver.Line, test.ExpectedPosition)
		if err != nil {
			t.Errorf(".DriverList[%s].Line -> %v", test.DriverNumber, err)
		}
	}

}

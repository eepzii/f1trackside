package types

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseDriverList(t *testing.T) {

	tests := []struct {
		name  string
		input []byte
		want  DriverListResponse
	}{
		{
			name: ".DriverList",
			input: []byte(`{
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
					"_kf": true
        		}
        	}`),
			want: DriverListResponse{
				DriverList: DriverList{
					Drivers: map[string]Driver{
						"4": {
							RacingNumber:  "4",
							BroadcastName: "L NORRIS",
							FullName:      "Lando NORRIS",
							Tla:           "NOR",
							Line:          1,
							TeamName:      "McLaren",
							TeamColor:     "F47600",
							FirstName:     "Lando",
							LastName:      "Norris",
							Reference:     "LANNOR01",
							HeadshotURL:   "https://media.formula1.com/d_driver_fallback_image.png/content/dam/fom-website/drivers/L/LANNOR01_Lando_Norris/lannor01.png.transform/1col/image.png",
							PublicIDRight: "common/f1/2025/mclaren/lannor01/2025mclarenlannor01right",
						},
					},
					KeyFrame: true,
				},
			},
		},
		{
			name: "multiple .DriverList.Line value changes",
			input: []byte(`{
                "DriverList":{
                    "30":{ "Line":9 },
                    "16":{ "Line":8 }
                }
            }`),
			want: DriverListResponse{
				DriverList: DriverList{
					Drivers: map[string]Driver{
						"30": {Line: 9},
						"16": {Line: 8},
					},
				},
			},
		},
		{
			name:  "empty .DriverList",
			input: []byte(`{}`),
			want: DriverListResponse{
				DriverList: DriverList{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var got DriverListResponse
			if err := json.Unmarshal(test.input, &got); err != nil {
				t.Fatalf("failed to unmarshal json: %v", err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("mismatch testing %s (-want +got):\n%s", test.name, diff)
			}
		})
	}

}

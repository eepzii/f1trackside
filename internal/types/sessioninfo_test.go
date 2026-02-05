package types

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseSessionInfo(t *testing.T) {

	tests := []struct {
		name  string
		input []byte
		want  SessionInfoResponse
	}{
		{
			name: ".SessionInfo",
			input: []byte(`{
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
            }`),
			want: SessionInfoResponse{
				SessionInfo: SessionInfo{
					Meeting: Meeting{
						Key:          1274,
						Name:         "Las Vegas Grand Prix",
						OfficialName: "FORMULA 1 HEINEKEN LAS VEGAS GRAND PRIX 2025",
						Location:     "Las Vegas",
						Number:       22,
						Country: Country{
							Key:  19,
							Code: "USA",
							Name: "United States",
						},
						Circuit: Circuit{
							Key:       152,
							ShortName: "Las Vegas",
						},
					},
					SessionStatus: "Inactive",
					ArchiveStatus: ArchiveStatus{
						Status: "Generating",
					},
					Key:       9858,
					Type:      "Race",
					Name:      "Race",
					StartDate: "2025-11-22T20:00:00",
					EndDate:   "2025-11-22T22:00:00",
					GmtOffset: "-08:00:00",
					Path:      "2025/2025-11-22_Las_Vegas_Grand_Prix/2025-11-22_Race/",
				},
			},
		},
		{
			name:  "empty .SessionInfo",
			input: []byte(`{"SessionInfo":{}}`),
			want:  SessionInfoResponse{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var got SessionInfoResponse
			if err := json.Unmarshal(test.input, &got); err != nil {
				t.Fatalf("failed to unmarshal json: %v", err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("mismatch testing %s (-want +got):\n%s", test.name, diff)
			}
		})
	}

}

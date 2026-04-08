package types

import (
	"encoding/json"
)

type DriverListResponse struct {
	DriverList DriverList `json:"DriverList"`
}

type DriverList struct {
	Drivers  map[string]Driver `json:"Drivers"`
	KeyFrame bool              `json:"_kf"`
}

type Driver struct {
	BroadcastName string   `json:"BroadcastName"`
	FirstName     string   `json:"FirstName"`
	FullName      string   `json:"FullName"`
	HeadshotURL   string   `json:"HeadshotUrl"`
	LastName      string   `json:"LastName"`
	Line          int      `json:"Line"`
	PublicIDRight string   `json:"PublicIdRight"`
	RacingNumber  string   `json:"RacingNumber"`
	Reference     string   `json:"Reference"`
	TeamColor     string   `json:"TeamColour"`
	TeamName      string   `json:"TeamName"`
	Tla           string   `json:"Tla"`
	Deleted       []string `json:"_deleted"`
}

func (d *DriverList) UnmarshalJSON(data []byte) error {

	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return err
	}

	d.Drivers = make(map[string]Driver)

	for key, entry := range rawMap {

		if key == "_kf" {
			if err := json.Unmarshal(entry, &d.KeyFrame); err != nil {
				return err
			}
			continue
		}

		var driver Driver
		if err := json.Unmarshal(entry, &driver); err != nil {
			return err
		}
		d.Drivers[key] = driver
	}

	return nil
}

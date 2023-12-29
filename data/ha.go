package data

type HAEntity struct {
	ID          string       `json:"entity_id"`
	State       string       `json:"state"`
	Attributes  HAAttributes `json:"attributes"`
	LastChanged string       `json:"last_changed"` // "2023-12-27T15:28:26.287133+00:00"
	LastUpdated string       `json:"last_updated"` // "2023-12-27T15:28:26.287133+00:00"
	Context     HAContext    `json:"context"`
}

type HAAttributes struct {
	ID                string `json:"id"`
	FriendlyName      string `json:"friendly_name"`
	DeviceClass       string `json:"device_class"`
	UnitOfMeasurement string `json:"unit_of_measurement"`
	Icon              string `json:"icon"`
}

type HAContext struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id"`
	UserID   string `json:"user_id"`
}

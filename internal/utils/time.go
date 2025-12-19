package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	d.Duration = parsed
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	if d.Duration < 0 {
		return nil, fmt.Errorf("duration must be > 0")
	}
	return json.Marshal(d.String())
}

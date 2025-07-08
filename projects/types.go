package projects

import (
	"encoding/json"
	"strconv"
	"time"
)

// LegacyDate is a type alias for time.Time, used to represent legacy date
// values in the API.
type LegacyDate time.Time

// MarshalJSON encodes the LegacyDate as a string in the format "20060102".
func (d LegacyDate) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(d).Format("20060102") + `"`), nil
}

// UnmarshalJSON decodes a JSON string into a LegacyDate type.
func (d *LegacyDate) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	parsedTime, err := time.Parse("20060102", str)
	if err != nil {
		return err
	}
	*d = LegacyDate(parsedTime)
	return nil
}

// LegacyNumber is a type alias for int64, used to represent numeric values in
// the API.
type LegacyNumber int64

// MarshalJSON encodes the LegacyNumber as a string.
func (n LegacyNumber) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatInt(int64(n), 10) + `"`), nil
}

// UnmarshalJSON decodes a JSON string into a LegacyNumber type.
func (n *LegacyNumber) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	parsedInt, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return err
	}
	*n = LegacyNumber(parsedInt)
	return nil
}

package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JSONField is a generic JSON column type for GORM
type JSONField map[string]interface{}

func (j JSONField) Value() (driver.Value, error) {
	if j == nil {
		return "null", nil
	}
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (j *JSONField) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into JSONField", value)
	}
	return json.Unmarshal(bytes, j)
}

// JSONSlice is a JSON array column type
type JSONSlice []interface{}

func (j JSONSlice) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (j *JSONSlice) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into JSONSlice", value)
	}
	return json.Unmarshal(bytes, j)
}

// JSONRaw stores any JSON value
type JSONRaw json.RawMessage

func (j JSONRaw) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "null", nil
	}
	return string(j), nil
}

func (j *JSONRaw) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = append((*j)[0:0], v...)
	case string:
		*j = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T into JSONRaw", value)
	}
	return nil
}

func (j JSONRaw) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// Response is the standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func SuccessResp(data interface{}) Response {
	return Response{Success: true, Data: data}
}

func ErrorResp(msg string) Response {
	return Response{Success: false, Error: msg}
}

// APIRequest is the standard request format
type APIRequest struct {
	Order string                 `json:"order"`
	Args  map[string]interface{} `json:"args"`
}

func (r *APIRequest) GetString(key string) string {
	if r.Args == nil {
		return ""
	}
	if v, ok := r.Args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (r *APIRequest) GetInt(key string, def int) int {
	if r.Args == nil {
		return def
	}
	if v, ok := r.Args[key]; ok {
		switch n := v.(type) {
		case int:
			return n
		case float64:
			return int(n)
		case int64:
			return int(n)
		}
	}
	return def
}

func (r *APIRequest) GetBool(key string) bool {
	if r.Args == nil {
		return false
	}
	if v, ok := r.Args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func (r *APIRequest) GetUint(key string) uint {
	if r.Args == nil {
		return 0
	}
	if v, ok := r.Args[key]; ok {
		switch n := v.(type) {
		case float64:
			return uint(n)
		case int:
			return uint(n)
		case uint:
			return n
		}
	}
	return 0
}

// TimePtr returns a pointer to time.Time
func TimePtr(t time.Time) *time.Time {
	return &t
}

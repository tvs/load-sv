package duration

import (
	"encoding/json"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration is a wrapper around time.Duration which supports marshaling to
// YAML and JSON. In particular, it marshals into strings, which can be
// used as map keys in JSON.
type Duration struct {
	time.Duration
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}

	pd, err := time.ParseDuration(str)
	if err != nil {
		return err
	}

	d.Duration = pd
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

// UnmarshalJSON implements the yaml.Unmarshaller interface.
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var str string
	err := value.Decode(&str)
	if err != nil {
		return err
	}

	pd, err := time.ParseDuration(str)
	if err != nil {
		return err
	}

	d.Duration = pd
	return nil
}

// MarshalYAML implements the yaml.Marshaler interface.
func (d Duration) MarshalYAML() (interface{}, error) {
	return d.Duration.String(), nil
}

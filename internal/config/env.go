package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type StringFromEnv string

func (e *StringFromEnv) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}

	*e = StringFromEnv(os.ExpandEnv(s))

	return nil
}

func (e StringFromEnv) String() string {
	return string(e)
}


package config

import (
	"io"
	"os"
	"strings"

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

type StringFromFile string

func (f *StringFromFile) UnmarshalYAML(value *yaml.Node) error {
	var s StringFromEnv
	if err := value.Decode(&s); err != nil {
		return err
	}

	file, err := os.Open(s.String())
	if err != nil {
		return err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	*f = StringFromFile(strings.TrimSpace(string(data)))

	return nil
}

func (f StringFromFile) String() string {
	return string(f)
}

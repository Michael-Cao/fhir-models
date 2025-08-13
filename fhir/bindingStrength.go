package fhir

import (
	"encoding/json"
	"fmt"
	"strings"
)

type BindingStrength int

const (
	BindingStrengthRequired BindingStrength = iota
	BindingStrengthExtensible
	BindingStrengthPreferred
	BindingStrengthExample
)

func (code BindingStrength) Code() string {
	switch code {
	case BindingStrengthRequired:
		return "required"
	case BindingStrengthExtensible:
		return "extensible"
	case BindingStrengthPreferred:
		return "preferred"
	case BindingStrengthExample:
		return "example"
	}
	return "<unknown>"
}

func (code BindingStrength) MarshalJSON() ([]byte, error) {
	return json.Marshal(code.Code())
}

func (code *BindingStrength) UnmarshalJSON(json []byte) error {
	s := strings.Trim(string(json), "\"")
	switch s {
	case "required":
		*code = BindingStrengthRequired
	case "extensible":
		*code = BindingStrengthExtensible
	case "preferred":
		*code = BindingStrengthPreferred
	case "example":
		*code = BindingStrengthExample
	default:
		return fmt.Errorf("unknown BindingStrength code `%s`", s)
	}
	return nil
}

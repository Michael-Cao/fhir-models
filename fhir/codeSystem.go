package fhir

import "encoding/json"

type CodeSystem struct {
	Id      *string             `bson:"id,omitempty" json:"id,omitempty"`
	Concept []CodeSystemConcept `bson:"concept,omitempty" json:"concept,omitempty"`
}

type CodeSystemConcept struct {
	Id         *string             `bson:"id,omitempty" json:"id,omitempty"`
	Code       string              `bson:"code" json:"code"`
	Display    *string             `bson:"display,omitempty" json:"display,omitempty"`
	Definition *string             `bson:"definition,omitempty" json:"definition,omitempty"`
	Concept    []CodeSystemConcept `bson:"concept,omitempty" json:"concept,omitempty"`
}

type OtherCodeSystem CodeSystem

// MarshalJSON marshals the given CodeSystem as JSON into a byte slice
func (r CodeSystem) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		OtherCodeSystem
		ResourceType string `json:"resourceType"`
	}{
		OtherCodeSystem: OtherCodeSystem(r),
		ResourceType:    "CodeSystem",
	})
}

// UnmarshalCodeSystem unmarshals a CodeSystem.
func UnmarshalCodeSystem(b []byte) (CodeSystem, error) {
	var codeSystem CodeSystem
	if err := json.Unmarshal(b, &codeSystem); err != nil {
		return codeSystem, err
	}
	return codeSystem, nil
}

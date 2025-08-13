package fhir

import "encoding/json"

type StructureDefinition struct {
	Id       *string                      `bson:"id,omitempty" json:"id,omitempty"`
	Url      string                       `bson:"url" json:"url"`
	Name     string                       `bson:"name" json:"name"`
	Kind     StructureDefinitionKind      `bson:"kind" json:"kind"`
	Snapshot *StructureDefinitionSnapshot `bson:"snapshot,omitempty" json:"snapshot,omitempty"`
}

type StructureDefinitionSnapshot struct {
	Id      *string             `bson:"id,omitempty" json:"id,omitempty"`
	Element []ElementDefinition `bson:"element" json:"element"`
}

type OtherStructureDefinition StructureDefinition

// MarshalJSON marshals the given StructureDefinition as JSON into a byte slice
func (r StructureDefinition) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		OtherStructureDefinition
		ResourceType string `json:"resourceType"`
	}{
		OtherStructureDefinition: OtherStructureDefinition(r),
		ResourceType:             "StructureDefinition",
	})
}

// UnmarshalStructureDefinition unmarshals a StructureDefinition.
func UnmarshalStructureDefinition(b []byte) (StructureDefinition, error) {
	var structureDefinition StructureDefinition
	if err := json.Unmarshal(b, &structureDefinition); err != nil {
		return structureDefinition, err
	}
	return structureDefinition, nil
}

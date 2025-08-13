package fhir

import "encoding/json"

type ValueSet struct {
	Id      *string          `bson:"id,omitempty" json:"id,omitempty"`
	Name    *string          `bson:"name,omitempty" json:"name,omitempty"`
	Compose *ValueSetCompose `bson:"compose,omitempty" json:"compose,omitempty"`
}

type ValueSetCompose struct {
	Id      *string                  `bson:"id,omitempty" json:"id,omitempty"`
	Include []ValueSetComposeInclude `bson:"include" json:"include"`
}

type ValueSetComposeInclude struct {
	Id *string `bson:"id,omitempty" json:"id,omitempty"`
}

type OtherValueSet ValueSet

// MarshalJSON marshals the given ValueSet as JSON into a byte slice
func (r ValueSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		OtherValueSet
		ResourceType string `json:"resourceType"`
	}{
		OtherValueSet: OtherValueSet(r),
		ResourceType:  "ValueSet",
	})
}

// UnmarshalValueSet unmarshals a ValueSet.
func UnmarshalValueSet(b []byte) (ValueSet, error) {
	var valueSet ValueSet
	if err := json.Unmarshal(b, &valueSet); err != nil {
		return valueSet, err
	}
	return valueSet, nil
}

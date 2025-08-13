package fhir

import "encoding/json"

type Bundle struct {
	Id    *string       `bson:"id,omitempty" json:"id,omitempty"`
	Type  BundleType    `bson:"type" json:"type"`
	Entry []BundleEntry `bson:"entry,omitempty" json:"entry,omitempty"`
}

type BundleEntry struct {
	Id       *string         `bson:"id,omitempty" json:"id,omitempty"`
	Resource json.RawMessage `bson:"resource,omitempty" json:"resource,omitempty"`
}

type OtherBundle Bundle

func (r Bundle) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		OtherBundle
		ResourceType string `json:"resourceType"`
	}{
		OtherBundle:  OtherBundle(r),
		ResourceType: "Bundle",
	})
}

// UnmarshalBundle unmarshals a Bundle.
func UnmarshalBundle(b []byte) (Bundle, error) {
	var bundle Bundle
	if err := json.Unmarshal(b, &bundle); err != nil {
		return bundle, err
	}
	return bundle, nil
}

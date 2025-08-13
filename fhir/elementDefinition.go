package fhir

type ElementDefinition struct {
	Id               *string                   `bson:"id,omitempty" json:"id,omitempty"`
	Path             string                    `bson:"path" json:"path"`
	Code             []Coding                  `bson:"code,omitempty" json:"code,omitempty"`
	Min              *int                      `bson:"min,omitempty" json:"min,omitempty"`
	Max              *string                   `bson:"max,omitempty" json:"max,omitempty"`
	ContentReference *string                   `bson:"contentReference,omitempty" json:"contentReference,omitempty"`
	Type             []ElementDefinitionType   `bson:"type,omitempty" json:"type,omitempty"`
	Binding          *ElementDefinitionBinding `bson:"binding,omitempty" json:"binding,omitempty"`
}

type ElementDefinitionType struct {
	Id   *string `bson:"id,omitempty" json:"id,omitempty"`
	Code *string `bson:"code,omitempty" json:"code,omitempty"`
}

type ElementDefinitionBinding struct {
	Id       *string         `bson:"id,omitempty" json:"id,omitempty"`
	Strength BindingStrength `bson:"strength" json:"strength"`
	ValueSet *string         `bson:"valueSet,omitempty" json:"valueSet,omitempty"`
}

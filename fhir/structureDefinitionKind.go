package fhir

import (
	"encoding/json"
	"fmt"
	"strings"
)

type StructureDefinitionKind int

const (
	StructureDefinitionKindPrimitiveType StructureDefinitionKind = iota
	StructureDefinitionKindComplexType
	StructureDefinitionKindResource
	StructureDefinitionKindLogical
)

func (code StructureDefinitionKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(code.Code())
}
func (code *StructureDefinitionKind) UnmarshalJSON(json []byte) error {
	s := strings.Trim(string(json), "\"")
	switch s {
	case "primitive-type":
		*code = StructureDefinitionKindPrimitiveType
	case "complex-type":
		*code = StructureDefinitionKindComplexType
	case "resource":
		*code = StructureDefinitionKindResource
	case "logical":
		*code = StructureDefinitionKindLogical
	default:
		return fmt.Errorf("unknown StructureDefinitionKind code `%s`", s)
	}
	return nil
}
func (code StructureDefinitionKind) String() string {
	return code.Code()
}
func (code StructureDefinitionKind) Code() string {
	switch code {
	case StructureDefinitionKindPrimitiveType:
		return "primitive-type"
	case StructureDefinitionKindComplexType:
		return "complex-type"
	case StructureDefinitionKindResource:
		return "resource"
	case StructureDefinitionKindLogical:
		return "logical"
	}
	return "<unknown>"
}
func (code StructureDefinitionKind) Display() string {
	switch code {
	case StructureDefinitionKindPrimitiveType:
		return "Primitive Data Type"
	case StructureDefinitionKindComplexType:
		return "Complex Data Type"
	case StructureDefinitionKindResource:
		return "Resource"
	case StructureDefinitionKindLogical:
		return "Logical"
	}
	return "<unknown>"
}
func (code StructureDefinitionKind) Definition() string {
	switch code {
	case StructureDefinitionKindPrimitiveType:
		return "A primitive type that has a value and an extension. These can be used throughout complex datatype, Resource and extension definitions. Only the base specification can define primitive types."
	case StructureDefinitionKindComplexType:
		return "A  complex structure that defines a set of data elements that is suitable for use in 'resources'. The base specification defines a number of complex types, and other specifications can define additional types. These structures do not have a maintained identity."
	case StructureDefinitionKindResource:
		return "A 'resource' - a directed acyclic graph of elements that aggregrates other types into an identifiable entity. The base FHIR resources are defined by the FHIR specification itself but other 'resources' can be defined in additional specifications (though these will not be recognised as 'resources' by the FHIR specification (i.e. they do not get end-points etc, or act as the targets of references in FHIR defined resources - though other specificatiosn can treat them this way)."
	case StructureDefinitionKindLogical:
		return "A pattern or a template that is not intended to be a real resource or complex type."
	}
	return "<unknown>"
}

package fhir

type BindingStrength int

const (
	BindingStrengthRequired BindingStrength = iota
	BindingStrengthExtensible
	BindingStrengthPreferred
	BindingStrengthExample
)

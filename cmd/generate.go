/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/dave/jennifer/jen"
	"github.com/spf13/cobra"

	"github.com/Michael-Cao/fhir-models/cmd/utils"
	"github.com/Michael-Cao/fhir-models/fhir"
)

type Resource struct {
	ResourceType string
	Url          *string
	Version      *string
	Name         *string
}

func UnmarshalResource(b []byte) (Resource, error) {
	var resource Resource
	if err := json.Unmarshal(b, &resource); err != nil {
		return resource, err
	}
	return resource, nil
}

var namePattern = regexp.MustCompile("^[A-Z]([A-Za-z0-9_]){0,254}$")

type ResourceMap = map[string]map[string][]byte

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate fhir models",
	Run: func(cmd *cobra.Command, args []string) {
		version, _ := cmd.Flags().GetString("version")
		url := fmt.Sprintf("https://www.hl7.org/fhir/%s/definitions.json.zip", version)

		var dir string
		tmpDir, _ := cmd.Flags().GetString("inputdir")
		if tmpDir == "" {
			pattern := fmt.Sprintf("fhir-%s-*", version)
			tmpDir, err := os.MkdirTemp("/tmp", pattern)
			if err != nil {
				fmt.Printf("failed to create directory %v: %v\n", tmpDir, err)
				return
			}

			filename, err := utils.Download(url, tmpDir)
			if err != nil {
				fmt.Printf("failed to download file: %v", err)
				return
			}
			fmt.Printf("\n%s\n", *filename)

			utils.Unzip(*filename)
			dir = tmpDir
		} else {
			dir = tmpDir
		}

		fmt.Println(dir)
		processFiles(dir)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.PersistentFlags().String("version", "R4", "Specify the fhir version")
	generateCmd.PersistentFlags().String("inputdir", "", "Specify the input directory")
}

func processFiles(tmpDir string) error {

	resources := make(ResourceMap)
	resources["StructureDefinition"] = make(map[string][]byte)
	resources["ValueSet"] = make(map[string][]byte)
	resources["CodeSystem"] = make(map[string][]byte)

	err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".json") {
			bytes, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			fmt.Printf("Generate Go sources from file: %s\n", path)
			resource, err := UnmarshalResource(bytes)
			if err != nil {
				return err
			}

			fmt.Println(resource.ResourceType)
			if resource.ResourceType == "Bundle" {
				bundle, err := fhir.UnmarshalBundle(bytes)
				if err != nil {
					fmt.Printf("%v\n", err)
					return err
				}
				for _, entry := range bundle.Entry {
					entryResource, err := UnmarshalResource(entry.Resource)
					if err != nil {
						fmt.Printf("%v\n", err)
						return err
					}
					switch entryResource.ResourceType {
					case "StructureDefinition":
						if entryResource.Name != nil {
							resources[entryResource.ResourceType][*entryResource.Name] = entry.Resource
						}
					case "CodeSystem":
						if entryResource.Url != nil {
							if entryResource.Version != nil {
								resources[entryResource.ResourceType][*entryResource.Url+"|"+*entryResource.Version] = entry.Resource
								resources[entryResource.ResourceType][*entryResource.Url] = entry.Resource
							} else {
								resources[entryResource.ResourceType][*entryResource.Url] = entry.Resource
							}
						}
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	requiredTypes := make(map[string]bool, 0)
	requiredValueSetBindings := make(map[string]bool, 0)

	for _, bytes := range resources["StructureDefinition"] {
		structureDefinition, err := fhir.UnmarshalStructureDefinition(bytes)
		if err != nil {
			return nil
		}

		if structureDefinition.Kind == fhir.StructureDefinitionKindResource &&
			structureDefinition.Name != "Element" &&
			structureDefinition.Name != "BackboneElement" {
			goFile, err := generateResourceOrType(resources, requiredTypes, requiredValueSetBindings, structureDefinition)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = goFile.Save(FirstLower(structureDefinition.Name) + ".go")
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func generateResourceOrType(resources ResourceMap, requiredTypes map[string]bool, requiredValueSetBindings map[string]bool, definition fhir.StructureDefinition) (*jen.File, error) {

	elementDefinitions := definition.Snapshot.Element
	if len(elementDefinitions) == 0 {
		return nil, fmt.Errorf("missing element definitions in structure definition `%s`", definition.Name)
	}

	fmt.Printf("Generate Go sources for StructureDefinition: %s\n", definition.Name)
	file := jen.NewFile("fhir")

	file.Commentf("%s is documented here %s", definition.Name, definition.Url)

	if definition.Kind == fhir.StructureDefinitionKindResource {
		file.Commentf("Unmarshal%s unmarshals a %s.", definition.Name, definition.Name)
	}

	var err error
	file.Type().Id(definition.Name).StructFunc(func(rootStruct *jen.Group) {
		_, err = appendFields(resources, requiredTypes, requiredValueSetBindings, file, rootStruct, definition.Name, elementDefinitions, 1, 1)
	})

	if err != nil {
		return nil, err
	}

	// generate marshal
	if definition.Kind == fhir.StructureDefinitionKindResource {
		file.Type().Id("Other" + definition.Name).Id(definition.Name)
		file.Commentf("MarshalJSON marshals the given %s as JSON into a byte slice", definition.Name)
		file.Func().Params(jen.Id("r").Id(definition.Name)).Id("MarshalJSON").Params().
			Params(jen.Op("[]").Byte(), jen.Error()).Block(
			jen.Return().Qual("encoding/json", "Marshal").Call(jen.Struct(
				jen.Id("Other"+definition.Name),
				jen.Id("ResourceType").String().Tag(map[string]string{"json": "resourceType"}),
			).Values(jen.Dict{
				jen.Id("Other" + definition.Name): jen.Id("Other" + definition.Name).Call(jen.Id("r")),
				jen.Id("ResourceType"):            jen.Lit(definition.Name),
			})),
		)
	}

	// generate unmarshal
	if definition.Kind == fhir.StructureDefinitionKindResource {
		file.Commentf("Unmarshal%s unmarshals a %s.", definition.Name, definition.Name)
		file.Func().Id("Unmarshal"+definition.Name).
			Params(jen.Id("b").Op("[]").Byte()).
			Params(jen.Id(definition.Name), jen.Error()).
			Block(
				jen.Var().Id(FirstLower(definition.Name)).Id(definition.Name),
				jen.If(
					jen.Err().Op(":=").Qual("encoding/json", "Unmarshal").Call(
						jen.Id("b"),
						jen.Op("&").Id(FirstLower(definition.Name)),
					),
					jen.Err().Op("!=").Nil(),
				).Block(
					jen.Return(jen.Id(FirstLower(definition.Name)), jen.Err()),
				),
				jen.Return(jen.Id(FirstLower(definition.Name)), jen.Nil()),
			)
	}

	return file, nil
}

func FirstLower(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

func appendFields(resources ResourceMap, requiredTypes map[string]bool, requiredValueSetBindings map[string]bool,
	file *jen.File, fields *jen.Group, parentName string, elementDefinitions []fhir.ElementDefinition, start,
	level int) (int, error) {
	//fmt.Printf("appendFields parentName=%s, start=%d, level=%d\n", parentName, start, level)
	for i := start; i < len(elementDefinitions); i++ {
		element := elementDefinitions[i]
		pathParts := strings.Split(element.Path, ".")
		if len(pathParts) == level+1 {
			// direct childs
			name := strings.Title(pathParts[level])

			// support contained resources later
			if name != "Contained" {
				switch len(element.Type) {
				case 0:
					if element.ContentReference != nil && (*element.ContentReference)[:1] == "#" {
						statement := fields.Id(name)

						if *element.Max == "*" {
							statement.Op("[]")
						} else if *element.Min == 0 {
							statement.Op("*")
						}

						typeIdentifier := ""
						for _, pathPart := range strings.Split((*element.ContentReference)[1:], ".") {
							typeIdentifier = typeIdentifier + strings.Title(pathPart)
						}
						statement.Id(typeIdentifier).Tag(map[string]string{"json": pathParts[level] + ",omitempty", "bson": pathParts[level] + ",omitempty"})
					}
				case 1:
					var err error
					i, err = addFieldStatement(resources, requiredTypes, requiredValueSetBindings, file, fields,
						pathParts[level], parentName, elementDefinitions, i, level, element.Type[0])

					if err != nil {
						return 0, err
					}
				default: //polymorphic type
					name = strings.Replace(pathParts[level], "[x]", "", -1)
					for _, eleType := range element.Type {
						name := name + strings.Title(*eleType.Code)

						var err error
						i, err = addFieldStatement(resources, requiredTypes, requiredValueSetBindings, file, fields,
							name, parentName, elementDefinitions, i, level, eleType)

						if err != nil {
							return 0, err
						}
					}
				}
			}
		} else {
			// index of the next parent sibling
			return i, nil
		}
	}
	return 0, nil
}

func addFieldStatement(
	resources ResourceMap,
	requiredTypes map[string]bool,
	requiredValueSetBindings map[string]bool,
	file *jen.File,
	fields *jen.Group,
	name string,
	parentName string,
	elementDefinitions []fhir.ElementDefinition,
	elementIndex, level int,
	elementType fhir.ElementDefinitionType,
) (idx int, err error) {
	fieldName := strings.Title(name)
	element := elementDefinitions[elementIndex]
	statement := fields.Id(fieldName)

	switch *elementType.Code {
	case "code":
		if *element.Max == "*" {
			statement.Op("[]")
		} else if *element.Min == 0 {
			statement.Op("*")
		}

		if url := requiredValueSetBinding(element); url != nil {
			if bytes := resources["ValueSet"][*url]; bytes != nil {
				valueSet, err := fhir.UnmarshalValueSet(bytes)
				if err != nil {
					return 0, err
				}
				if name := valueSet.Name; name != nil {
					if !namePattern.MatchString(*name) {
						fmt.Printf("Skip generating an enum for a ValueSet binding to `%s` because the ValueSet has a non-conforming name.\n", *name)
						statement.Id("string")
					} else if len(valueSet.Compose.Include) > 1 {
						fmt.Printf("Skip generating an enum for a ValueSet binding to `%s` because the ValueSet includes more than one CodeSystem.\n", *valueSet.Name)
						statement.Id("string")
					} else if codeSystemUrl := canonical(valueSet.Compose.Include[0]); resources["CodeSystem"][codeSystemUrl] == nil {
						fmt.Printf("Skip generating an enum for a ValueSet binding to `%s` because the ValueSet includes the non-existing CodeSystem with canonical URL `%s`.\n", *valueSet.Name, codeSystemUrl)
						statement.Id("string")
					} else {
						requiredValueSetBindings[*url] = true
						statement.Id(*name)
					}
				} else {
					return 0, fmt.Errorf("missing name in ValueSet with canonical URL `%s`", *url)
				}
			} else {
				statement.Id("string")
			}
		} else {
			statement.Id("string")
		}
	case "Resource":
		statement.Qual("encoding/json", "RawMessage")
	default:
		if *element.Max == "*" {
			statement.Op("[]")
		} else if *element.Min == 0 {
			statement.Op("*")
		}

		var typeIdentifier string
		if parentName == "Element" && fieldName == "Id" ||
			parentName == "Extension" && fieldName == "Url" {
			typeIdentifier = "string"
		} else {
			typeIdentifier = typeCodeToTypeIdentifier(elementType.Code)
		}
		if typeIdentifier == "Element" || typeIdentifier == "BackboneElement" {
			backboneElementName := parentName + fieldName
			statement.Id(backboneElementName)
			var err error
			file.Type().Id(backboneElementName).StructFunc(func(childFields *jen.Group) {
				//var err error
				elementIndex, err = appendFields(resources, requiredTypes, requiredValueSetBindings, file, childFields,
					backboneElementName, elementDefinitions, elementIndex+1, level+1)
			})
			if err != nil {
				return 0, err
			}
			elementIndex--
		} else if typeIdentifier == "decimal" {
			statement.Qual("encoding/json", "Number")
		} else {
			if unicode.IsUpper(rune(typeIdentifier[0])) {
				requiredTypes[typeIdentifier] = true
			}
			statement.Id(typeIdentifier)
		}
	}

	if *element.Min == 0 {
		statement.Tag(map[string]string{"json": name + ",omitempty", "bson": name + ",omitempty"})
	} else {
		statement.Tag(map[string]string{"json": name, "bson": name})
	}

	return elementIndex, err
}

func requiredValueSetBinding(elementDefinition fhir.ElementDefinition) *string {
	if elementDefinition.Binding != nil {
		binding := *elementDefinition.Binding
		if binding.Strength == fhir.BindingStrengthRequired {
			return binding.ValueSet
		}
	}
	return nil
}

func typeCodeToTypeIdentifier(typeCode *string) string {
	switch *typeCode {
	case "base64Binary":
		return "string"
	case "boolean":
		return "bool"
	case "canonical":
		return "string"
	case "code":
		return "string"
	case "date":
		return "string"
	case "dateTime":
		return "string"
	case "id":
		return "string"
	case "instant":
		return "string"
	case "integer":
		return "int"
	case "markdown":
		return "string"
	case "oid":
		return "string"
	case "positiveInt":
		return "int"
	case "string":
		return "string"
	case "time":
		return "string"
	case "unsignedInt":
		return "int"
	case "uri":
		return "string"
	case "url":
		return "string"
	case "uuid":
		return "string"
	case "xhtml":
		return "string"
	case "http://hl7.org/fhirpath/System.String":
		return "string"
	default:
		return *typeCode
	}
}

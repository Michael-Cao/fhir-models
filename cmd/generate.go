/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	return file, nil
}

func FirstLower(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

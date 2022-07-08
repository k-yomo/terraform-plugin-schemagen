package terraform_plugin_schemagen

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"sort"
	"strings"
	"text/template"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:embed schemafile.gotmpl
var schemaFileTemplate []byte

//go:embed schema.gotmpl
var schemaTemplate []byte

type ResourceAttribute struct {
}

func Generate(schema *tfjson.ProviderSchema) ([]byte, error) {
	schemaFileTpl := template.Must(template.New("schemaFile").Parse(string(schemaFileTemplate)))
	schemaTpl := template.Must(template.New("schema").
		Funcs(template.FuncMap{
			"ToUpperCamel":           ToUpperCamel,
			"CityTypeToGoTypeString": CityTypeToGoTypeString,
		}).
		Parse(string(schemaTemplate)))

	body := new(bytes.Buffer)
	for name, resource := range schema.ResourceSchemas {
		type Attribute struct {
			*tfjson.SchemaAttribute
			Name          string
			IsNestedField bool
		}
		fmt.Println(name)
		attributes := make([]Attribute, 0, len(resource.Block.Attributes))
		for name, block := range resource.Block.NestedBlocks {
			fmt.Println("        ", name)
			fmt.Println("        ", block.Block)
		}
		for name, attr := range resource.Block.Attributes {
			attributes = append(attributes, Attribute{
				SchemaAttribute: attr,
				Name:            name,
				IsNestedField:   attr.AttributeNestedType != nil,
			})
		}
		// for consistent output
		sort.Slice(attributes, func(i, j int) bool {
			return attributes[i].Name < attributes[i].Name
		})
		params := struct {
			ResourceName string
			ReceiverName string
			Attributes   []Attribute
		}{
			ResourceName: ToUpperCamel(name),
			ReceiverName: string(name[0]),
			Attributes:   attributes,
		}
		if err := schemaTpl.Execute(body, params); err != nil {
			return nil, fmt.Errorf("execute factory template: %w", err)
		}
	}

	out := new(bytes.Buffer)
	params := struct {
		Body string
	}{
		Body: body.String(),
	}
	if err := schemaFileTpl.Execute(out, params); err != nil {
		return nil, fmt.Errorf("execute schemaFile template: %w", err)
	}
	formattedOut, err := format.Source(out.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format.Source: %w", err)
	}

	return formattedOut, nil
}

func ToUpperCamel(s string) string {
	strs := strings.Split(s, "_")
	for i, str := range strs {
		strs[i] = cases.Title(language.English).String(str)
	}
	return strings.Join(strs, "")
}

func CityTypeToGoTypeString(t cty.Type) string {
	switch t {
	case cty.String:
		return "string"
	case cty.Number:
		return "int"
	case cty.Bool:
		return "bool"
	}
	if t.IsListType() {
		if t.ListElementType() != nil {
			return fmt.Sprintf("[]%s", CityTypeToGoTypeString(*t.ListElementType()))
		}
		return "[]interface{}"
	}
	if t.IsSetType() {
		if t.SetElementType() != nil {
			return fmt.Sprintf("[]%s", CityTypeToGoTypeString(*t.SetElementType()))
		}
		return "[]interface{}"
	}
	if t.IsMapType() {
		if t.MapElementType() != nil {
			return fmt.Sprintf("map[string]%s", CityTypeToGoTypeString(*t.MapElementType()))
		}
		return "map[string]interface{}"
	}
	if t.IsObjectType() {
		return "map[string]interface{}"
	}
	panic("unexpected type")
}

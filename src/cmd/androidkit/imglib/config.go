package imglib

import (
	"fmt"

	"github.com/containerd/containerd/reference"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v2"
)

// AndroidX is the type of an AndroidX config file
type AndroidX struct {
	Kernel KernelConfig `kernel:"cmdline,omitempty" json:"kernel,omitempty"`
	Init   []string     `init:"cmdline" json:"init"`
	Trust  TrustConfig  `yaml:"trust,omitempty" json:"trust,omitempty"`

	InitRefs []*reference.Spec
}

// KernelConfig is the type of the config for a kernel
type KernelConfig struct {
	Image   string `yaml:"image" json:"image"`
	Cmdline string `yaml:"cmdline,omitempty" json:"cmdline,omitempty"`
	//Tar     *string `yaml:"tar,omitempty" json:"tar,omitempty"`

	Ref *reference.Spec
}

// TrustConfig is the type of a content trust config
type TrustConfig struct {
	Image []string `yaml:"image,omitempty" json:"image,omitempty"`
	Org   []string `yaml:"org,omitempty" json:"org,omitempty"`
}

// Image is the type of an image config
type Image struct {
	Name  string `yaml:"name" json:"name"`
	Image string `yaml:"image" json:"image"`
}

// github.com/go-yaml/yaml treats map keys as interface{} while encoding/json
// requires them to be strings, integers or to implement encoding.TextMarshaler.
// Fix this up by recursively mapping all map[interface{}]interface{} types into
// map[string]interface{}.
// see http://stackoverflow.com/questions/40737122/convert-yaml-to-json-without-struct-golang#answer-40737676
func convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convert(v)
		}
	}
	return i
}

func extractReferences(x *AndroidX) error {
	if x.Kernel.Image != "" {
		r, err := reference.Parse(x.Kernel.Image)
		if err != nil {
			return fmt.Errorf("extract kernel image reference: %v", err)
		}
		x.Kernel.Ref = &r
	}
	for _, ii := range x.Init {
		r, err := reference.Parse(ii)
		if err != nil {
			return fmt.Errorf("extract on boot image reference: %v", err)
		}
		x.InitRefs = append(x.InitRefs, &r)
	}
	return nil
}

// NewConfig parses a config file
func NewConfig(config []byte) (AndroidX, error) {
	x := AndroidX{}

	// Parse raw yaml
	var rawYaml interface{}
	err := yaml.Unmarshal(config, &rawYaml)
	if err != nil {
		return x, err
	}

	// Convert to raw JSON
	rawJSON := convert(rawYaml)

	// Validate raw yaml with JSON schema
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewGoLoader(rawJSON)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return x, err
	}
	if !result.Valid() {
		fmt.Printf("The configuration file is invalid:\n")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
		return x, fmt.Errorf("invalid configuration file")
	}

	// Parse yaml
	err = yaml.Unmarshal(config, &x)
	if err != nil {
		return x, err
	}

	if err := extractReferences(&x); err != nil {
		return x, err
	}
	return x, nil
}

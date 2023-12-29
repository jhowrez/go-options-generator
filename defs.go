package main

type OptionDefinition struct {
	Name         string      `yaml:"name"`
	Description  string      `yaml:"description"`
	Yaml         string      `yaml:"yaml"`
	Format       string      `yaml:"format"`
	DefaultValue interface{} `yaml:"default"`
}

type OptionsDefinitionFile struct {
	PackageName  string
	ExtraImports []string
	Options      []OptionDefinition `yaml:"options"`
	StructMap    map[string]interface{}
}

type GeneratorConfiguration struct {
	InputYaml      string
	OutputMarkdown string
	OutputGo       string
	Package        string
}

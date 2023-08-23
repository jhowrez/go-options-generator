package main

import (
	"bytes"
	"flag"
	"go/format"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"
)

func templateLoadExecuteSave(name string, data any, outputfilename string, opts templateExecuteOptions) {
	outFile, err := os.Create(outputfilename)
	fatalOnErr(err)
	defer outFile.Close()

	tpl := template.Must(
		template.New(name).Funcs(sprig.FuncMap()).Funcs(localFuncMap()).ParseFiles("templates/" + name),
	)

	tmplOutBuf := bytes.NewBuffer(nil)
	err = tpl.Execute(tmplOutBuf, data)
	fatalOnErr(err)

	if opts.isGolang {
		formattedOutput, err := format.Source(tmplOutBuf.Bytes())
		fatalOnErr(err)
		tmplOutBuf.Reset()
		tmplOutBuf.Write(formattedOutput)
	}

	_, err = outFile.Write(tmplOutBuf.Bytes())
	fatalOnErr(err)
}

func main() {
	genCfg := GeneratorConfiguration{}
	// parse input flags
	flag.StringVar(&genCfg.InputYaml, "in", "options.yaml", "input options yaml definition file")
	flag.StringVar(&genCfg.OutputMarkdown, "out_md", "OPTIONS.md", "output markdown file")
	flag.StringVar(&genCfg.OutputGo, "out_go", "options_generated.go", "output options golang file")
	flag.StringVar(&genCfg.Package, "pkg", "options", "output go package string")
	flag.Parse()

	// load definition file
	inOptsFileDataRaw, err := os.ReadFile(genCfg.InputYaml)
	fatalOnErr(err)

	optsMap := OptionsDefinitionFile{PackageName: genCfg.Package}

	// unmarshal definition file
	err = yaml.Unmarshal(inOptsFileDataRaw, &optsMap)
	fatalOnErr(err)
	optsMap.consolidateStructMap()

	// OPTIONS.md
	templateLoadExecuteSave("OPTIONS.md.gotmpl", optsMap, genCfg.OutputMarkdown, templateExecuteOptions{})
	// options_types.go
	templateLoadExecuteSave("options.go.gotmpl", optsMap, genCfg.OutputGo, templateExecuteOptions{
		isGolang: true,
	})
}

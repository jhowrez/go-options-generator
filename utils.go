package main

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"text/template"
)

func GetValueByPathFromMap(data map[string]any, key string, passedKey string) (result any, found bool) {
	keyAndPath := strings.SplitN(key, ".", 2)
	currentKey := keyAndPath[0]
	if passedKey != "" {
		passedKey = passedKey + "." + currentKey
	} else {
		passedKey = currentKey
	}

	if _, isKeyExistInData := data[currentKey]; !isKeyExistInData {
		// panic(fmt.Sprintf("[W] key path { %s } not found", passedKey))
		return nil, false
	} else {

		if len(keyAndPath) > 1 {
			remainingPath := keyAndPath[1]
			switch data[currentKey].(type) {
			case map[string]any:
				if result, found = GetValueByPathFromMap(data[currentKey].(map[string]any), remainingPath, passedKey); found {
					return
				}
			}
		} else {
			return data[currentKey], true
		}
	}

	return nil, false
}

func getKeyLevelOffset(fullkey string, offset int) (string, bool, string) {
	l := strings.Split(fullkey, ".")
	if offset+1 >= len(l) {
		return fullkey, true, l[len(l)-1]
	}
	l = l[:offset+1]
	return strings.Join(l, "."), false, l[offset]
}

func getAndSetFromKey(inMap map[string]interface{}, fullKey string, curLevel int, value any) {
	_, lastOffset, refKey := getKeyLevelOffset(fullKey, curLevel)
	_, exists := inMap[refKey]
	if !exists {
		if lastOffset {
			inMap[refKey] = value
			return
		} else {
			inMap[refKey] = make(map[string]interface{})
		}
	}
	getAndSetFromKey(inMap[refKey].(map[string]interface{}), fullKey, curLevel+1, value)
}

func (odf *OptionsDefinitionFile) consolidate() {
	// struct map
	outMap := map[string]interface{}{}
	for _, opt := range odf.Options {
		getAndSetFromKey(outMap, opt.Yaml, 0, []any{opt.DefaultValue, opt.Yaml, opt.Description, opt.Format})
	}
	odf.StructMap = outMap

	// extra imports
	for _, opt := range odf.Options {
		if opt.Format != "" {
			importStr := getImportForFormat(opt.Format)
			if importStr != "" {
				odf.ExtraImports = append(odf.ExtraImports, importStr)
			}
		}
	}
	odf.ExtraImports = removeDuplicate(odf.ExtraImports)
}

type templateExecuteOptions struct {
	isGolang bool
}

func fatalOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getTypeForFormat(f string) string {
	if f == "" {
		return ""
	}
	switch f {
	case "duration":
		return "time.Duration"
	default:
		log.Fatalf("format %s is not defined", f)
	}

	return "any"
}

func getImportForFormat(f string) string {
	if f == "" {
		return ""
	}
	switch f {
	case "duration":
		return "time"
	default:
		log.Fatalf("format %s is not defined", f)
	}

	return ""
}

func localFuncMap() template.FuncMap {
	return template.FuncMap{
		"typeOf": func(a any, f string) string {
			if f != "" {
				return getTypeForFormat(f)
			}
			if a == nil {
				return "any"
			}
			return fmt.Sprintf("%T", a)
		},
		"doubleQuotes": func(a any) string {
			switch a.(type) {
			case string:
				return fmt.Sprintf("\"%s\"", a)
			case nil:
				return "nil"
			}
			return fmt.Sprintf("%v", a)
		},
		"splitYamlKey": func(key string) []string {
			return strings.Split(key, ".")
		},
		"isMap": func(in any) bool {
			k := reflect.TypeOf(in).Kind()
			return k == reflect.Map
		},
		"isDuration": func(f string) bool {
			return f == "duration"
		},
	}
}

func removeDuplicate[T string | int](sliceList []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

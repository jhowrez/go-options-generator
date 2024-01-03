package main

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"text/template"

	"github.com/huandu/xstrings"
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
			if len(importStr) > 0 {
				odf.ExtraImports = append(odf.ExtraImports, importStr...)
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

func getImportForFormat(f string) []string {
	if f == "" {
		return []string{}
	}
	switch f {
	case "duration":
		return []string{"github.com/jhowrez/go-options-generator/pkg/wrappers", "time"}
	default:
		log.Fatalf("format %s is not defined", f)
	}

	return []string{}
}

func typeForValue(v any) string {
	vKind := reflect.ValueOf(v).Kind()
	tStr := ""
	switch vKind {
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		tStr += "[]"
		isSingleType := true
		valueList := v.([]interface{})
		lastValue := reflect.Interface
		for i, e := range valueList {
			k := reflect.ValueOf(e).Kind()
			if i == 0 {
				lastValue = k
				continue
			}
			if lastValue != k {
				isSingleType = false
				break
			}
		}

		if isSingleType {
			tStr += lastValue.String()
		} else {
			tStr += "interface{}"
		}

	default:
		tStr = fmt.Sprintf("%T", v)
	}

	return tStr

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
			return typeForValue(a)
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
		"varAccessName": func(keys []string) string {
			accessName := ""
			for i, key := range keys {
				if i > 0 {
					accessName += "."
				}
				accessName += xstrings.ToCamelCase(key)
			}
			return accessName
		},
		"defaultValueWrapper": func(varName string, v any, format string) any {
			vKind := reflect.ValueOf(v).Kind()
			t := fmt.Sprintf("%s = ", varName)
			vStr := ""

			switch vKind {
			case reflect.Array:
				fallthrough
			case reflect.Slice:
				valueList := v.([]interface{})
				t += typeForValue(valueList)
				t += "{"
				for i, e := range valueList {
					if i > 0 {
						t += ","
					}
					if reflect.ValueOf(e).Kind() == reflect.String {
						t += fmt.Sprintf("\"%s\"", e)
					} else {
						t += fmt.Sprintf("%v", e)
					}

				}
				t += "}"
			case reflect.String:
				vStr += fmt.Sprintf("\"%s\"", v)
			default:
				vStr += fmt.Sprintf("%v", v)
			}
			switch format {
			case "duration":
				t += fmt.Sprintf("wrappers.MustParseDuration(%s)", vStr)
			case "":
				t += vStr
			default:
				log.Panicf("invalid format '%s'", format)
			}
			return t
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

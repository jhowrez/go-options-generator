package main

import (
	"fmt"
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

func (odf *OptionsDefinitionFile) consolidateStructMap() {
	outMap := map[string]interface{}{}
	for _, opt := range odf.Options {
		getAndSetFromKey(outMap, opt.Yaml, 0, []any{opt.DefaultValue, opt.Yaml, opt.Description})
	}
	odf.StructMap = outMap
}

type templateExecuteOptions struct {
	isGolang bool
}

func fatalOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func localFuncMap() template.FuncMap {
	return template.FuncMap{
		"typeOf": func(a any) string {
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
	}
}

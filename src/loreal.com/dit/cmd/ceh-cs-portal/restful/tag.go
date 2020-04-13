package restful

import (
	"log"
	"reflect"
	"strconv"
	"strings"
)

//FieldTag - table field tag for restful models
type FieldTag struct {
	Name      string
	FieldName string
	DataType  string
	Default   string
	Index     string
	GoType    reflect.Type
}

//ParseTags - Parse field tags from source struct
func ParseTags(source interface{}) []FieldTag {
	t := reflect.TypeOf(source)
	nFields := t.NumField()
	r := make([]FieldTag, 0, nFields)
	for i := 0; i < nFields; i++ {
		field := t.Field(i)
		r = append(r, FieldTag{
			Name:      field.Name,
			FieldName: getTagStr(&field, "name", strings.ToLower(field.Name)),
			DataType:  getTagStr(&field, "type", "text"),
			Default:   getTagStr(&field, "default", ""),
			Index:     getTagStr(&field, "index", ""),
			GoType:    field.Type,
		})
	}
	return r
}

func getTagInt(field *reflect.StructField, tagName string, defaultValue int) int {
	v := field.Tag.Get(tagName)
	if v == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalln(err)
		return defaultValue
	}
	return intValue
}

func getTagStr(field *reflect.StructField, tagName string, defaultValue string) string {
	v := field.Tag.Get(tagName)
	if v == "" {
		return defaultValue
	}
	return v
}

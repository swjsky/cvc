package restful

import (
	"database/sql"
	"log"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func (a *SQLiteAdapter) parseParams(params url.Values) (keys []string, values []interface{}) {
	var (
		err error
		val interface{}
	)
	keys = make([]string, 0, len(a.tags))
	values = make([]interface{}, 0, len(a.tags))
	for _, tag := range a.tags {
		fname := strings.ToLower(tag.FieldName)
		s := params.Get(fname)
		if s == "" {
			continue
		}
		switch strings.ToLower(tag.DataType) {
		case "int", "integer":
			val, err = strconv.Atoi(s)
			if err != nil {
				log.Printf("[ERR] - [parseParams][int]: s='%s', err=%v\n", s, err)
			}
		case "real", "float", "double":
			val, err = strconv.ParseFloat(s, 64)
			if err != nil {
				log.Printf("[ERR] - [parseParams][float]: s='%s', err=%v\n", s, err)
			}
		case "bool", "boolean":
			val, err = strconv.ParseBool(s)
			if err != nil {
				log.Printf("[ERR] - [parseParams][bool]: s='%s', err=%v\n", s, err)
			}
		case "date", "datetime":
			switch s {
			case "now":
				val = time.Now().Local()
			default:
				val, err = time.ParseInLocation(time.RFC3339, s, time.Local)
				if err != nil {
					val, err = time.ParseInLocation("2006-01-02", s, time.Local)
					if err != nil {
						log.Printf("[ERR] - [parseParams][datetime]: s='%s', err=%v\n", s, err)
					}
				}
			}
		default:
			val = s
		}
		keys = append(keys, fname)
		values = append(values, val)
	}
	return
}

func (a *SQLiteAdapter) getFields(includeID bool) (fields []string) {
	fields = make([]string, 0, len(a.tags))
	for _, tag := range a.tags {
		fname := strings.ToLower(tag.FieldName)
		if fname == "id" && !includeID {
			continue
		}
		fields = append(fields, fname)
	}
	return
}

func (a *SQLiteAdapter) newRecord(buffer *[]interface{}) (record map[string]interface{}) {
	record = make(map[string]interface{}, len(a.tags))
	*buffer = (*buffer)[:0]
	for _, tag := range a.tags {
		// var ptr interface{}
		// switch tag.GoType.Kind() {
		// case reflect.Bool:
		// 	ptr = &sql.NullBool{}
		// case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 	ptr = &sql.NullInt64{}
		// case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// 	ptr = &sql.NullInt64{}
		// case reflect.Float32, reflect.Float64:
		// 	ptr = &sql.NullFloat64{}
		// case reflect.String:
		// 	ptr = &sql.NullString{}
		// case reflect.Struct:
		// 	switch strings.ToLower(tag.DataType) {
		// 	case "date", "datetime":
		// 		ptr = &sql.NullString{}
		// 	default:
		// 		ptr = &sql.NullString{}
		// 	}
		// // case reflect.Uintptr:
		// // case reflect.Complex64:
		// // case reflect.Complex128:
		// // case reflect.Array:
		// // case reflect.Chan:
		// // case reflect.Func:
		// // case reflect.Interface:
		// // case reflect.Map:
		// // case reflect.Ptr:
		// // case reflect.Slice:
		// // case reflect.UnsafePointer:
		// default:
		// 	ptr = reflect.New(tag.GoType).Interface()
		// }
		ptr := reflect.New(tag.GoType).Interface()
		record[strings.ToLower(tag.Name)] = ptr
		*buffer = append(*buffer, ptr)
	}
	return
}

func (a *SQLiteAdapter) scanRows(rows *sql.Rows, err error) (records []*map[string]interface{}) {
	records = make([]*map[string]interface{}, 0, 0)
	if err != nil {
		log.Printf("[ERR] - [scanRows] err: %v\n", err)
		return
	}
	if rows == nil {
		return
	}
	defer rows.Close()
	buffer := make([]interface{}, 0, len(a.tags))
	for rows.Next() {
		record := a.newRecord(&buffer)
		if err := rows.Scan(buffer...); err != nil {
			log.Printf("[ERR] - [scanRows] err: %v\n", err)
		}
		records = append(records, &record)
	}
	return
}

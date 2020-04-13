package restful

import (
	"reflect"
	"strings"
	"time"
)

func defaultForSQLNull(tag FieldTag) string {
	switch tag.GoType.Kind() {
	case reflect.Bool:
		return "false"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "-1"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "0"
	case reflect.Float32, reflect.Float64:
		return "-1"
	case reflect.String:
		return "''"
	case reflect.Struct:
		switch strings.ToLower(tag.DataType) {
		case "date", "datetime":
			return "'" + time.Time{}.Local().Format(time.RFC3339) + "'"
		default:
			return "''"
		}
	// case reflect.Uintptr:
	// case reflect.Complex64:
	// case reflect.Complex128:
	// case reflect.Array:
	// case reflect.Chan:
	// case reflect.Func:
	// case reflect.Interface:
	// case reflect.Map:
	// case reflect.Ptr:
	// case reflect.Slice:
	// case reflect.UnsafePointer:
	default:
		return "''"
	}
}

func (a *SQLiteAdapter) createTableSQL() string {
	const space = ' '
	const tab = '\t'
	b := strings.Builder{}
	b.WriteString("CREATE TABLE IF NOT EXISTS ")
	b.WriteString(a.TableName)
	b.WriteString(" (\r\n")
	// lastIdx := len(a.tags) - 1
	for idx, tag := range a.tags {
		if idx > 0 {
			b.Write([]byte{',', '\n'})
		}
		b.WriteByte(tab)
		b.WriteString(tag.FieldName)
		b.WriteByte(space)
		b.WriteString(tag.DataType)
		if strings.ToLower(tag.FieldName) == "id" {
			b.WriteString(" PRIMARY KEY ASC")
			continue
		}
		b.WriteString(" DEFAULT ")
		if tag.Default != "" {
			b.WriteByte('(')
			b.WriteString(tag.Default)
			b.WriteByte(')')
		} else {
			b.WriteString(defaultForSQLNull(tag))
		}
		// b.Write(ret)
		// if idx == lastIdx {
		// 	b.Write(ret)
		// } else {
		// 	b.Write([]byte{',', '\r', '\n'})
		// }
	}
	b.Write([]byte{')', ';'})
	return b.String()
}

func (a *SQLiteAdapter) createIndexSQLs() []string {
	sqlcmds := make([]string, 0, len(a.tags))
	const space = ' '
	b := strings.Builder{}
	for _, tag := range a.tags {
		if tag.Index == "" {
			continue
		}
		b.Reset()
		b.WriteString("CREATE INDEX IF NOT EXISTS Idx")
		b.WriteString(a.TableName)
		b.WriteString(strings.ToTitle(tag.FieldName))
		b.WriteString(" ON ")
		b.WriteString(a.TableName)
		b.WriteByte('(')
		b.WriteString(tag.FieldName)
		if strings.ToLower(tag.Index) == "desc" {
			b.WriteString(" DESC")
		}
		b.WriteByte(')')
		sqlcmds = append(sqlcmds, b.String())
	}
	return sqlcmds
}

//buildQuerySQL - generate query sql, keys => list of field names in where term
func (a *SQLiteAdapter) buildQuerySQL(keys []string, orderby string, paged bool) string {
	var lastIdx int
	var lastKeyIdx int
	b1 := strings.Builder{}
	b1.WriteString("SELECT ")
	lastIdx = len(a.tags) - 1
	for idx, tag := range a.tags {
		b1.WriteString(tag.FieldName)
		if idx != lastIdx {
			b1.WriteByte(',')
		}
	}
	b1.WriteString(" FROM ")
	b1.WriteString(a.TableName)
	if len(keys) == 0 {
		goto orderby
	}
	b1.WriteString(" WHERE ")
	lastKeyIdx = len(keys) - 1
	for idx, key := range keys {
		b1.WriteString(key)
		b1.Write([]byte{'=', '?'})
		if idx != lastKeyIdx {
			b1.WriteString(" AND ")
		}
	}

orderby:
	if orderby != "" {
		b1.WriteString(" ORDER BY ")
		b1.WriteString(orderby)
	}
	if !paged {
		b1.WriteByte(';')
		return b1.String()
	}
	b1.WriteString(" LIMIT ? OFFSET ?;")
	return b1.String()
}

//buildCountSQL - generate query sql, keys => list of field names in where term
func (a *SQLiteAdapter) buildCountSQL(keys []string) string {
	b1 := strings.Builder{}
	b1.WriteString("SELECT count(*) FROM ")
	b1.WriteString(a.TableName)
	if len(keys) == 0 {
		b1.WriteByte(';')
		return b1.String()
	}
	b1.WriteString(" WHERE ")
	lastKeyIdx := len(keys) - 1
	for idx, key := range keys {
		b1.WriteString(key)
		b1.Write([]byte{'=', '?'})
		if idx != lastKeyIdx {
			b1.WriteString(" AND ")
		}
	}
	b1.WriteByte(';')
	return b1.String()
}

func (a *SQLiteAdapter) selectOneSQL() string {
	b1 := strings.Builder{}
	b1.WriteString("SELECT ")
	lastIdx := len(a.tags) - 1
	for idx, tag := range a.tags {
		b1.WriteString(tag.FieldName)
		if idx != lastIdx {
			b1.WriteByte(',')
		}
	}
	b1.WriteString(" FROM ")
	b1.WriteString(a.TableName)
	b1.WriteString(" WHERE id=?;")
	return b1.String()
}

func (a *SQLiteAdapter) insertSQL(fields []string) string {
	b1 := strings.Builder{}
	b2 := strings.Builder{}
	b1.WriteString("INSERT INTO ")
	b2.WriteString(" VALUES (")
	b1.WriteString(a.TableName)
	b1.WriteString(" (")
	lastIdx := len(fields) - 1
	for idx, field := range fields {
		b1.WriteString(field)
		b2.WriteByte('?')
		if idx != lastIdx {
			b1.WriteByte(',')
			b2.WriteByte(',')
		}
	}
	b1.Write([]byte{')'})
	b2.Write([]byte{')', ';'})
	return b1.String() + b2.String()
}

func (a *SQLiteAdapter) setSQL(fields []string) string {
	b1 := strings.Builder{}
	b1.WriteString("UPDATE ")
	b1.WriteString(a.TableName)
	b1.WriteString(" SET ")
	l1 := len(fields) - 1
	for i, f := range fields {
		b1.WriteString(f)
		b1.Write([]byte{'=', '?'})
		if i != l1 {
			b1.WriteByte(',')
		}
	}
	b1.WriteString(" WHERE id=?;")
	return b1.String()
}

func (a *SQLiteAdapter) updateSQL(fields, where []string) string {
	b1 := strings.Builder{}
	b1.WriteString("UPDATE ")
	b1.WriteString(a.TableName)
	b1.WriteString(" SET ")
	l1 := len(fields) - 1
	for i, f := range fields {
		b1.WriteString(f)
		b1.Write([]byte{'=', '?'})
		if i != l1 {
			b1.WriteByte(',')
		}
	}
	if len(where) == 0 {
		b1.WriteByte(';')
		return b1.String()
	}
	l2 := len(where) - 1
	b1.WriteString(" WHERE ")
	for i, f := range where {
		b1.WriteString(f)
		b1.Write([]byte{'=', '?'})
		if i != l2 {
			b1.WriteString(" AND ")
		}
	}
	b1.WriteByte(';')
	return b1.String()
}

func (a *SQLiteAdapter) deleteSQL() string {
	b1 := strings.Builder{}
	b1.WriteString("DELETE FROM ")
	b1.WriteString(a.TableName)
	b1.WriteString(" WHERE id=?;")
	return b1.String()
}

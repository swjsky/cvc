package restful

import (
	"database/sql"
	"reflect"
	"sync"
	"testing"
)

func TestNewSQLiteAdapter(t *testing.T) {
	type modelTemplateStruct struct {
		ID      int    `name:"id" type:"INTEGER"`
		Name    string `type:"TEXT"`
		Phone   string `type:"TEXT"`
		Address string `type:"TEXT"`
	}
	template := modelTemplateStruct{}
	type args struct {
		db            *sql.DB
		mutex         *sync.RWMutex
		tableName     string
		modelTemplate interface{}
	}
	tests := []struct {
		name string
		args args
		want *SQLiteAdapter
	}{
		// TODO: Add test cases.
		{
			name: "case1",
			args: args{
				db:            nil,
				mutex:         nil,
				tableName:     "TestTable",
				modelTemplate: template,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSQLiteAdapter(tt.args.db, tt.args.mutex, tt.args.tableName, tt.args.modelTemplate); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSQLiteAdapter() = %v, want %v", got, tt.want)
			}
		})
	}
}

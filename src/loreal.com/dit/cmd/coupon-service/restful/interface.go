package restful

import (
	"net/url"
)

//Inserter - can insert
type Inserter interface {
	Insert(data url.Values) (id int64, err error)
}

//Deleter - can delete
type Deleter interface {
	Delete(id int64) (rowsAffected int64, err error)
}

//Querier - can run query
type Querier interface {
	Find(query url.Values) (total int64, records []*map[string]interface{})
	FindByID(id int64) map[string]interface{}
}

//Setter - can update by ID
type Setter interface {
	Set(id int64, data url.Values) error
}

//Updater - can update by condition
type Updater interface {
	Update(data url.Values, where url.Values) (rowsAffected int64, err error)
}

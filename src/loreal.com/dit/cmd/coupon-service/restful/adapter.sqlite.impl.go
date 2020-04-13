package restful

import (
	"database/sql"
	"log"
	"net/url"
	"strconv"
)

//Find - implementation of restful model interface
func (a *SQLiteAdapter) Find(query url.Values) (total int64, records []*map[string]interface{}) {
	orderby := query.Get("orderby")
	if orderby == "" {
		orderby = "id"
	}
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))
	paged := limit > 0

	keys, values := a.parseParams(query)
	countSQL := a.buildCountSQL(keys)
	sql := a.buildQuerySQL(keys, orderby, paged)
	if DEBUG {
		log.Printf("[DEBUG] - [INFO] Count SQL: %s\n", countSQL)
		log.Printf("[DEBUG] - [INFO] SQL: %s\n", sql)
	}
	if paged {
		values = append(values, limit, offset)
	}
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	if err := a.DB.QueryRow(countSQL, values...).Scan(&total); err != nil {
		log.Printf("[ERR] - [Count SQL]: %s, err: %v\n", countSQL, err)
	}
	records = a.scanRows(a.DB.Query(sql, values...))
	return
}

//FindByID - implementation of restful model interface
func (a *SQLiteAdapter) FindByID(id int64) map[string]interface{} {
	buffer := make([]interface{}, 0, len(a.tags))
	record := a.newRecord(&buffer)
	a.Mutex.Lock()
	defer a.Mutex.Unlock()
	row := a.stmts["one"].QueryRow(id)
	if err := row.Scan(buffer...); err != nil {
		log.Println(err)
		return nil
	}
	return record
}

//Insert - implementation of restful model interface
func (a *SQLiteAdapter) Insert(data url.Values) (id int64, err error) {
	keys, values := a.parseParams(data)
	for i, key := range keys {
		if key == "id" {
			values[i] = sql.NullInt64{Valid: false}
		}
	}
	r, err := a.DB.Exec(a.insertSQL(keys), values...)
	if err != nil {
		return -1, err
	}
	newID, err := r.LastInsertId()
	if err != nil {
		return -1, err
	}
	return newID, nil
}

//Set - implementation of restful model interface
func (a *SQLiteAdapter) Set(id int64, data url.Values) error {
	keys, values := a.parseParams(data)
	values = append(values, id)
	sql := a.setSQL(keys)
	if DEBUG {
		log.Printf("[DEBUG] - [INFO] SQL: %s\n", sql)
	}
	_, err := a.DB.Exec(a.setSQL(keys), values...)
	if err != nil {
		return err
	}
	return nil
}

//Update - implementation of restful model interface
func (a *SQLiteAdapter) Update(data url.Values, where url.Values) (rowsAffected int64, err error) {
	keys, values := a.parseParams(data)
	whereKeys, whereValues := a.parseParams(where)
	values = append(values, whereValues...)
	sql := a.updateSQL(keys, whereKeys)
	if DEBUG {
		log.Printf("[DEBUG] - [INFO] SQL: %s\n", sql)
	}
	r, err := a.DB.Exec(sql, values...)
	if err != nil || r == nil {
		return 0, err
	}
	return r.RowsAffected()
}

//Delete - implementation of restful model interface
func (a *SQLiteAdapter) Delete(id int64) (rowsAffected int64, err error) {
	r, err := a.stmts["delete"].Exec(id)
	if err != nil || r == nil {
		return 0, err
	}
	return r.RowsAffected()
}

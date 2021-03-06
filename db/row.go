package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type typeRow struct {
	*sql.Row
	t *typeTable
}

func (r *typeRow) Scan(dest ...interface{}) error {
	scans := r.t.makeNullableScans()
	err := r.Row.Scan(scans...)
	if err != nil {
		return err
	}
	for i := range dest {
		if dest[i] == nil {
			continue
		}
		err = convertValue(dest[i], scans[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *typeRow) Struct(dest interface{}) error {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr {
		return newErrorf("db: the object (%s) is not a pointer", rv.Kind())
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return newErrorf("db: the pointer (%s) is not point to a struct object", rv.Kind())
	}
	if rv.NumField() != r.t.ColumnNumbers {
		return newErrorf("db: the object field numbers (%d) not equals table column numbers (%d)", rv.NumField(), r.t.ColumnNumbers)
	}

	var err error
	var scans = r.t.makeNullableScans()
	if err = r.Row.Scan(scans...); err != nil {
		return err
	}
	for i := range scans {
		if err = convertValue(rv.Field(i).Addr().Interface(), scans[i]); err != nil {
			return err
		}
	}
	return nil
}

func (r *typeRow) Slice() ([]interface{}, error) {
	scans := r.t.makeNullableScans()
	err := r.Row.Scan(scans...)
	if err != nil {
		return nil, err
	}
	return r.t.parseSlice(scans), nil
}

func (r *typeRow) Map() (map[string]interface{}, error) {
	scans := r.t.makeNullableScans()
	err := r.Row.Scan(scans...)
	if err != nil {
		return nil, err
	}
	return r.t.parseMap(scans), nil
}

type typeRows struct {
	*sql.Rows
	t     *typeTable
	scans []interface{}
}

func (rs *typeRows) Scan(dest ...interface{}) error {
	err := rs.Rows.Scan(rs.scans...)
	if err != nil {
		return err
	}
	for i := range dest {
		if dest[i] == nil {
			continue
		}
		err = convertValue(dest[i], rs.scans[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (rs *typeRows) Struct(dest interface{}) error {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr {
		return newErrorf("db: the object (%s) is not a pointer", rv.Kind())
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return newErrorf("db: the pointer (%s) is not point to a struct object", rv.Kind())
	}
	if rv.NumField() != rs.t.ColumnNumbers {
		return newErrorf("db: the object field numbers (%d) not equals table column numbers (%d)", rv.NumField(), rs.t.ColumnNumbers)
	}

	var err error
	if err = rs.Rows.Scan(rs.scans...); err != nil {
		return err
	}
	for i := range rs.scans {
		if err = convertValue(rv.Field(i).Addr().Interface(), rs.scans[i]); err != nil {
			return err
		}
	}
	return nil
}

func (rs *typeRows) Slice() ([]interface{}, error) {
	err := rs.Rows.Scan(rs.scans...)
	if err != nil {
		return nil, err
	}
	return rs.t.parseSlice(rs.scans), nil
}

func (rs *typeRows) Map() (map[string]interface{}, error) {
	err := rs.Rows.Scan(rs.scans...)
	if err != nil {
		return nil, err
	}
	return rs.t.parseMap(rs.scans), nil
}

type typeSetter struct {
	t     *typeTable
	query string
	args  []interface{}
}

func (s *typeSetter) Values(values ...interface{}) (int64, error) {
	listkey := make([]string, 0)
	listvalue := make([]interface{}, 0)
	for i := range values {
		if values[i] == nil {
			continue
		}
		listkey = append(listkey, s.t.Columns[i].FullName+"=?")
		listvalue = append(listvalue, values[i])
	}
	strSql := fmt.Sprintf("%s SET %s %s", s.t.sqlUpdate, strings.Join(listkey, ", "), s.query)
	res, err := dbExec(strSql, append(listvalue, s.args...)...)
	if err != nil {
		return -1, err
	}
	return res.RowsAffected()
}

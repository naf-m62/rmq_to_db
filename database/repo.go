package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

type (
	Repo interface {
		Save(ctx context.Context, fieldMap map[string]string, valueMap map[string]interface{}) error
		SaveBatch(ctx context.Context, fieldMap map[string]string, valueMapList []map[string]interface{}) (err error)
	}
	repo struct {
		db        *sql.DB
		tableName string
	}
)

func (r *repo) Save(ctx context.Context, fieldMap map[string]string, valueMap map[string]interface{}) (err error) {
	var (
		fieldList []string
		argList   []interface{}
		countList []string
		i         int
	)
	for k, fn := range fieldMap {
		var (
			v  interface{}
			ok bool
		)
		if v, ok = valueMap[k]; !ok {
			continue
		}
		i++
		fieldList = append(fieldList, fn)
		countList = append(countList, "$"+strconv.Itoa(i))

		if _, ok := v.([]interface{}); ok {
			argList = append(argList, pq.Array(v))
		} else {
			argList = append(argList, v)
		}
	}
	var query = fmt.Sprintf("INSERT INTO %s(%s) VALUES (%s)",
		r.tableName,
		strings.Join(fieldList, ","),
		strings.Join(countList, ","),
	)
	err = r.db.QueryRowContext(ctx, query, argList...).Err()
	return err
}

func (r *repo) SaveBatch(
	ctx context.Context, fieldMap map[string]string, valueMapList []map[string]interface{},
) (err error) {
	var (
		valueQueryList []string
		argList        []interface{}
	)

	var orderedFields, fieldDb []string
	for fmk, fmv := range fieldMap {
		orderedFields = append(orderedFields, fmk)
		fieldDb = append(fieldDb, fmv)
	}

	var all = 1
	for _, valueMap := range valueMapList {
		var orderedArgs = getOrderValueList(orderedFields, valueMap)
		argList = append(argList, orderedArgs...)

		var count = getCount(all, len(orderedFields))
		all += len(orderedFields)

		var valueQuery = fmt.Sprintf("(%s)", strings.Join(count, ","))
		valueQueryList = append(valueQueryList, valueQuery)
	}

	var query = fmt.Sprintf("INSERT INTO %s(%s) VALUES ",
		r.tableName,
		strings.Join(fieldDb, ","),
	)

	query += strings.Join(valueQueryList, ",")

	err = r.db.QueryRowContext(ctx, query, argList...).Err()
	if pqErr, ok := err.(pq.Error); ok {
		return errors.New("code name " + pqErr.Code.Name() + " details " + pqErr.Detail)
	}
	return err
}

func getOrderValueList(orderedFields []string, valueMap map[string]interface{}) (orderedArgs []interface{}) {
	for _, f := range orderedFields {
		var (
			v  interface{}
			ok bool
		)
		if v, ok = valueMap[f]; !ok {
			orderedArgs = append(orderedArgs, nil)
			continue
		}

		if _, ok := v.([]interface{}); ok {
			orderedArgs = append(orderedArgs, pq.Array(v))
		} else {
			orderedArgs = append(orderedArgs, v)
		}

	}
	return orderedArgs
}

func getCount(all int, fill int) (countList []string) {
	for j := all; j < fill+all; j++ {
		countList = append(countList, "$"+strconv.Itoa(j))
	}
	return countList
}

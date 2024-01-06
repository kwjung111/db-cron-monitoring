package util

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

func GetSQLValueStr(nullString sql.NullString) string {
	if nullString.Valid {
		return nullString.String
	}
	return "null"
}

// execute sql, returns row
func GetResultfromDB(db *sql.DB, query string) ([]map[string]string, error) {

	rows, err := db.Query(query)
	if err != nil {
		return nil, errors.Wrap(err, "query Error")
	}

	data, err := RowsToObject(rows)
	if err != nil {
		return nil, errors.Wrap(err, "parsing Error")
	}

	defer rows.Close()

	return data, nil
}

// row 를 []Map[string]string 형태로 반환
func RowsToObject(rows *sql.Rows) ([]map[string]string, error) {

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var objects []map[string]string

	for rows.Next() {
		columnsData := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))

		for i := range columnsData {
			columnPointers[i] = &columnsData[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]string)

		for i, col := range columnsData {
			switch v := col.(type) {
			case []byte:
				// 바이트 슬라이스를 문자열로 변환
				rowMap[columns[i]] = string(v)
			case nil:
				rowMap[columns[i]] = ""
			default:
				// 기타 타입 처리
				rowMap[columns[i]] = fmt.Sprintf("%v", col)
			}
		}

		objects = append(objects, rowMap)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return objects, nil
}

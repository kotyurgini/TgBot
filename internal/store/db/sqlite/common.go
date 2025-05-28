package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
)

func placeholder(n int) string {
	return "?" + fmt.Sprint(n)
}

func placeholdersRange(start, count int) string {
	list := []string{}
	for i := 0; i < count; i++ {
		list = append(list, placeholder(start+i))
	}
	return strings.Join(list, ", ")
}

func placeholders(n int) string {
	list := []string{}
	for i := 0; i < n; i++ {
		list = append(list, placeholder(i+1))
	}
	return strings.Join(list, ", ")
}

func closeRows(rows *sql.Rows) {
	_ = rows.Close()
}

func dateLayout() string {
	return "2006-01-02 15:04:05.9999999 -0700 MST"
}

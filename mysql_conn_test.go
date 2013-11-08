package db

import (
	"database/sql"
	. "github.com/acsellers/assert"
	"testing"
)

func TestUserMapping(t *testing.T) {
	Within(t, func(test *Test) {

		db, err := sql.Open("mysql", mysqlConnectionString())
		if err == nil {
			conn := new(Connection)
			conn.DB = db
			conn.Dialect = newMysql()
			conn.Config = NewSimpleConfig()
			conn.sources = make(map[string]*source)
			conn.mappedStructs = make(map[string]*source)
			conn.mappableStructs = make(map[string][]*source)
			if !verifyTableExists(db, test) {
				t.Error("Could not locate testing tables")
			}

			test.Section("Dialect")
			cols := conn.Dialect.ColumnsInTable(conn, "db_test", "user")
			verifyUserTable(test, cols)

			test.Section("Mapping")
			mapper, err := conn.CreateMapper("User", &user{})
			if err != nil {
				test.T.Error(err)
			}
			verifyMapper(test, mapper)

		} else {
			t.Error(err)
		}
	})
}

func verifyTableExists(db *sql.DB, test *Test) bool {
	test.Section("Verify Test Database Exists")
	rows, err := db.Query("SHOW TABLES FROM db_test")
	if err != nil {
		test.T.Error(err)
	}
	defer rows.Close()
	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		if err != nil {
			test.T.Error(err)
		}
		if table == "user" {
			return true
		}
	}

	return false
}

func verifyUserTable(test *Test, cols map[string]*ColumnInfo) {
	test.AreEqual(6, len(cols))
	if id, ok := cols["id"]; ok {
		test.AreEqual(id.Name, "id")
		test.AreEqual(id.SqlTable, "user")
		test.AreEqual(id.SqlColumn, "id")
		test.AreEqual(id.SqlType, "int")
		test.AreEqual(id.Length, 255)
		test.AreEqual(id.Number, 0)
	} else {
		test.T.Error("id column doesn't exist")
	}
	if name, ok := cols["name"]; ok {
		test.AreEqual(name.Name, "name")
		test.AreEqual(name.SqlTable, "user")
		test.AreEqual(name.SqlType, "varchar")
		test.AreEqual(name.Length, 255)
		test.AreEqual(name.Number, 1)
	} else {
		test.T.Error("name column doesn't exist")
	}
	if story, ok := cols["story"]; ok {
		test.AreEqual(story.SqlType, "text")
		test.AreEqual(story.Length, 0)
		test.AreEqual(story.Number, 4)
	} else {
		test.T.Error("story column doesn't exist")
	}
}

func verifyMapper(test *Test, mapper Mapper) {
	test.AreEqual(mapper.(*source).Name, "User")
	test.AreEqual(len(mapper.(*source).Fields), 8)
	test.AreEqual(mapper.(*source).ColNum, 6)
	columnMappings := map[string][]string{
		"Id":       []string{"id", "int"},
		"Name":     []string{"name", "varchar"},
		"Email":    []string{"email", "varchar"},
		"Password": []string{"password", "varchar"},
		"Story":    []string{"story", "text"},
	}

	for _, field := range mapper.(*source).Fields[0:5] {
		col := columnMappings[field.structOptions.Name]
		test.AreEqual(col[0], field.ColumnInfo.Name)
		test.AreEqual(col[1], field.ColumnInfo.SqlType)
	}
}

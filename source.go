package db

import (
	"database/sql"
	"fmt"
	"reflect"
)

type source struct {
	ID          *sourceMapping
	Name        string
	FullName    string
	SqlName     string
	ColNum      int
	hasMixin    bool
	mixinField  int
	multiMapped bool
	config      *Config
	conn        *Connection
	Fields      []*sourceMapping

	structName, tableName string
}

type sourceMapping struct {
	*structOptions
	*columnInfo
}

func (sm *sourceMapping) Column() string {
	return sm.SqlTable + "." + sm.SqlColumn
}

type planner struct {
	scanners []interfaceable
}

func (s *source) mapPlan(v reflector) *planner {
	p := &planner{[]interfaceable{}}

	for _, col := range s.Fields {
		if col.columnInfo != nil && col.structOptions != nil {
			p.scanners = append(
				p.scanners,
				&reflectScanner{parent: v, index: col.columnInfo.Number},
			)
		}
	}

	return p
}

func (s *source) selectColumns() []string {
	output := []string{}
	for _, col := range s.Fields {
		if col.columnInfo != nil && col.structOptions != nil {
			output = append(
				output,
				fmt.Sprintf("%s.%s", s.SqlName, col.columnInfo.SqlColumn),
			)
		}
	}
	return output
}

func (p *planner) Items() []interface{} {
	output := make([]interface{}, len(p.scanners))
	for i, _ := range output {
		output[i] = p.scanners[i].iface()
	}

	return output
}

type reflectScanner struct {
	parent reflector
	index  int
}

type interfaceable interface {
	iface() interface{}
}

type reflector struct {
	item reflect.Value
}

func (rf *reflectScanner) iface() interface{} {
	return rf.parent.item.Elem().Field(rf.index).Addr().Interface()
}

type structOptions struct {
	Name       string
	FullName   string
	Index      int
	Kind       reflect.Kind
	ColumnHint string
	Options    map[string]interface{}
}

type columnInfo struct {
	Name      string
	SqlTable  string
	SqlColumn string
	SqlType   string
	Length    int
	Nullable  bool
	Number    int
}

func (s *source) runQuery(query string, values []interface{}) (*sql.Rows, error) {
	return s.conn.DB.Query(query, values...)
}

func (s *source) runQueryRow(query string, values []interface{}) *sql.Row {
	return s.conn.DB.QueryRow(query, values...)
}

func (s *source) runExec(query string, values []interface{}) (sql.Result, error) {
	return s.conn.DB.Exec(query, values...)
}

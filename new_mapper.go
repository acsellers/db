package ar

import (
	"reflect"
	"strconv"
)

func (c *Connection) CreateMapper(mapee interface{}) (*Mapper, error) {
	name, mapper := c.createMapperForPtr(mapee)
	c.models[name] = mapper

	return mapper, nil
}

func (c *Connection) CreateMapperPlus(v interface{}) {
	rv := reflect.ValueOf(v).Elem()
	fv := rv.Field(0)
	mp := new(MapperPlus)
	mp.model = createModelFor(v)
	vmp := reflect.ValueOf(mp)
	if fv.Type().Kind() == reflect.Ptr {
		fv.Set(vmp)
	}
}

func (c *Connection) createMapperForPtr(ptr interface{}) (string, *Mapper) {
	return "model", new(Mapper)
}

func (c *Connection) createModelFor(ptr interface{}) *model {
	newModel := &model{
		pk:      nil,
		table:   c.tableNameFor(ptr),
		Fields:  c.fieldsFor(ptr),
		indexes: Indexes{},
	}

	return newModel
}

func (c *Connection) tableNameFor(ptr interface{}) string {
	if t, ok := ptr.(string); ok {
		return t
	}

	t := reflect.TypeOf(ptr).Elem()
	if t.Kind() != reflect.Struct {
		return ""
	}

	return c.Config.StructToTable(t.Name())
}

func (c *Connection) fieldsFor(ptr interface{}) []*fieldInfo {
	fields := make([]*fieldInfo, 0, reflect.TypeOf(ptr).Elem().NumField())
	ptrType := reflect.TypeOf(ptr).Elem()

	for i := 0; i < ptrType.NumField(); i++ {
		currentField := ptrType.Field(i)

		switch currentField.Type.Kind() {
		case reflect.Ptr, reflect.Map:
			continue
		case reflect.Slice:
			if currentField.Type.Elem().Kind() != reflect.Uint8 {
				continue
			}
		}

		fieldInfo := c.parseStructTags(currentField.Name, currentField.Tag)
		if fieldInfo.Valid {
			fields = append(fields, fieldInfo)
		}
	}

	return fields
}

func (c *Connection) parseStructTags(name string, tags reflect.StructTag) *fieldInfo {
	pk := name == c.Config.IdName || tags.Get("ar_pk") == "true" || tags.Get("pk") == "true"
	fk := tags.Get("ar_fk") == "true" || tags.Get("fk") == "true"
	var size int
	if explicitSize := tags.Get("ar_size"); explicitSize != "" {
		eSize, err := strconv.ParseInt(explicitSize, 10, 32)
		if err != nil {
			size = int(eSize)
		}
	} else if explicitSize := tags.Get("size"); explicitSize != "" {
		eSize, err := strconv.ParseInt(explicitSize, 10, 32)
		if err != nil {
			size = int(eSize)
		}
	}

	return &fieldInfo{
		PrimaryKey: pk,
		ForeignKey: fk,
		Size:       size,
	}

}

type fieldInfo struct {
	Name       string
	CamelName  string
	PrimaryKey bool
	ForeignKey bool
	Size       int
	Default    interface{}
	Join       bool
	Index      bool
	Unique     bool
	NotNull    bool
	Valid      bool
}
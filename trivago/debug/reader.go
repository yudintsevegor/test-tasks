package debug

import (
	"log"
	"reflect"
)

type Iterator struct {
	AllowEmptyFields         bool
	IgnoreFieldCountMismatch bool

	r       Reader
	headers []string
	line    int
	fields  []reflect.Value
	kinds   []ValueKind
	tags    []int
	err     error
}

type Reader interface {
	Read() ([]string, error)
}

func Headers(r Reader) ([]string, error) {
	return r.Read()
}

type ValueKind int

const (
	csvField = "csvField"

	unKnownKind ValueKind = iota
	boolKind
	intKind
	uintKind
	floatKind
	stringKind
)

func NewCSVIterator(r Reader, headers []string) *Iterator {
	return &Iterator{
		AllowEmptyFields:         false,
		IgnoreFieldCountMismatch: false,
		r:                        nil,
		headers:                  headers,
		line:                     0,
		fields:                   nil,
		kinds:                    nil,
		tags:                     nil,
		err:                      nil,
	}
}

func (iter *Iterator) GetHeaderFieldRelation(v interface{}) {
	vType := reflect.TypeOf(v).Elem()
	vValue := reflect.ValueOf(v).Elem()
	numFields := vType.NumField()

	iter.kinds = make([]ValueKind, numFields)
	iter.tags = make([]int, numFields)
	iter.fields = make([]reflect.Value, numFields)

	headerMap := make(map[string]int, len(iter.headers))
	for column, header := range iter.headers {
		headerMap[header] = column
	}

	for i := 0; i < numFields; i++ {
		field := vType.Field(i)
		tag := field.Tag.Get(csvField)
		column, ok := headerMap[tag]

		if len(tag) == 0 {
			continue
		} else if !ok {
			continue
		}

		kind := unKnownKind
		switch field.Type.Kind() {
		case reflect.Bool:
			kind = boolKind
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			kind = intKind
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			kind = uintKind
		case reflect.Float32, reflect.Float64:
			kind = floatKind
		case reflect.String:
			kind = stringKind
		}

		iter.kinds[i] = kind
		iter.tags[i] = column
		iter.fields[i] = vValue.Field(i)
	}

	for i := 0; i < len(iter.kinds); i++ {
		log.Printf("Field: %v; Header: %v", iter.kinds[i], iter.headers[i])
	}
}

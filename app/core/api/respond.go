package api

import (
	"encoding"
	"encoding/json"
	"reflect"

	"github.com/gin-gonic/gin"
)

const maxDepth = 32

var jsonMarshaler = reflect.TypeFor[json.Marshaler]()

var textMarshaler = reflect.TypeFor[encoding.TextMarshaler]()

func JSON(c *gin.Context, status int, body any) {
	value := reflect.ValueOf(body)
	if value.IsValid() == false {
		c.JSON(status, body)

		return
	}

	c.JSON(status, withEmptyCollections(value, 0).Interface())
}

func withEmptyCollections(v reflect.Value, depth int) reflect.Value {
	kind := v.Kind()
	switch {
	case depth > maxDepth || marshalsItself(v.Type()):
		return v
	case kind == reflect.Slice && v.Type().Elem().Kind() != reflect.Uint8:
		out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for i := range v.Len() {
			out.Index(i).Set(withEmptyCollections(v.Index(i), depth+1))
		}

		return out
	case kind == reflect.Map:
		out := reflect.MakeMapWithSize(v.Type(), v.Len())
		for entry := v.MapRange(); entry.Next(); {
			out.SetMapIndex(entry.Key(), withEmptyCollections(entry.Value(), depth+1))
		}

		return out
	case kind == reflect.Pointer && v.IsNil() == false:
		out := reflect.New(v.Type().Elem())
		out.Elem().Set(withEmptyCollections(v.Elem(), depth+1))

		return out
	case kind == reflect.Interface && v.IsNil() == false:
		return withEmptyCollections(v.Elem(), depth+1)
	case kind == reflect.Struct:
		return structWithEmptyCollections(v, depth)
	default:
		return v
	}
}

func structWithEmptyCollections(v reflect.Value, depth int) reflect.Value {
	out := reflect.New(v.Type()).Elem()
	out.Set(v)

	for i := range v.NumField() {
		if field := out.Field(i); field.CanSet() {
			field.Set(withEmptyCollections(v.Field(i), depth+1))
		}
	}

	return out
}

func marshalsItself(t reflect.Type) bool {
	return t.Implements(jsonMarshaler) || t.Implements(textMarshaler) ||
		reflect.PointerTo(t).Implements(jsonMarshaler) || reflect.PointerTo(t).Implements(textMarshaler)
}

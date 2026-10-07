package api

import (
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

const KeyRequired Key = "validation.required"

const KeyEmail Key = "validation.email"

const KeyUUID Key = "validation.uuid"

const KeyMinLength Key = "validation.min_length"

const KeyMaxLength Key = "validation.max_length"

const KeyMinItems Key = "validation.min_items"

const KeyMaxItems Key = "validation.max_items"

const KeyMin Key = "validation.min"

const KeyMax Key = "validation.max"

const KeyOneOf Key = "validation.one_of"

const KeyInvalid Key = "validation.invalid"

var countedKinds = []reflect.Kind{reflect.Slice, reflect.Array, reflect.Map}

var wireNames sync.Once

func useWireNames() {
	if engine, ok := binding.Validator.Engine().(*validator.Validate); ok {
		engine.RegisterTagNameFunc(func(field reflect.StructField) string {
			return wireName(field.Name, field.Tag)
		})
	}
}

func wireName(fieldName string, tag reflect.StructTag) string {
	name, _, _ := strings.Cut(tag.Get("json"), ",")
	if name == "" {
		name, _, _ = strings.Cut(tag.Get("form"), ",")
	}

	switch name {
	case "-":
		return ""
	case "":
		return fieldName
	default:
		return name
	}
}

func fieldErrors(fields validator.ValidationErrors) Errors {
	errs := make(Errors, len(fields))
	for _, field := range fields {
		errs[jsonPath(field)] = message(field)
	}

	return errs
}

func jsonPath(field validator.FieldError) string {
	_, path, _ := strings.Cut(field.Namespace(), ".")

	return path
}

func message(field validator.FieldError) Message {
	switch field.Tag() {
	case "required":
		return Message{Key: KeyRequired}
	case "email":
		return Message{Key: KeyEmail}
	case "uuid", "uuid4", "uuid7":
		return Message{Key: KeyUUID}
	case "min":
		return limit(field, KeyMinLength, KeyMinItems, KeyMin)
	case "max":
		return limit(field, KeyMaxLength, KeyMaxItems, KeyMax)
	case "oneof":
		return oneOf(strings.Fields(field.Param()))
	default:
		return Message{Key: KeyInvalid}
	}
}

func limit(field validator.FieldError, forLength, forItems, forValue Key) Message {
	key := forLength

	switch {
	case slices.Contains(countedKinds, field.Kind()):
		key = forItems
	case field.Kind() != reflect.String:
		return Message{Key: forValue, Params: map[string]any{field.Tag(): field.Param()}}
	}

	count, err := strconv.ParseInt(field.Param(), 0, 64)
	if err != nil {
		return Message{Key: KeyInvalid}
	}

	return Message{Key: key, Params: map[string]any{field.Tag(): count}}
}

func oneOf[S ~string](values []S) Message {
	names := make([]string, len(values))
	for i, value := range values {
		names[i] = string(value)
	}

	return Message{Key: KeyOneOf, Params: map[string]any{"values": strings.Join(names, ", ")}}
}

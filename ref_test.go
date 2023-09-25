package ginplus

import (
	"reflect"
	"testing"
)

func Test_getTag(t *testing.T) {
	type My struct {
		Name    string `json:"name"`
		Id      uint   `uri:"id"`
		Keyword string `form:"keyword"`
	}

	fieldList := getTag(reflect.TypeOf(&My{}))

	t.Log(fieldList)
}

func Test_getTag2(t *testing.T) {
	var b bool
	t.Logf("%T", b)
}

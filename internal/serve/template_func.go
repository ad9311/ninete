package serve

import (
	"html/template"
	"reflect"
	"strconv"
	"time"

	"github.com/ad9311/ninete/internal/prog"
)

func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"currency":  currency,
		"sumAmount": sumAmount,
		"timeStamp": timeStamp,
	}
}

func currency(v uint64) string {
	base := float64(v) / 100.00

	return "$" + strconv.FormatFloat(base, 'f', 2, 64)
}

func timeStamp(v int64) string {
	return prog.UnixToStringDate(v, time.DateOnly)
}

func sumAmount(rows any) uint64 {
	value := reflect.ValueOf(rows)
	if !value.IsValid() || value.Kind() != reflect.Slice {
		return 0
	}

	var total uint64
	for i := 0; i < value.Len(); i++ {
		item := value.Index(i)
		if item.Kind() == reflect.Pointer {
			if item.IsNil() {
				continue
			}
			item = item.Elem()
		}
		if item.Kind() != reflect.Struct {
			continue
		}

		amount := item.FieldByName("Amount")
		if !amount.IsValid() {
			continue
		}

		switch amount.Kind() {
		case reflect.Uint64:
			total += amount.Uint()
		}
	}

	return total
}

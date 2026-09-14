package obs

import (
	"reflect"
	"strings"
)

// Mask 按结构体 tag `pii:"mask"` 脱敏后返回可日志化的 map。
// 未标注字段原样保留；密码/密钥类仍应避免进入日志。
func Mask(v any) any {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return nil
	}
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return v
	}
	rt := rv.Type()
	out := make(map[string]any, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.PkgPath != "" {
			continue
		}
		name := jsonFieldName(f)
		if name == "-" {
			continue
		}
		fv := rv.Field(i)
		if tag := f.Tag.Get("pii"); tag == "mask" {
			out[name] = redacted
			continue
		}
		if fv.Kind() == reflect.Struct || (fv.Kind() == reflect.Pointer && !fv.IsNil() && fv.Elem().Kind() == reflect.Struct) {
			out[name] = Mask(fv.Interface())
			continue
		}
		if fv.CanInterface() {
			out[name] = fv.Interface()
		}
	}
	return out
}

func jsonFieldName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" {
		return f.Name
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		return f.Name
	}
	return name
}

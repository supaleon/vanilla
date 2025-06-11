package pages

import "testing"

type MyMap = map[string]string

type AnyMap = map[any]any

func TestName(t *testing.T) {
	m := make(map[string]string)
	m["a"] = "b"
	hello(m)
}

func hello(val any) {
	if v, ok := val.(MyMap); ok {
		println(v["a"])
	}
}

// <div Var1={}></div>
// export let MyMap = {} -> map[any]any
// export let MyList = [] -> []any
// export let TagList = make(TagList)
// export let TagList = new Struct()
// response.withLayout(template, data) -> map[string]any|struct
// response.toHtml(template, data)
// response.setData(data)

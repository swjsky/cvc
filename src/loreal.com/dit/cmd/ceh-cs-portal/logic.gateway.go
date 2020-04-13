package main

import "strings"

/* jsonBodyfilter - 用于过滤混淆bodyObject里的内容
bodyObject: 存放需混淆的 JSON 对象；
mixupFn： 混淆函数， 如果为nil则会删除相应的键值；
keys： 需要混淆或过滤的键；
*/
func jsonBodyfilter(
	bodyObject *map[string]interface{},
	mixupFn func(interface{}) interface{},
	keys ...string,
) {
	for k, v := range *bodyObject {
		switch val := v.(type) {
		case map[string]interface{}:
			jsonBodyfilter(&val, mixupFn, keys...)
		default:
			for _, hotKey := range keys {
				if k == hotKey {
					if mixupFn == nil {
						delete(*bodyObject, k)
						continue
					}
					(*bodyObject)[k] = mixupFn(val)
				}
			}

		}
	}
}

func mixupString(s interface{}) interface{} {
	switch value := s.(type) {
	case string:
		return strings.Repeat("*", len(value))
	default:
		return value
	}
}

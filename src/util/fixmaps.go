package util

import "fmt"

func FixMaps(v interface{}) interface{} {
	switch v := v.(type) {
	case map[interface{}]interface{}:
		return KeyifyMap(v)
	case map[string]interface{}:
		for k, item := range v {
			v[k] = FixMaps(item)
		}
		return v
	case []interface{}:
		for i, item := range v {
			v[i] = FixMaps(item)
		}
		return v
	}
	return v
}

func KeyifyMap(mii map[interface{}]interface{}) map[string]interface{} {
	msi := map[string]interface{}{}
	for k, item := range mii {
		msi[fmt.Sprintf("%v", k)] = item
	}
	return msi
}

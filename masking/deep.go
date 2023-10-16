//
//
// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 THL A29 Limited, a Tencent company.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.
//
//

// Package masking 敏感信息脱敏拦截器
package masking

import (
	"reflect"
	"time"
)

// DeepCheck 递归处理待脱敏数据
func DeepCheck(src interface{}) {
	if src == nil {
		return
	}

	original := reflect.ValueOf(src)

	checkRecursive(original)
}

// checkRecursive 调用masking插件导出的脱敏方法
func checkRecursive(original reflect.Value) {

	switch original.Kind() {
	case reflect.Ptr:
		originalValue := original.Elem()

		if !originalValue.IsValid() {
			return
		}
		checkRecursive(originalValue)

	case reflect.Interface:
		if original.IsNil() {
			return
		}
		originalValue := original.Elem()
		if v, ok := originalValue.Interface().(Masking); ok {
			v.Masking()
		}
		checkRecursive(originalValue)

	case reflect.Struct:
		_, ok := original.Interface().(time.Time)
		if ok {
			return
		}
		for i := 0; i < original.NumField(); i++ {
			if original.Type().Field(i).PkgPath != "" {
				continue
			}
			if v, ok := original.Field(i).Interface().(Masking); ok {
				v.Masking()
			}
			checkRecursive(original.Field(i))
		}

	case reflect.Slice:
		if original.IsNil() {
			return
		}
		for i := 0; i < original.Len(); i++ {
			if v, ok := original.Index(i).Interface().(Masking); ok {
				v.Masking()
			}
			checkRecursive(original.Index(i))
		}

	case reflect.Map:
		if original.IsNil() {
			return
		}
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			checkRecursive(originalValue)
			if v, ok := key.Interface().(Masking); ok {
				v.Masking()
			}
		}
	}
}

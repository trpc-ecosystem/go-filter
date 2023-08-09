// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package debuglog

// RuleItem is the basic configuration for a single rule.
type RuleItem struct {
	Method  *string
	Retcode *int
}

// Matched is the result of rule matching.
func (e RuleItem) Matched(destMethod string, destRetCode int) bool {
	return (e.Method == nil || *e.Method == destMethod) &&
		(e.Retcode == nil || *e.Retcode == destRetCode)
}

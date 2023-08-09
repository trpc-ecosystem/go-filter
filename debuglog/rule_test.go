// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package debuglog

import "testing"

func TestRuleItem_Matched(t *testing.T) {
	type fields struct {
		Method  *string
		Retcode *int
	}
	type args struct {
		destMethod  string
		destRetCode int
	}

	method := "ruleMethod"
	retCode := 123
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "only given method name | matched",
			fields: fields{Method: &method},
			args:   args{destMethod: method, destRetCode: 0},
			want:   true,
		},
		{
			name:   "only given method name | not matched",
			fields: fields{Method: &method},
			args:   args{destMethod: "other_method", destRetCode: retCode},
			want:   false,
		},
		{
			name:   "only given retcode | matched",
			fields: fields{Retcode: &retCode},
			args:   args{destMethod: "other_method", destRetCode: retCode},
			want:   true,
		},
		{
			name:   "only given retcode | not matched",
			fields: fields{Retcode: &retCode},
			args:   args{destMethod: method, destRetCode: 0},
			want:   false,
		},
		{
			name:   "all fields given | matched",
			fields: fields{Method: &method, Retcode: &retCode},
			args:   args{destMethod: method, destRetCode: retCode},
			want:   true,
		},
		{
			name:   "all fields given | not matched 1",
			fields: fields{Method: &method, Retcode: &retCode},
			args:   args{destMethod: method, destRetCode: 0},
			want:   false,
		},
		{
			name:   "all fields given | not matched 2",
			fields: fields{Method: &method, Retcode: &retCode},
			args:   args{destMethod: "other_method", destRetCode: retCode},
			want:   false,
		},
		{
			name:   "all fields given | not matched 3",
			fields: fields{Method: &method, Retcode: &retCode},
			args:   args{destMethod: "other_method", destRetCode: 0},
			want:   false,
		},
		{
			name:   "none fields given would always true",
			fields: fields{},
			args:   args{destMethod: method, destRetCode: retCode},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := RuleItem{
				Method:  tt.fields.Method,
				Retcode: tt.fields.Retcode,
			}
			if got := e.Matched(tt.args.destMethod, tt.args.destRetCode); got != tt.want {
				t.Errorf("RuleItem.Matched() = %v, want %v", got, tt.want)
			}
		})
	}
}

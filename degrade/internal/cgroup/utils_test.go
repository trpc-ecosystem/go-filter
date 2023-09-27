// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package cgroup

import (
	"testing"
)

func Test_readUint64FromFile(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name    string
		args    args
		wantRet uint64
		wantErr bool
	}{
		{
			name:    "",
			args:    args{"/sys/fs/cgroup/cpuacct/cpuacct.usage"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := readUint64FromFile(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("readUint64FromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func Test_readInt64FromFile(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name    string
		args    args
		wantRet int64
		wantErr bool
	}{
		{
			name:    "",
			args:    args{"/sys/fs/cgroup/cpuacct/cpuacct.usage"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := readInt64FromFile(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("readInt64FromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_readMapFromFile(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantRet map[string]uint64
		wantErr bool
	}{
		{
			args:    args{"/sys/fs/cgroup/memory/memory.stat"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := readMapFromFile(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("readMapFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func Test_readFromFile(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantRet string
		wantErr bool
	}{
		{
			args:    args{"/sys/fs/cgroup/memory/memory.stat"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := readFromFile(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("readFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package cgroup

import (
	"testing"
	"time"
)

func TestGetDockerCPUUsage(t *testing.T) {
	type args struct {
		interval time.Duration
	}
	tests := []struct {
		name      string
		args      args
		wantUsage float64
		wantErr   bool
	}{
		{
			args:    args{1 * time.Second},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUsage, err := GetDockerCPUUsage(tt.args.interval)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDockerCPUUsage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUsage <= 0 {
				t.Errorf("GetDockerCPUUsage() = %v, want %v", gotUsage, tt.wantUsage)
			}
		})
	}
}

func TestGetContainerCPUTotal(t *testing.T) {
	tests := []struct {
		name      string
		wantUsage uint64
		wantErr   bool
	}{
		{
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUsage, err := GetContainerCPUTotal()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetContainerCPUTotal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUsage < tt.wantUsage {
				t.Errorf("GetContainerCPUTotal() = %v, want %v", gotUsage, tt.wantUsage)
			}
		})
	}
}

func TestGetCoreCount(t *testing.T) {
	tests := []struct {
		name    string
		wantNum uint64
		wantErr bool
	}{
		{
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNum, err := GetCoreCount()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCoreCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotNum <= tt.wantNum {
				t.Errorf("GetCoreCount() = %v, want %v", gotNum, tt.wantNum)
			}
		})
	}
}

func TestGetLimitedCoreCount(t *testing.T) {
	tests := []struct {
		name    string
		wantMum float64
		wantErr bool
	}{
		{
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMum, err := GetLimitedCoreCount()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLimitedCoreCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMum <= tt.wantMum {
				t.Errorf("GetLimitedCoreCount() = %v, want %v", gotMum, tt.wantMum)
			}
		})
	}
}

func Test_getValidCPUSet(t *testing.T) {
	tests := []struct {
		name    string
		wantRet float64
		wantErr bool
	}{
		{
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRet, err := getValidCPUSet()
			if (err != nil) != tt.wantErr {
				t.Errorf("getValidCPUSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRet <= tt.wantRet {
				t.Errorf("getValidCPUSet() = %v, want %v", gotRet, tt.wantRet)
			}
		})
	}
}

func Test_getCPUTick(t *testing.T) {
	tests := []struct {
		name    string
		wantCnt int
		wantErr bool
	}{
		{
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCnt, err := getCPUTick()
			if (err != nil) != tt.wantErr {
				t.Errorf("getCPUTick() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCnt <= tt.wantCnt {
				t.Errorf("getCPUTick() = %v, want %v", gotCnt, tt.wantCnt)
			}
		})
	}
}

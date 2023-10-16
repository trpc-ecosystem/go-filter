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

package masking

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

var handler = func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
	return req, nil
}

type PbData struct {
	Data *MockRspData
}

type MockRspData struct {
	Uid      string
	SLice    []string
	Map      map[string]interface{}
	External interface{}
	Children *MockChildrenData
}

type MockChildrenData struct {
	Name string
}

func (m *MockRspData) Masking() {
	m.Uid = "masking_uid"
}

func (c *MockChildrenData) Masking() {
	c.Name = "masking_name"
}

func TestServerFilter(t *testing.T) {
	filter := ServerFilter()
	originData := PbData{
		Data: &MockRspData{
			Uid:   "123",
			SLice: []string{"a"},
			Map: map[string]interface{}{
				"k": "v",
			},
			Children: &MockChildrenData{
				Name: "masking",
			},
		},
	}
	expectRspData := PbData{
		Data: &MockRspData{
			Uid:   "masking_uid",
			SLice: []string{"a"},
			Map: map[string]interface{}{
				"k": "v",
			},
			Children: &MockChildrenData{
				Name: "masking_name",
			},
		},
	}
	rsp, err := filter(context.TODO(), originData, handler)
	assert.Nil(t, err)
	assert.Equal(t, expectRspData, rsp)
	_, err = filter(context.TODO(), nil, handler)
	assert.Nil(t, err)
}

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

package jwt

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// userInfo 用于签名的用户信息
type userInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Role int    `json:"role"`
}

func mockUserInfo() *userInfo {
	return &userInfo{ID: 100, Name: "forrestsun", Role: 1}
}
func mockSigner() Signer {
	return NewJwtSign([]byte("123456"), time.Hour, "issuer")
}

func TestJwtSign(t *testing.T) {
	user := mockUserInfo()
	sign := mockSigner()
	// verify fail
	_, err := sign.Verify("abc")
	assert.NotNil(t, err)
	// sign
	token, err := sign.Sign(user)
	assert.Nil(t, err)
	// verify ok
	data, err := sign.Verify(token)
	assert.Nil(t, err)
	// map[string]interface{}, map[id:100 name:forrestsun role:1]
	t.Logf("%v, %+v", reflect.TypeOf(data), data)
	// verify data
	ctx := context.WithValue(context.Background(), AuthJwtCtxKey, data)
	var ctxUser = &userInfo{}
	err = GetCustomInfo(ctx, ctxUser)
	assert.Nil(t, err)
	reflect.DeepEqual(user, ctxUser)
}

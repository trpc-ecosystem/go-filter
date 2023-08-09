// Tencent is pleased to support the open source community by making tRPC available.
// Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.

package jwt

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
)

var (
	// ErrInvalidToken 非法 token
	ErrInvalidToken = fmt.Errorf("invalid token")
)

// Signer 数字签名器接口
type Signer interface {
	Sign(custom interface{}) (string, error)
	Verify(token string) (interface{}, error)
}

// jwtSign jwt 用户身份数字签名器
type jwtSign struct {
	Secret  []byte
	Expired time.Duration
	Issuer  string
}

// NewJwtSign 构造 jwt 签名
func NewJwtSign(secret []byte, expired time.Duration, issuer string) Signer {
	return &jwtSign{
		Secret:  secret,
		Expired: expired,
		Issuer:  issuer,
	}
}

// claims 元数据
type claims struct {
	jwt.StandardClaims
	// 业务自定义的数据
	Custom interface{} `json:"custom,omitempty"`
}

// Sign 生成签名
func (t *jwtSign) Sign(custom interface{}) (string, error) {
	now := time.Now()
	cl := claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwt.At(now.Add(t.Expired)),
			IssuedAt:  jwt.At(now),
			Issuer:    t.Issuer,
		},
		Custom: custom,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, cl)
	return token.SignedString(t.Secret)
}

// Verify 校验 token, 成功则返回用户自定义数据
func (t *jwtSign) Verify(tokenStr string) (interface{}, error) {
	// 方法内部主要是具体的解码和校验的过程
	token, err := jwt.ParseWithClaims(tokenStr, &claims{},
		func(token *jwt.Token) (i interface{}, err error) {
			return t.Secret, nil
		})
	if err != nil {
		return nil, err
	}
	if claim, ok := token.Claims.(*claims); ok && token.Valid {
		return claim.Custom, nil
	}
	return nil, ErrInvalidToken
}

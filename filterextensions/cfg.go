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

package filterextensions

type cfg struct {
	Client []cfgService `yaml:"client"`
	Server []cfgService `yaml:"server"`
}

type cfgService struct {
	Name    string      `yaml:"name"`
	Methods []cfgMethod `yaml:"methods"`
}

type cfgMethod struct {
	Name    string   `yaml:"name"`
	Filters []string `yaml:"filters"`
}

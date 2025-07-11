//
//
// Tencent is pleased to support the open source community by making tRPC available.
//
// Copyright (C) 2023 Tencent.
// All rights reserved.
//
// If you have downloaded a copy of the tRPC source code from Tencent,
// please note that tRPC source code is licensed under the Apache 2.0 License,
// A copy of the Apache 2.0 License is included in this file.
//
//

package cpmsg

import (
	"fmt"

	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-utils/copyutils"
)

// CopyMsg is not a common Message copy util. It's specially customized for slime.
func CopyMsg(dst, src codec.Msg) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("CopyMsg paniced, this usually means slime may not support your protocol: %v", r)
		}
	}()

	oriReqHead := dst.ClientReqHead()
	oriRspHead := dst.ClientRspHead()

	codec.CopyMsg(dst, src)

	if oriReqHead != nil {
		// This must be copying back to user.
		// There will be no data access to src ReqHead. A shallow copy is just ok.
		dst.WithClientReqHead(oriReqHead)
		if err := copyutils.ShallowCopy(oriReqHead, src.ClientReqHead()); err != nil {
			return fmt.Errorf("failed to shallow copy back to original ClientReqHead, err: %w", err)
		}
	} else if src.ClientReqHead() != nil {
		// This must be copying to a new created msg. DeepCopy should be used to avoid data races.
		head, err := copyutils.DeepCopy(src.ClientReqHead())
		if err != nil {
			return fmt.Errorf("failed to deepcopy ClientReqHead, err: %w", err)
		}
		dst.WithClientReqHead(head)
	}

	if oriRspHead != nil {
		dst.WithClientRspHead(oriRspHead)
		if err := copyutils.ShallowCopy(oriRspHead, src.ClientRspHead()); err != nil {
			return fmt.Errorf("failed to shallow copy back to original ClientRspHead, err: %w", err)
		}
	} else if src.ClientRspHead() != nil {
		head, err := copyutils.DeepCopy(src.ClientRspHead())
		if err != nil {
			return fmt.Errorf("failed to deepcopy ClientRspHead, err: %w", err)
		}
		dst.WithClientRspHead(head)
	}

	return nil
}

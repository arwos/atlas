/*
 *  Copyright (c) 2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

// Package contract contains transport generation directives.
package contract

//go:generate goppy tb --mod=json-rpc-server --out=./../transport
//go:generate easyjson ./../transport/jsonrpc_server_model.go

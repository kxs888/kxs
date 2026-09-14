package migrations

import "embed"

// SQL 为 golang-migrate iofs 源。仅 schema；种子数据见 data/（运行时 argon2id 引导用户）。
//
//go:embed *.sql
var SQL embed.FS

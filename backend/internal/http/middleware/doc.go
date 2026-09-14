// Package middleware 实现 HTTP 包装白名单。见 docs/http-wrappers.md。
//
// 全局最多 4 层：Recover、RequestID+Trace、Limits、AccessLog。
// 第 5 层 SMCrypto 仅在 SM_CRYPTO_ENABLED=true 时挂到 /api/v1。
// RequireAuth 与幂等禁止 engine.Use 全局安装，仅 Group 局部挂载。
package middleware

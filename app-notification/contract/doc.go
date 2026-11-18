// Package contract 是通知中心对业务子系统暴露的契约。
// 只依赖标准库。业务方 import 本包后通过 Get 取得发送入口；未安装通知中心时 Get 返回 nil。
package contract

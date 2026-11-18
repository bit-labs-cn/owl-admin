package contract

import "sync"

var (
	registryMu sync.RWMutex
	current    Notifier
)

// Register 由通知中心在启动时调用。重复注册以最后一次为准。
func Register(n Notifier) {
	registryMu.Lock()
	current = n
	registryMu.Unlock()
}

// Get 返回已安装的通知入口。未安装时返回 nil，调用方应跳过发送。
func Get() Notifier {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return current
}

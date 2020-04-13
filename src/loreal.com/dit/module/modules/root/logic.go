package root

func (m *Module) setShutdownKey(key string) string {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	m.Context["shutdownKey"] = key
	return key
}

func (m *Module) verifyShutdownKey(key string) bool {
	m.Mutex.RLock()
	defer m.Mutex.RUnlock()
	if shutdownKey, ok := m.Context["shutdownKey"].(string); ok {
		return key != "" && key == shutdownKey
	}
	return false
}

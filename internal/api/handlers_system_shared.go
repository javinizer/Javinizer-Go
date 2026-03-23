package api

import "sync"

// Mutex to serialize config updates (prevents concurrent read-modify-write races)
var configMutex sync.Mutex

const defaultProxyTestURL = "https://javdb.com"

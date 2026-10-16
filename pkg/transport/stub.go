//go:build !darwin || !cgo

package transport

// On non-darwin OS or when CGO is disabled, defaultProvider remains nil
// unless set manually or provided by a platform-specific transport (e.g. future windows/linux bridge).

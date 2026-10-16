package transport

var defaultProvider Provider

// DefaultProvider returns the active platform Bluetooth transport provider.
func DefaultProvider() (Provider, error) {
	if defaultProvider == nil {
		return nil, ErrUnsupportedOS
	}
	return defaultProvider, nil
}

// SetDefaultProvider overrides the active transport provider (useful for testing or custom drivers).
func SetDefaultProvider(p Provider) {
	defaultProvider = p
}

package environment

import "sort"

// Keys returns a sorted slice of all configuration keys in the snapshot.
func (s *Snapshot) Keys() []string {
	keys := make([]string, 0, len(s.Data))
	for k := range s.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ToMap returns a copy of the snapshot's configuration data as a plain map.
func (s *Snapshot) ToMap() map[string]string {
	copy := make(map[string]string, len(s.Data))
	for k, v := range s.Data {
		copy[k] = v
	}
	return copy
}

// Get retrieves a single configuration value by key.
// Returns the value and a boolean indicating whether the key was found.
func (s *Snapshot) Get(key string) (string, bool) {
	v, ok := s.Data[key]
	return v, ok
}

// Size returns the number of configuration keys in the snapshot.
func (s *Snapshot) Size() int {
	return len(s.Data)
}

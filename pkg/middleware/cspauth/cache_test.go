package cspauth

import (
	"reflect"
	"sync"
	"testing"
)

func Test_cache_Get(t *testing.T) {
	tests := []struct {
		description string
		// input
		appKey string
		keys   map[string]EncryptionKey
		// output
		output EncryptionKey
	}{
		{
			description: "appKey exist in cache",
			appKey:      "sample-app-key",
			keys:        map[string]EncryptionKey{"sample-app-key": {[]byte("sample-encryption-key"), []byte("sample-iv")}},
			output:      EncryptionKey{[]byte("sample-encryption-key"), []byte("sample-iv")},
		},
		{
			description: "appKey exist in cache",
			appKey:      "demo-app-key",
			keys:        map[string]EncryptionKey{"sample-app-key": {[]byte("sample-encryption-key"), []byte("sample-iv")}},
			output:      EncryptionKey{},
		},
	}

	for i, tc := range tests {
		c := &cache{tc.keys, sync.RWMutex{}}

		output := c.Get(tc.appKey)

		if !reflect.DeepEqual(output, tc.output) {
			t.Errorf("TEST[%d] Expected output %v, got %v %s", i+1, tc.description, tc.output, output)
		}
	}
}

func Test_cache_set(t *testing.T) {
	tests := []struct {
		description string
		c           *cache
		// input
		appKey    string
		sharedKey string
		// output
		keys EncryptionKey
	}{
		{
			description: "keys do not exist in cache",
			c:           &cache{make(map[string]EncryptionKey), sync.RWMutex{}},
			appKey:      "sample-app-key",
			sharedKey:   "sample-shared-key",
			keys: EncryptionKey{[]uint8{0xf4, 0x1f, 0x45, 0xbc, 0x3e, 0xf8, 0x91, 0x15, 0x84, 0xed, 0x32, 0x14, 0xa, 0xef,
				0x6a, 0xf1, 0x24, 0x4b, 0x4d, 0xc0, 0x55, 0xb4, 0x91, 0x18, 0x2c, 0x67, 0xe5, 0x6a, 0xcc, 0x84, 0xbe, 0x46},
				[]uint8{0xf8, 0xb8, 0xef, 0x2b, 0x3, 0x43, 0x82, 0x78, 0x63, 0xb1, 0x30, 0x44, 0x7b, 0x54, 0x66, 0xd7}},
		},
		{
			description: "keys exist in cache",
			c:           &cache{map[string]EncryptionKey{"sample-app-key": {[]byte("sample-key"), []byte("sample-iv")}}, sync.RWMutex{}},
			appKey:      "sample-app-key",
			keys:        EncryptionKey{[]byte("sample-key"), []byte("sample-iv")},
		},
	}

	for i, tc := range tests {
		tc.c.Set(tc.appKey, tc.sharedKey)
		output := tc.c.Get(tc.appKey)

		if !reflect.DeepEqual(output, tc.keys) {
			t.Errorf("TEST[%d] Expected output %v, got %v %s", i+1, tc.description, tc.keys, output)
		}
	}
}

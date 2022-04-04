package service

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"developer.zopsmart.com/go/gofr/pkg"
	"developer.zopsmart.com/go/gofr/pkg/log"
)

// TestNewHTTPServiceWithNilOptions tests the values set in httpService when no options are set
func TestNewHTTPServiceWithNilOptions(t *testing.T) {
	testCase := []struct {
		resourceURL string
		expectedURL string
	}{
		{"http://example.com", "http://example.com"},
		{"http://zopsmart.com//", "http://zopsmart.com"},
	}
	for i := range testCase {
		httpService := NewHTTPServiceWithOptions(testCase[i].resourceURL, log.NewMockLogger(io.Discard), nil)

		if httpService.url != testCase[i].expectedURL {
			t.Errorf("Testcase Number: %v Expected: %v\nGot: %v", i, testCase[i].expectedURL, httpService.url)
		}
	}
}

func TestHttpServiceWithOptions_EmptyResourceAddress(t *testing.T) {
	b := new(bytes.Buffer)
	logger := log.NewMockLogger(b)

	expLog := "value for resourceAddress is empty"

	_ = NewHTTPServiceWithOptions("", logger, nil)

	if !strings.Contains(b.String(), expLog) {
		t.Errorf("TEST FAILED, Expected logs contains %v,contains %v", expLog, b.String())
	}
}

func TestNewHTTPServiceNotNilOptions(t *testing.T) {
	testCases := []struct {
		resourceURL string
		options     Options
		expectedURL string
	}{

		{"http://example.com", Options{SurgeProtectorOption: &SurgeProtectorOption{Disable: true}}, "http://example.com"},
		{"", Options{SurgeProtectorOption: &SurgeProtectorOption{Disable: true}}, ""},
		{"http://example.com", Options{SurgeProtectorOption: &SurgeProtectorOption{Disable: false}}, "http://example.com"},
		{"http://example.com//", Options{SurgeProtectorOption: &SurgeProtectorOption{Disable: false}}, "http://example.com"},
	}

	for i := range testCases {
		httpSvc := NewHTTPServiceWithOptions(testCases[i].resourceURL, log.NewMockLogger(io.Discard), &testCases[i].options)

		if httpSvc.sp.isEnabled != !testCases[i].options.SurgeProtectorOption.Disable {
			t.Errorf("expected : %v\tgot: %v", testCases[i].options.SurgeProtectorOption.Disable, httpSvc.sp.isEnabled)
		}
	}
}

// TestNewHTTPServiceAuth tests the values set when auth is set
func TestNewHTTPServiceAuth(t *testing.T) {
	testCases := []struct {
		options Options
		auth    string
	}{
		{Options{Auth: &Auth{UserName: "user", Password: "secret"}}, "Basic dXNlcjpzZWNyZXQ="},
		// both auth and oauth cannot be set
		{Options{Auth: &Auth{UserName: "user", Password: "abc", OAuthOption: &OAuthOption{}}}, ""},
		{Options{Auth: &Auth{OAuthOption: &OAuthOption{}}}, ""},
	}

	for i := range testCases {
		resourceURL := "http://example.com"
		httpSvc := NewHTTPServiceWithOptions(resourceURL, log.NewMockLogger(io.Discard), &testCases[i].options)

		if httpSvc.auth != testCases[i].auth {
			t.Errorf("expected auth: %v\tgot: %v", testCases[i].auth, httpSvc.auth)
		}

		if httpSvc.url != resourceURL {
			t.Errorf("resource url is not set\t got %v\texpected %v", httpSvc.url, resourceURL)
		}
	}
}

// TestNewHTTPService_WithHeaders tests the values set when headers are passed
func TestNewHTTPService_WithHeaders(t *testing.T) {
	testCases := []struct {
		options Options
		headers map[string]string
	}{
		{Options{Headers: nil}, nil},
		{Options{Headers: map[string]string{}}, map[string]string{}},
		{Options{Headers: map[string]string{"new header": "val"}}, map[string]string{"new header": "val"}},
	}

	for i := range testCases {
		resourceURL := "http://example.com"

		httpSvc := NewHTTPServiceWithOptions(resourceURL, log.NewMockLogger(io.Discard), &testCases[i].options)
		if !reflect.DeepEqual(httpSvc.customHeaders, testCases[i].headers) {
			t.Errorf("expected headers: %v\tgot: %v", testCases[i].headers, httpSvc.customHeaders)
		}

		if httpSvc.url != resourceURL {
			t.Errorf("resource url is not set\t got %v\texpected %v", httpSvc.url, resourceURL)
		}
	}
}

func TestNewHTTPService_WithSurgeProtection(t *testing.T) {
	testCases := []struct {
		options           Options
		surgeProtectionOp surgeProtector
	}{
		{Options{}, surgeProtector{isEnabled: true, customHeartbeatURL: "/.well-known/heartbeat", retryFrequencySeconds: 5,
			logger: log.NewLogger()}},
		{Options{SurgeProtectorOption: &SurgeProtectorOption{}}, surgeProtector{isEnabled: true, customHeartbeatURL: "/.well-known/heartbeat",
			retryFrequencySeconds: RetryFrequency, logger: log.NewLogger()}},
		{Options{SurgeProtectorOption: &SurgeProtectorOption{HeartbeatURL: "custom url"}}, surgeProtector{isEnabled: true,
			customHeartbeatURL: "custom url", retryFrequencySeconds: RetryFrequency, logger: log.NewLogger()}},
		{Options{SurgeProtectorOption: &SurgeProtectorOption{RetryFrequency: 10}}, surgeProtector{isEnabled: true,
			customHeartbeatURL: "/.well-known/heartbeat", retryFrequencySeconds: 10, logger: log.NewLogger()}},
	}

	for i := range testCases {
		resourceURL := "http://new.com"
		httpSvc := NewHTTPServiceWithOptions(resourceURL, log.NewMockLogger(io.Discard), &testCases[i].options)

		if httpSvc.sp.isEnabled != testCases[i].surgeProtectionOp.isEnabled {
			t.Errorf("expected %v\tgot %v", testCases[i].surgeProtectionOp.isEnabled, httpSvc.sp.isEnabled)
		}

		if httpSvc.sp.customHeartbeatURL != testCases[i].surgeProtectionOp.customHeartbeatURL {
			t.Errorf("expected %v\tgot %v", testCases[i].surgeProtectionOp.customHeartbeatURL, httpSvc.sp.customHeartbeatURL)
		}

		if httpSvc.sp.retryFrequencySeconds != testCases[i].surgeProtectionOp.retryFrequencySeconds {
			t.Errorf("expected %v\tgot %v", testCases[i].surgeProtectionOp.retryFrequencySeconds, httpSvc.sp.retryFrequencySeconds)
		}

		if httpSvc.url != resourceURL {
			t.Errorf("resource url is not set\t got %v\texpected %v", httpSvc.url, resourceURL)
		}
	}
}

func TestNewHTTPServiceWithOptions_WithCache(t *testing.T) {
	testCases := []struct {
		options Options
		cache   cachedHTTPService
	}{
		{Options{Cache: &Cache{}}, cachedHTTPService{}},
		{Options{Cache: &Cache{Cacher: mockCache{}}}, cachedHTTPService{cacher: mockCache{}}},
		{Options{Cache: &Cache{Cacher: mockCache{}, TTL: RetryFrequency}}, cachedHTTPService{cacher: mockCache{}, ttl: RetryFrequency}},
	}

	for i := range testCases {
		resourceURL := "http://example2.com"
		httpSvc := NewHTTPServiceWithOptions(resourceURL, log.NewMockLogger(io.Discard), &testCases[i].options)

		if httpSvc.cache.cacher != testCases[i].cache.cacher {
			t.Errorf("cacher not set")
		}

		if httpSvc.cache.ttl != testCases[i].cache.ttl {
			t.Errorf("expected cache ttl: %v\tgot %v", testCases[i].cache.ttl, httpSvc.cache.ttl)
		}

		if httpSvc.url != resourceURL {
			t.Errorf("resource url is not set\t got %v\texpected %v", httpSvc.url, resourceURL)
		}
	}
}

// nolint:gocognit // want to compare each field individually
func TestNewHTTPServiceWithOptions_MultipleFeatures(t *testing.T) {
	testCases := []struct {
		options Options
		httpSvc httpService
	}{
		{Options{Auth: &Auth{UserName: "abc", Password: "pwd"}, Cache: &Cache{Cacher: mockCache{}, TTL: 10}},
			httpService{auth: "Basic YWJjOnB3ZA==", cache: &cachedHTTPService{cacher: mockCache{}, ttl: 10},
				sp: surgeProtector{isEnabled: true, customHeartbeatURL: "/.well-known/heartbeat",
					retryFrequencySeconds: RetryFrequency, logger: log.NewLogger()}}},
		{Options{Auth: &Auth{UserName: "abc", Password: "pwd"}, Cache: &Cache{Cacher: mockCache{}, TTL: 10},
			Headers: map[string]string{"h": "hb"}}, httpService{auth: "Basic YWJjOnB3ZA==",
			cache: &cachedHTTPService{cacher: mockCache{}, ttl: 10}, customHeaders: map[string]string{"h": "hb"},
			sp: surgeProtector{isEnabled: true, customHeartbeatURL: "/.well-known/heartbeat",
				retryFrequencySeconds: RetryFrequency, logger: log.NewLogger()}}},
		{Options{Auth: &Auth{UserName: "abc", Password: "pwd"}, Cache: &Cache{Cacher: mockCache{}, TTL: 10},
			SurgeProtectorOption: &SurgeProtectorOption{RetryFrequency: RetryFrequency}},
			httpService{auth: "Basic YWJjOnB3ZA==", cache: &cachedHTTPService{cacher: mockCache{}, ttl: 10},
				sp: surgeProtector{isEnabled: true, customHeartbeatURL: "/.well-known/heartbeat",
					retryFrequencySeconds: RetryFrequency, logger: log.NewLogger()}}},
	}

	for i := range testCases {
		resourceURL := "http://example1.com"
		httpSvc := NewHTTPServiceWithOptions(resourceURL, log.NewMockLogger(io.Discard), &testCases[i].options)

		if httpSvc.cache.cacher != testCases[i].httpSvc.cache.cacher {
			t.Errorf("cacher not set")
		}

		if httpSvc.sp.isEnabled != testCases[i].httpSvc.sp.isEnabled {
			t.Errorf("expected surgeProtectionEnabled: %v\tgot %v", testCases[i].httpSvc.sp.isEnabled, httpSvc.sp.isEnabled)
		}

		if httpSvc.sp.customHeartbeatURL != testCases[i].httpSvc.sp.customHeartbeatURL {
			t.Errorf("expected heart beat URL%v\tgot %v", testCases[i].httpSvc.sp.customHeartbeatURL, httpSvc.sp.customHeartbeatURL)
		}

		if httpSvc.sp.retryFrequencySeconds != testCases[i].httpSvc.sp.retryFrequencySeconds {
			t.Errorf("expected retry frequency %v\tgot %v", testCases[i].httpSvc.sp.retryFrequencySeconds, httpSvc.sp.retryFrequencySeconds)
		}

		if httpSvc.cache.ttl != testCases[i].httpSvc.cache.ttl {
			t.Errorf("expected cache ttl: %v\tgot %v", testCases[i].httpSvc.cache.ttl, httpSvc.cache.ttl)
		}

		if httpSvc.auth != testCases[i].httpSvc.auth {
			t.Errorf("expected auth %v\tgot %v", testCases[i].httpSvc.auth, httpSvc.auth)
		}

		if !reflect.DeepEqual(httpSvc.customHeaders, testCases[i].httpSvc.customHeaders) {
			t.Errorf("expected headers: %v\tgot: %v", testCases[i].httpSvc.customHeaders, httpSvc.customHeaders)
		}

		if httpSvc.url != resourceURL {
			t.Errorf("resource url is not set\t got %v\texpected %v", httpSvc.url, resourceURL)
		}
	}
}

func TestNewHTTPServiceWithOptions_Oauth(t *testing.T) {
	clientID := "Bob"
	clientSecret := "123456"
	url := "http://dummmy"
	logger := log.NewLogger()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sampleTokenResponse := map[string]interface{}{
			"expires_in":   10,
			"access_token": "sample_token",
			"token_type":   "bearer",
		}
		_ = json.NewEncoder(w).Encode(sampleTokenResponse)
	}))

	oauthOption := OAuthOption{
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		KeyProviderURL: server.URL,
		Scope:          "some_scope",
	}

	svc := NewHTTPServiceWithOptions(url, logger, &Options{Auth: &Auth{OAuthOption: &oauthOption}})

	expectedSvc := &httpService{
		url:       url,
		auth:      "Bearer sample_token",
		logger:    logger,
		isHealthy: true,
	}

	time.Sleep(time.Duration(5) * time.Second)
	svc.mu.Lock()
	if expectedSvc.auth != svc.auth {
		t.Errorf("Expected: %v \nGot: %v", expectedSvc, svc)
	}
	svc.mu.Unlock()
}

func TestHttpServiceWithOptions_CSP(t *testing.T) {
	tcs := []struct {
		opts Options
		str  string
	}{
		{Options{Auth: &Auth{CSPOption: &CSPOption{AppKey: "mock-app-key", SharedKey: "mock-shared-key"}}}, ""},
		{Options{Auth: &Auth{CSPOption: &CSPOption{AppKey: "mock-app-key", SharedKey: ""}}}, "CSP Auth is not enabled"},
	}

	for i, tc := range tcs {
		b := new(bytes.Buffer)
		logger := log.NewMockLogger(b)

		_ = NewHTTPServiceWithOptions("http://dummy", logger, &tc.opts)

		if !strings.Contains(b.String(), tc.str) {
			t.Errorf("TESTCASE[%v] Expected logs contains %v,contains %v", i, tc.str, b.String())
		}
	}
}

func TestHttpService_HealthCheck(t *testing.T) {
	h := NewHTTPServiceWithOptions("test", log.NewLogger(), nil)

	healthCheck := h.HealthCheck()
	if healthCheck.Status != pkg.StatusUp {
		t.Errorf("Failed. Expected: UP, Got: %v", healthCheck.Status)
	}

	healthCheck = (&httpService{}).HealthCheck()
	if healthCheck.Status != pkg.StatusDown {
		t.Errorf("Failed. Expected: Down, Got: %v", healthCheck.Status)
	}
}

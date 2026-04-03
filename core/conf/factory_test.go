package conf

import (
	"context"
	"testing"
)

type stubProvider struct {
	data map[string]string
}

func (s *stubProvider) Lookup(key string) (string, bool) {
	v, ok := s.data[key]
	return v, ok
}

func (s *stubProvider) Scan(fn ScanFunc) {
	for k, v := range s.data {
		fn(k, v)
	}
}

func (s *stubProvider) Load(_ context.Context, _ []Provider) error {
	return nil
}

func newTestValueMap(data map[string]string) ValueMap {
	return &valueMap{
		providers: []Provider{&stubProvider{data: data}},
	}
}

var testData = map[string]string{
	"APP_NAME": "myapp",
	"PORT":     "8080",
	"DEBUG":    "true",
	"SECRET":   "abc123",
	"INVALID":  "notanumber",
}

// --- GetString ---

func TestGetString(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.GetString("APP_NAME"); got != "myapp" {
		t.Errorf("got %q, want %q", got, "myapp")
	}
}

func TestGetString_Missing(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.GetString("MISSING"); got != "" {
		t.Errorf("got %q, want %q", got, "")
	}
}

// --- GetInt ---

func TestGetInt(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.GetInt("PORT"); got != 8080 {
		t.Errorf("got %d, want %d", got, 8080)
	}
}

func TestGetInt_Missing(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.GetInt("MISSING"); got != 0 {
		t.Errorf("got %d, want %d", got, 0)
	}
}

func TestGetInt_InvalidValue(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.GetInt("INVALID"); got != 0 {
		t.Errorf("got %d, want %d", got, 0)
	}
}

// --- GetBool ---

func TestGetBool(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.GetBool("DEBUG"); got != true {
		t.Errorf("got %v, want %v", got, true)
	}
}

func TestGetBool_Missing(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.GetBool("MISSING"); got != false {
		t.Errorf("got %v, want %v", got, false)
	}
}

func TestGetBool_InvalidValue(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.GetBool("INVALID"); got != false {
		t.Errorf("got %v, want %v", got, false)
	}
}

// --- GetBytes ---

func TestGetBytes(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.GetBytes("SECRET"); string(got) != "abc123" {
		t.Errorf("got %q, want %q", got, "abc123")
	}
}

func TestGetBytes_Missing(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.GetBytes("MISSING"); got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

// --- GetStringOrDefault ---

func TestGetStringOrDefault(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.GetStringOrDefault("APP_NAME", "fallback"); got != "myapp" {
		t.Errorf("got %q, want %q", got, "myapp")
	}
}

func TestGetStringOrDefault_Missing(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.GetStringOrDefault("MISSING", "fallback"); got != "fallback" {
		t.Errorf("got %q, want %q", got, "fallback")
	}
}

// --- GetIntOrDefault ---

func TestGetIntOrDefault(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.GetIntOrDefault("PORT", 3000); got != 8080 {
		t.Errorf("got %d, want %d", got, 8080)
	}
}

func TestGetIntOrDefault_Missing(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.GetIntOrDefault("MISSING", 3000); got != 3000 {
		t.Errorf("got %d, want %d", got, 3000)
	}
}

// --- GetBoolOrDefault ---

func TestGetBoolOrDefault(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.GetBoolOrDefault("DEBUG", false); got != true {
		t.Errorf("got %v, want %v", got, true)
	}
}

func TestGetBoolOrDefault_Missing(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.GetBoolOrDefault("MISSING", true); got != true {
		t.Errorf("got %v, want %v", got, true)
	}
}

// --- GetBytesOrDefault ---

func TestGetBytesOrDefault(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.GetBytesOrDefault("SECRET", []byte("default")); string(got) != "abc123" {
		t.Errorf("got %q, want %q", got, "abc123")
	}
}

func TestGetBytesOrDefault_Missing(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.GetBytesOrDefault("MISSING", []byte("default")); string(got) != "default" {
		t.Errorf("got %q, want %q", got, "default")
	}
}

// --- MustGetString ---

func TestMustGetString(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.MustGetString("APP_NAME"); got != "myapp" {
		t.Errorf("got %q, want %q", got, "myapp")
	}
}

func TestMustGetString_Panics(t *testing.T) {
	vm := newTestValueMap(testData)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, got none")
		}
	}()
	vm.MustGetString("MISSING")
}

// --- MustGetInt ---

func TestMustGetInt(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.MustGetInt("PORT"); got != 8080 {
		t.Errorf("got %d, want %d", got, 8080)
	}
}

func TestMustGetInt_Panics(t *testing.T) {
	vm := newTestValueMap(testData)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, got none")
		}
	}()
	vm.MustGetInt("MISSING")
}

// --- MustGetBool ---

func TestMustGetBool(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.MustGetBool("DEBUG"); got != true {
		t.Errorf("got %v, want %v", got, true)
	}
}

func TestMustGetBool_Panics(t *testing.T) {
	vm := newTestValueMap(testData)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, got none")
		}
	}()
	vm.MustGetBool("MISSING")
}

// --- MustGetBytes ---

func TestMustGetBytes(t *testing.T) {
	vm := newTestValueMap(testData)

	if got := vm.MustGetBytes("SECRET"); string(got) != "abc123" {
		t.Errorf("got %q, want %q", got, "abc123")
	}
}

func TestMustGetBytes_Panics(t *testing.T) {
	vm := newTestValueMap(testData)

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, got none")
		}
	}()
	vm.MustGetBytes("MISSING")
}

// --- Multiple providers (priority) ---

func TestMultipleProviders_FirstWins(t *testing.T) {
	primary := &stubProvider{data: map[string]string{"KEY": "primary"}}
	fallback := &stubProvider{data: map[string]string{"KEY": "fallback", "ONLY_FALLBACK": "yes"}}

	vm := &valueMap{providers: []Provider{primary, fallback}}

	if got := vm.GetString("KEY"); got != "primary" {
		t.Errorf("got %q, want %q", got, "primary")
	}
	if got := vm.GetString("ONLY_FALLBACK"); got != "yes" {
		t.Errorf("got %q, want %q", got, "yes")
	}
}

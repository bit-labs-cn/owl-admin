package contract

import "testing"

func TestGetNilBeforeRegister(t *testing.T) {
	Register(nil)
	if Get() != nil {
		t.Fatal("清空后应返回 nil")
	}
}

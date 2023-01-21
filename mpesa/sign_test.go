package mpesa

import (
	"testing"
	"time"
)

func TestSign(t *testing.T) {
	ti := time.Now().String()
	s := "fgg"
	passkey := "fgg"
	f, err := GeneratePassword(s, passkey, ti)
	if err != nil {
		t.Error(err)
	}
	t.Log(f)
}
func TestSignNillValues(t *testing.T) {
	ti := time.Now().String()
	s := ""
	passkey := ""
	f, err := GeneratePassword(s, passkey, ti)
	if err == nil {
		t.Error(err)
	}
	t.Log(f)
}

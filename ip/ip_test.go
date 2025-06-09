package ip

import (
	"context"
	"testing"
)

func TestAnonymize(t *testing.T) {

	data := map[string]string{
		"127.0.0.1":  "127.0.0.1",
		"86.76.17.1": "86.76.0.0",
		"::1":        "::1",
		"[::1]":      "::1",
	}

	for ip, expected := range data {
		ano := Anonymize(context.Background(), ip)
		if ano != expected {
			t.Errorf("Anonymize(%s)=%s expected %s", ip, ano, expected)
		}
	}
}

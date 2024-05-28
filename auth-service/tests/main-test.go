package tests

import (
	"testing"
)

func getMessage() string {
	return "Hello"
}

func DummyTest(t *testing.T) {
	if msg := getMessage(); msg != "Hello" {
		t.Fatalf(`DummyTest("") = %q`, msg)
	}
}

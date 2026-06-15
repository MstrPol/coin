package main

import "testing"

func TestSmoke(t *testing.T) {
	if serviceName == "" {
		t.Fatal("serviceName must be set")
	}
}

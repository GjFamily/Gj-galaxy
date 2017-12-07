package Gj_galaxy

import (
	"testing"
)

func TestPreRun(t *testing.T) {

	config.LoadConfig("config.json")
	inited = false
	err := PreRun()
	if err != nil {
		t.Fatal(err)
	}
}

func TestRun(t *testing.T) {
	err := Run()
	if err != nil {
		t.Fatal(err)
	}
}

func TestExit(t *testing.T) {
	err := Exit()
	if err != nil {
		t.Fatal(err)
	}
}

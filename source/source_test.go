package source

import (
	"testing"
)

func TestNonExistent(t *testing.T) {
	t.Parallel()

	src := Source{}

	_, err := src.ExecuteString("doesnotexist", nil)

	if err == nil {
		t.Error("Expected to get an error when creating a source based off of a non-existent local profile")
	}
}

func TestBase(t *testing.T) {
	t.Parallel()

	src := Source{}

	bytes, err := src.ExecuteString("base", nil)

	if err != nil {
		t.Fatal(err)
	}

	if string(bytes) != "ENVIRONATOR_TEST=base\n" {
		t.Fatalf("Unexpected response: %s", string(bytes))
	}
}

func TestInheritance(t *testing.T) {
	t.Parallel()

	src := Source{}

	bytes, err := src.ExecuteString("test", nil)

	if err != nil {
		t.Fatal(err)
	}

	if string(bytes) != "ENVIRONATOR_TEST=base\n\nENVIRONATOR_TEST=bar:$ENVIRONATOR_TEST\n" {
		t.Fatalf("Unexpected response: %s", string(bytes))
	}
}

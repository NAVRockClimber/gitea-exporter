package gitea

import (
	"testing"
)

func TestGetURLSimple(t *testing.T) {
	client := New("https://git.example.com", "secret-token")
	got := client.getURL("orgs")

	want := "https://git.example.com/api/v1/orgs"
	if got != want {
		t.Fatalf("URL not correct. Got: %s, want %s", got, want)
	}
}

func TestGetURLWithReplace(t *testing.T) {
	client := New("https://git.example.com", "secret-token")
	got := client.getURL("repos", "some-org")

	want := "https://git.example.com/api/v1/orgs/some-org/repos"
	if got != want {
		t.Fatalf("URL not correct. Got: %s, want %s", got, want)
	}
}

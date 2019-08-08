package scm_clients

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

var noContext = context.Background()

func BaseTest_GetFileContents(t *testing.T, client ScmClient) {
	actualContent, err := client.GetFileContents(noContext, "afolder/.drone.yml", "8ecad91991d5da985a2a8dd97cc19029dc1c2899")
	actualContent = strings.Replace(actualContent, "\r", "", -1)
	if err != nil {
		t.Error(err)
		return
	}

	if want, got := "kind: pipeline\nname: default\n\nsteps:\n- name: build\n  image: golang\n  commands:\n  - go build\n  - go test -short\n\n- name: integration\n  image: golang\n  commands:\n  - go test -v\n", actualContent; want != got {
		t.Errorf("Test failed:\n  want %q\n   got %q", want, got)
	}
}

func BaseTest_ChangedFilesInDiff(t *testing.T, client ScmClient) {
	actualFiles, err := client.ChangedFilesInDiff(noContext, "2897b31ec3a1b59279a08a8ad54dc360686327f7", "8ecad91991d5da985a2a8dd97cc19029dc1c2899")
	if err != nil {
		t.Error(err)
		return
	}

	expectedFiles := []string{
		"a/b/c/d/file",
	}

	if want, got := expectedFiles, actualFiles; !reflect.DeepEqual(want, got) {
		t.Errorf("Test failed:\n  want %q\n   got %q", want, got)
	}
}

func BaseTest_ChangedFilesInPullRequest(t *testing.T, client ScmClient) {
	actualFiles, err := client.ChangedFilesInPullRequest(noContext, 3)
	if err != nil {
		t.Error(err)
		return
	}

	expectedFiles := []string{
		"e/f/g/h/.drone.yml",
	}

	if want, got := expectedFiles, actualFiles; !reflect.DeepEqual(want, got) {
		t.Errorf("Test failed:\n  want %q\n   got %q", want, got)
	}
}

func BaseTest_GetFileListing(t *testing.T, client ScmClient) {
	actualFiles, err := client.GetFileListing(noContext, "afolder", "8ecad91991d5da985a2a8dd97cc19029dc1c2899")
	if err != nil {
		t.Error(err)
		return
	}

	expectedFiles := []FileListingEntry{
		{Type: "file", Path: "afolder/.drone.yml", Name: ".drone.yml"},
		{Type: "dir", Path: "afolder/abfolder", Name: "abfolder"},
	}

	if want, got := expectedFiles, actualFiles; !reflect.DeepEqual(want, got) {
		t.Errorf("Test failed:\n  want %q\n   got %q", want, got)
	}
}

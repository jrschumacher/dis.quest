package pds

import (
	"testing"
)

func TestMockService(t *testing.T) {
	mock := NewMockService()
	post, err := mock.CreatePost("hello world")
	if err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}
	if post.Content != "hello world" {
		t.Errorf("expected content 'hello world', got '%s'", post.Content)
	}
	fetched, err := mock.GetPost(post.ID)
	if err != nil {
		t.Fatalf("GetPost failed: %v", err)
	}
	if fetched == nil || fetched.Content != "hello world" {
		t.Errorf("expected to fetch post with content 'hello world', got %+v", fetched)
	}

	// set and verify selected answer
	if err := mock.SetSelectedAnswer(post.ID, "reply-1"); err != nil {
		t.Fatalf("SetSelectedAnswer failed: %v", err)
	}
	fetched, err = mock.GetPost(post.ID)
	if err != nil {
		t.Fatalf("GetPost failed: %v", err)
	}
	if fetched.SelectedAnswer != "reply-1" {
		t.Errorf("expected selected answer 'reply-1', got '%s'", fetched.SelectedAnswer)
	}
}

// Package pds provides interfaces and mocks for Personal Data Server (PDS) interactions.
package pds

import "fmt"

// Post represents a minimal post structure for testing.
type Post struct {
	ID      string
	Content string
}

// Service defines the interface for PDS operations.
type Service interface {
	CreatePost(content string) (*Post, error)
	GetPost(id string) (*Post, error)
}

// MockService is an in-memory mock implementation of Service.
type MockService struct {
	posts map[string]*Post
}

// NewMockService creates a new MockService.
func NewMockService() *MockService {
	return &MockService{posts: make(map[string]*Post)}
}

func (m *MockService) CreatePost(content string) (*Post, error) {
	id := fmt.Sprintf("mock-%d", len(m.posts)+1)
	post := &Post{ID: id, Content: content}
	m.posts[id] = post
	return post, nil
}

func (m *MockService) GetPost(id string) (*Post, error) {
	post, ok := m.posts[id]
	if !ok {
		return nil, nil
	}
	return post, nil
}

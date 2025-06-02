package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type CreateSessionRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type CreateSessionResponse struct {
	AccessJwt  string `json:"accessJwt"`
	RefreshJwt string `json:"refreshJwt"`
	Did        string `json:"did"`
	Handle     string `json:"handle"`
}

// CreateSession calls the ATProto createSession endpoint with handle and app password
func CreateSession(pds, handle, password string) (*CreateSessionResponse, error) {
	url := fmt.Sprintf("%s/xrpc/com.atproto.server.createSession", pds)
	body, _ := json.Marshal(CreateSessionRequest{
		Identifier: handle,
		Password:   password,
	})
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New("invalid credentials or failed to create session")
	}
	var out CreateSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

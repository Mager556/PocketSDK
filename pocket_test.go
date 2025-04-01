package pocket

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (s roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return s(r)
}

func newClient(t *testing.T, statusCode int, path string, body string) *Client {
	return &Client{
		client: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				assert.Equal(t, path, r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)

				return &http.Response{
					StatusCode: statusCode,
					Body:       io.NopCloser(strings.NewReader(body)),
				}, nil
			}),
		},
		consumerKey: "key",
	}
}

func TestClient_GetAccessToken(t *testing.T) {
	tests := []struct {
		name         string
		requestToken string
		response     string
		statusCode   int
		want         string
		wantErr      bool
	}{
		{
			name:         "Default-OK",
			requestToken: "12345-qwerty",
			response:     "access_token=qwe-rty-123",
			statusCode:   200,
			want:         "qwe-rty-123",
			wantErr:      false,
		},
		{
			name:    "Empty requestToken",
			wantErr: true,
		},
		{
			name:         "Empty response code",
			requestToken: "12345-qwerty",
			response:     "access_token=",
			statusCode:   200,
			wantErr:      true,
		},
		{
			name:         "Not-2XX Response",
			requestToken: "12345-qwerty",
			statusCode:   400,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newClient(t, tt.statusCode, "/v3/oauth/authorize", tt.response)

			got, err := client.GetAccessToken(context.Background(), tt.requestToken)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestClient_GetRequestToken(t *testing.T) {
	tests := []struct {
		name        string
		redirectUrl string
		statusCode  int
		response    string
		want        string
		wantErr     bool
	}{
		{
			name:        "Default-OK",
			statusCode:  200,
			redirectUrl: "https://localhost",
			response:    "code=qwe-rty-123",
			want:        "qwe-rty-123",
			wantErr:     false,
		},
		{
			name:    "Empty redirect URL",
			wantErr: true,
		},
		{
			name:        "Empty response code",
			redirectUrl: "https://localhost",
			statusCode:  200,
			response:    "code=",
			wantErr:     true,
		},
		{
			name:        "Non-2XX Response",
			redirectUrl: "https://localhost",
			statusCode:  400,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newClient(t, tt.statusCode, "/v3/oauth/request", tt.response)

			got, err := client.GetRequestToken(context.Background(), tt.redirectUrl)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}

}

func TestClient_Add(t *testing.T) {
	tests := []struct {
		name       string
		input      AddInput
		statusCode int
		wantErr    bool
	}{
		{
			name: "Default-OK",
			input: AddInput{
				URL:         "some_url.com",
				Title:       "some_title",
				Tags:        []string{"some_tag_1", "some_tag_2"},
				AccessToken: "access-to-ken",
			},
			statusCode: 200,
			wantErr:    false,
		},
		{
			name: "Default-EmptyTags-OK",
			input: AddInput{
				URL:         "some_url.com",
				Title:       "some_title",
				AccessToken: "access-to-ken",
			},
			statusCode: 200,
			wantErr:    false,
		},
		{
			name: "Empty accessToken",
			input: AddInput{
				URL:         "some_url.com",
				Title:       "some_title",
				Tags:        []string{"some_tag_1", "some_tag_2"},
				AccessToken: "",
			},
			wantErr: true,
		},
		{
			name:    "Empty input",
			wantErr: true,
		},
		{
			name: "Empty URL",
			input: AddInput{
				URL:         "",
				Title:       "some_title",
				Tags:        []string{"some_tag_1", "some_tag_2"},
				AccessToken: "access-to-ken",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newClient(t, tt.statusCode, "/v3/add", "")

			err := client.Add(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

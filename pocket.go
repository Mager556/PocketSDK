package pocket

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	host = "https://getpocket.com/v3"

	authorizeURL = "https://getpocket.com/auth/authorize?request_token=%s&redirect_uri=%s"

	endpointRequestToken = "/oauth/request"
	endpointAuthorize    = "/oauth/authorize"
	endpointAdd          = "/add"

	xErrorHeader = "X-Error"

	defaultTimeout = 5 * time.Second
)

type (
	requestTokenRequest struct {
		ConsumerKey string `json:"consumer_key"`
		RedirectURI string `json:"redirectUri"`
	}

	accessTokenRequest struct {
		ConsumerKey string `json:"consumer_key"`
		Code        string `json:"code"`
	}

	AuthorizeResponse struct {
		AccessToken string `json:"access_token"`
		Username    string `json:"username"`
	}

	addRequest struct {
		Url         string `json:"url"`
		Title       string `json:"title"`
		Tags        string `json:"tags"`
		TweetId     int    `json:"tweet_id"`
		ConsumerKey string `json:"consumer_key"`
		AccessToken string `json:"access_token"`
	}

	AddInput struct {
		URL         string
		Title       string
		Tags        []string
		AccessToken string
	}
)

func (i AddInput) validate() error {
	if i.URL == "" {
		return errors.New("required URL values is empty")
	}

	if i.AccessToken == "" {
		return errors.New("access token is empty")
	}

	return nil
}

func (i AddInput) generateRequest(consumerKey string) addRequest {
	return addRequest{
		Url:         i.URL,
		Tags:        strings.Join(i.Tags, ","),
		Title:       i.Title,
		AccessToken: i.AccessToken,
		ConsumerKey: consumerKey,
	}
}

type Client struct {
	client      *http.Client
	consumerKey string
}

func NewClient(consumerKey string) (*Client, error) {
	if consumerKey == "" {
		return nil, errors.New("Consumer key is empty")
	}

	return &Client{
		client: &http.Client{
			Timeout: defaultTimeout,
		},
		consumerKey: consumerKey,
	}, nil
}

func (c *Client) GetRequestToken(ctx context.Context, redirectUri string) (string, error) {
	if redirectUri == "" {
		return "", errors.New("ReditectUri is empty")
	}

	inp := requestTokenRequest{
		ConsumerKey: c.consumerKey,
		RedirectURI: redirectUri,
	}

	values, err := c.doHTTP(ctx, endpointRequestToken, inp)
	if err != nil {
		return "", err
	}

	requestToken := values.Get("code")
	if requestToken == "" {
		return "", errors.New("Empty request token in API response")
	}

	return requestToken, nil
}

func (c *Client) GetAuthorizationURL(ctx context.Context, requestToken, redirectUrl string) (string, error) {
	if requestToken == "" {
		return "", errors.New("RequestToken is empty")
	}
	if redirectUrl == "" {
		return "", errors.New("RedirectUrl is empty")
	}

	return fmt.Sprintf(authorizeURL, requestToken, redirectUrl), nil
}

func (c *Client) Add(ctx context.Context, input AddInput) error {
	if err := input.validate(); err != nil {
		return err
	}

	inp := input.generateRequest(c.consumerKey)

	_, err := c.doHTTP(ctx, endpointAdd, inp)

	return err
}

func (c *Client) GetAccessToken(ctx context.Context, requestToken string) (string, error) {
	if requestToken == "" {
		return "", errors.New("RequestToken is empty")
	}

	inp := accessTokenRequest{
		ConsumerKey: c.consumerKey,
		Code:        requestToken,
	}

	values, err := c.doHTTP(ctx, endpointAuthorize, inp)
	if err != nil {
		return "", err
	}

	accessToken := values.Get("code")
	if accessToken == "" {
		return "", errors.New("Empty access token in API response")
	}

	return accessToken, nil
}

func (c *Client) doHTTP(ctx context.Context, endpoint string, body interface{}) (url.Values, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return url.Values{}, errors.Join(err, errors.New("Failed to marshal body"))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, host+endpoint, bytes.NewBuffer(b))
	if err != nil {
		return url.Values{}, errors.Join(err, errors.New("Failed to create request"))
	}

	req.Header.Add("Content-Type", "application/json; charset=UTF8")

	resp, err := c.client.Do(req)
	if err != nil {
		return url.Values{}, errors.Join(err, errors.New("Failed to send http request..."))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Sprintf("API Error : %v", resp.Header.Get(xErrorHeader))
		return url.Values{}, errors.New(err)
	}

	respB, err := io.ReadAll(resp.Body)
	if err != nil {
		return url.Values{}, errors.Join(err, errors.New("Failed read response"))
	}

	values, err := url.ParseQuery(string(respB))
	if err != nil {
		return url.Values{}, errors.Join(err, errors.New("Failed to parse response values"))
	}

	return values, nil
}

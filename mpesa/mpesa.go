package mpesa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type ClientOption func(*Mpesa)

type Mpesa struct {
	// if set to true the access token will be reused.
	// until it expires.
	// otherwise each request will always get a new token.
	// which will slow down your requests.
	// the default is true.
	CacheAccessToken bool
	// For sandbox use false and for production use true.
	Live           bool
	ConsumerKey    string
	ConsumerSecret string
	TimeOut        time.Duration
	// for those using only one payBill.
	// you can set a default passkey to be used.
	// instead of passing a passkey evey time.
	// when doing an stk push.
	// use SetDefaultPassKey to change default.
	// pass key at runtime..
	DefaultPassKey string
	// You can pass the Mpesa shortcode.
	// you want In case you dont to pass the shortcode each time.
	// you are sending a request.
	// this is ideal for those using a single shortcode.
	DefaultC2BShortCode string
	DefaultB2CShortCode string
	// for b2c.
	DefaultInitiatorName string
	// for b2c.
	DefaultSecurityCredential string
	cache                     *Cache
}

func New(consumerKey, consumerSecret string, opts ...ClientOption) *Mpesa {
	client := &Mpesa{
		Live:             false,
		ConsumerKey:      consumerKey,
		ConsumerSecret:   consumerSecret,
		TimeOut:          20 * time.Second,
		CacheAccessToken: true,
		cache:            NewCache(),
	}
	// Here, It applies all user-defined options
	// to the target struct.
	for _, opt := range opts {
		opt(client)
	}

	return client
}

// WithCache wether to cache the access token.
func WithCache(cache bool) ClientOption {
	return func(m *Mpesa) {
		m.CacheAccessToken = cache
	}
}

// WithB2CShortCode will set the default shortcode
// to use if you do not provide any.
func WithB2CShortCode(shortCode string) ClientOption {
	return func(m *Mpesa) {
		m.DefaultC2BShortCode = shortCode
	}
}

// WithPassKey You can set the default pass key
// Over here so that you dont have to pass it each time
// You are sending an StkRequest.
func WithPassKey(passKey string) ClientOption {
	return func(m *Mpesa) {
		m.DefaultPassKey = passKey
	}
}

// WithTimeout this will set the connection timeout to Mpesa
// the default is 20 seconds when sending an http request.
func WithTimeout(timeOut time.Duration) ClientOption {
	return func(m *Mpesa) {
		m.TimeOut = timeOut
	}
}

// WithLiveMode changes from production to sandbox and viceversa
// at runtime.
func WithLiveMode(mode bool) ClientOption {
	return func(m *Mpesa) {
		m.Live = mode
	}
}

// B2CRequest Sends Money from a business to the Customer.
func (m *Mpesa) B2CRequest(b2c B2CRequestBody) (*MpesaResult, error) {
	err := b2c.Validate()
	if err != nil {
		return nil, err
	}
	var mpesaResult MpesaResult
	err = m.sendAndProcessStkPushRequest(context.Background(), m.getMpesaURL(string(b2cURL)), b2c, &mpesaResult)
	return &mpesaResult, err
}

// StkPushRequest send an Mpesa express request.
func (m *Mpesa) StkPushRequest(ctx context.Context, body StKPushRequestBody) (*StkPushResult, error) {
	err := body.Validate()
	if err != nil {
		return nil, err
	}
	t := time.Now().Format("20060102150405")
	pass, err := GeneratePassword(body.BusinessShortCode, m.DefaultPassKey, t)
	if err != nil {
		return nil, err
	}
	requestBody := StkRequestFullBody{
		StKPushRequestBody: body,
		Password:           pass,
		Timestamp:          t,
		TransactionType:    CustomerPayBillOnline,
		PartyA:             body.PhoneNumber,
		PartyB:             body.BusinessShortCode,
	}
	var stkPushResult StkPushResult
	err = m.sendAndProcessStkPushRequest(ctx, m.getMpesaURL(string(stkPush)), requestBody, &stkPushResult)
	return &stkPushResult, err
}

// StkPushVerification use this to confirm your stk push if it was a failure or success
// CheckoutRequestID is the CheckoutRequestID you got when you sent the StkPushRequest request
// you dont have to send a passkey if you have a DefaultPassKey set.
func (m *Mpesa) StkPushVerification(checkoutRequestID string, businessShortCode string) (*StkPushQueryResponseBody, error) {
	t := time.Now().Format("20060102150405")
	pass, err := GeneratePassword(businessShortCode, m.DefaultPassKey, t)
	if err != nil {
		return nil, err
	}
	body := StkPushQueryRequestBody{
		BusinessShortCode: businessShortCode,
		Password:          pass,
		Timestamp:         t,
		CheckoutRequestID: checkoutRequestID,
	}
	var stkPushResult StkPushQueryResponseBody
	err = m.sendAndProcessStkPushRequest(context.Background(), m.getMpesaURL(string(stkPushQuery)), body, &stkPushResult)
	return &stkPushResult, err
}

func (m *Mpesa) StkPushQuery(ctx context.Context, body StkPushQueryRequestBody) (*StkPushQueryResponseBody, error) {
	var stkPushResult StkPushQueryResponseBody
	err := m.sendAndProcessStkPushRequest(ctx, m.getMpesaURL(string(stkPushQuery)), body, &stkPushResult)
	return &stkPushResult, err
}

func (m *Mpesa) sendAndProcessStkPushRequest(ctx context.Context, url string, data, respItem any) error {
	token, err := m.GetAccessToken(ctx)
	if err != nil {
		return err
	}
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Authorization"] = "Bearer " + token.AccessToken

	resp, err := postRequest(ctx, url, data, headers, m.TimeOut)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		errx := Body.Close()
		if errx != nil {
			return
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("url: %s  status code: %d  body: %s", resp.Request.URL, resp.StatusCode, b)
	}
	if errx := json.NewDecoder(resp.Body).Decode(respItem); errx != nil {
		return errors.New("error converting from json")
	}
	return nil
}

// GetAccessToken will get the token to be used to query data.
func (m *Mpesa) GetAccessToken(ctx context.Context) (*AccessTokenResponse, error) {
	// if we have allowed caching.
	if m.CacheAccessToken {
		if m2, ok := m.cache.Get("token"); ok {
			return m2, nil
		}
		goto token
	}

token:
	ctx, cancelFunc := context.WithTimeout(ctx, m.TimeOut)
	defer cancelFunc()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.getMpesaURL(string(tokenURL)), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(m.ConsumerKey, m.ConsumerSecret)
	req.Header.Set("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	client := &http.Client{
		Timeout: m.TimeOut,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		errx := Body.Close()
		if errx != nil {
			return
		}
	}(resp.Body)

	// Check the reponse code and return early.
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, errors.WithMessagef(nil, "url: %s  status code: %d  body: %s", resp.Request.URL, resp.StatusCode, b)
	}

	var token AccessTokenResponse
	if errx := json.NewDecoder(resp.Body).Decode(&token); errx != nil {
		return nil, errors.Wrap(errx, "error converting from json")
	}
	// cache the token.
	m.cache.Set(&AccessTokenResponse{AccessToken: token.AccessToken, ExpireTime: time.Now().Add(time.Minute * 50)})
	return &token, nil
}

func postRequest(ctx context.Context, url string, data any, headers map[string]string, timeOut time.Duration) (*http.Response, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	ctx, cancelFunc := context.WithTimeout(ctx, timeOut)
	defer cancelFunc()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	client := &http.Client{
		Timeout: timeOut,
	}
	return client.Do(req)
}

type RequestError struct {
	StatusCode int
	Message    string
	URL        string
}

func (r *RequestError) Error() error {
	return fmt.Errorf("url: %s  code: %d  body  : %s", r.URL, r.StatusCode, r.Message)
}

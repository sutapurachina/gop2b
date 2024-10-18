package gop2b

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// HeaderXTxcAPIKey is the HTTP Header for account API key
	HeaderXTxcAPIKey = "X-TXC-APIKEY"
	// HeaderXTxcPayloard is the HTTP Header for body json encoded in base64
	HeaderXTxcPayload = "X-TXC-PAYLOAD"
	// HeaderXTxcSignature is the HTTP Header for encrypted in HmacSHA512 payload
	HeaderXTxcSignature = "X-TXC-SIGNATURE"
)

type auth struct {
	APIKey    string
	APISecret string
}

type client struct {
	http  *http.Client
	auth  *auth
	url   string
	wsUrl string
}

type response struct {
	Header     http.Header
	Body       io.ReadCloser
	StatusCode int
	Status     string
}

type wsRequest struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	Id     int64    `json:"id"`
}

func newWsRequest(method string, params ...string) *wsRequest {
	req := &wsRequest{
		Method: method,
		Params: []string{},
	}
	for _, p := range params {
		req.Params = append(req.Params, p)
	}
	req.Id = time.Now().Unix()
	return req
}

func newPingRequest() *wsRequest {
	return newWsRequest("server.ping")
}

func newUnsubscribeRequest(endpoint string) *wsRequest {
	return newWsRequest(endpoint + ".unsubscribe")
}

func checkHTTPStatus(resp response, expected ...int) error {
	for _, e := range expected {
		if resp.StatusCode == e {
			return nil
		}
	}
	return fmt.Errorf("http response status != %+v, got %d", expected, resp.StatusCode)
}

func mergeHeaders(firstHeaders map[string]string, secondHeaders map[string]string) map[string]string {
	if secondHeaders == nil {
		return firstHeaders
	}
	if firstHeaders == nil {
		return secondHeaders
	}
	for k, v := range secondHeaders {
		if firstHeaders[k] == "" {
			firstHeaders[k] = v
		}
	}
	return firstHeaders
}

func (c *client) sendPost(url string, additionalHeaders map[string]string, body io.Reader) (*response, error) {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return &response{}, fmt.Errorf("error creating POST request, %v", err)
	}

	if additionalHeaders == nil {
		additionalHeaders = make(map[string]string)
	}
	additionalHeaders[HeaderXTxcPayload] = base64.StdEncoding.EncodeToString(bodyBytes)

	if c.auth != nil {
		h := hmac.New(sha512.New, []byte(c.auth.APISecret))
		h.Write([]byte(additionalHeaders[HeaderXTxcPayload]))
		signature := hex.EncodeToString(h.Sum(nil))
		additionalHeaders[HeaderXTxcSignature] = signature
	}

	return c.sendRequest(req, additionalHeaders)
}

func (c *client) sendGet(url string, additionalHeaders map[string]string) (*response, error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return &response{}, fmt.Errorf("error creating GET request, %v", err)
	}

	return c.sendRequest(req, additionalHeaders)
}

func (c *client) sendRequest(request *http.Request, additionalHeaders map[string]string) (*response, error) {

	for k, v := range additionalHeaders {
		request.Header.Add(k, v)
	}

	thisHeaders := map[string]string{}
	thisHeaders["Content-type"] = "application/json"
	if c.auth != nil {
		thisHeaders[HeaderXTxcAPIKey] = c.auth.APIKey
	}
	headers := mergeHeaders(additionalHeaders, thisHeaders)
	for k, v := range headers {
		request.Header.Add(k, v)
	}
	resp, err := c.http.Do(request)
	if err != nil {
		fmt.Println(fmt.Sprintf("erro: %v", err))
		return nil, err
	}
	return &response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Header:     resp.Header,
		Body:       resp.Body,
	}, nil
}

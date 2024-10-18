package gop2b

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"io"
	"net/http"
	"strconv"
	"time"
)

type AccountBalancesResp struct {
	Response
	Result map[string]AccountBalance `json:"result"`
}

type AccountBalance struct {
	Available decimal.Decimal `json:"available,string"`
	Freeze    decimal.Decimal `json:"freeze,string"`
}

type AccountBalancesRequest struct {
	Request
}

type AccountCurrencyBalanceResp struct {
	Response
	Result AccountBalance `json:"result"`
}

type AccountCurrencyBalanceRequest struct {
	Request
	Currency string `json:"currency"`
}

func (c *client) PostBalances(request *AccountBalancesRequest) (*AccountBalancesResp, error) {
	url := fmt.Sprintf("%s/account/balances", c.url)
	asJSON, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendPost(url, nil, bytes.NewReader(asJSON))
	if err != nil {
		return nil, err
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = checkHTTPStatus(*resp, http.StatusOK)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s\n", err.Error(), string(bodyBytes)))
	}

	var result AccountBalancesResp
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) PostCurrencyBalance(request *AccountCurrencyBalanceRequest) (*AccountCurrencyBalanceResp, error) {
	url := fmt.Sprintf("%s/account/balance", c.url)
	request.Request.Nonce = strconv.FormatInt(time.Now().UnixMilli(), 10)
	request.Request.Request = "/api/v2/account/balance"
	asJSON, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	resp, err := c.sendPost(url, nil, bytes.NewReader(asJSON))
	if err != nil {
		return nil, err
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = checkHTTPStatus(*resp, http.StatusOK)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s\n", err.Error(), string(bodyBytes)))
	}
	var result AccountCurrencyBalanceResp
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

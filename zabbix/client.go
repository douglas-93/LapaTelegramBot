package zabbix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Client struct {
	URL   string
	Token string
}

type request struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
	Auth    string      `json:"auth,omitempty"`
}

type response struct {
	Result json.RawMessage `json:"result"`
	Error  *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func NewClient() *Client {
	url := os.Getenv("ZABBIX_API_URL")
	token := os.Getenv("ZABBIX_API_TOKEN")

	return &Client{
		URL:   url,
		Token: token,
	}
}

func (c *Client) Call(method string, params interface{}) (json.RawMessage, error) {
	if strings.TrimSpace(c.URL) == "" {
		return nil, fmt.Errorf("ZABBIX_API_URL não foi definida")
	}
	if strings.TrimSpace(c.Token) == "" {
		return nil, fmt.Errorf("ZABBIX_API_TOKEN não foi definido")
	}

	req := request{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", c.URL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json-rpc")
	httpReq.Header.Set("Authorization", "Bearer "+strings.TrimSpace(c.Token))

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("Zabbix HTTP %s", resp.Status)
	}

	var r response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	if r.Error != nil {
		if r.Error.Data != "" {
			return nil, fmt.Errorf("Zabbix API (%d): %s: %s", r.Error.Code, r.Error.Message, r.Error.Data)
		}
		return nil, fmt.Errorf("Zabbix API (%d): %s", r.Error.Code, r.Error.Message)
	}

	return r.Result, nil
}

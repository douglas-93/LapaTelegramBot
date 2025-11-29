package zabbix

import "encoding/json"

type HostResponse struct {
	HostID     string `json:"hostid"`
	Host       string `json:"host"`
	Interfaces []struct {
		InterfaceID string `json:"interfaceid"`
		IP          string `json:"ip"`
	} `json:"interfaces"`
}

func (c *Client) ListIps() ([]HostResponse, error) {
	params := map[string]interface{}{
		"output":           []string{"hostid", "host"},
		"filter":           map[string]string{"status": "0"},
		"selectInterfaces": []string{"interfaceid", "ip"},
	}

	resp, err := c.Call("host.get", params)
	if err != nil {
		return nil, err
	}

	var hosts []HostResponse
	err = json.Unmarshal(resp, &hosts)
	if err != nil {
		return nil, err
	}
	return hosts, nil
}

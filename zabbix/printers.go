package zabbix

import "encoding/json"

func (c *Client) GetPrinters() ([]Host, error) {
	params := map[string]interface{}{
		"output":   "extend",
		"groupids": "22", /* ID do grupo de Impressoras */
		"filter": map[string]string{
			"status": "0",
		},
	}

	resp, err := c.Call("host.get", params)
	if err != nil {
		return nil, err
	}

	var hosts []Host
	json.Unmarshal(resp, &hosts)
	return hosts, nil
}

package monitor

import (
	"TelegramNotify/zabbix"
	"encoding/json"
	"strconv"
	"sync"
)

type Printer struct {
	HostData     zabbix.Host
	BlackCounter int64
	ColorCounter int64
	TotalCounter int64
}

func GetPrintersCounter(z *zabbix.Client) ([]Printer, error) {
	hosts, err := z.GetPrinters()
	if err != nil {
		return nil, err
	}

	printers := make([]Printer, len(hosts))
	for i, host := range hosts {
		printers[i] = Printer{HostData: host, BlackCounter: 0, ColorCounter: 0, TotalCounter: 0}
	}

	var wg sync.WaitGroup

	for i := range printers {
		wg.Add(1)
		go getCounterItemValue(z, &printers[i], &wg)
	}

	func() {
		wg.Wait()
	}()

	return printers, nil
}

func getCounterItemValue(z *zabbix.Client, printer *Printer, wg *sync.WaitGroup) {
	defer wg.Done()

	params := map[string]interface{}{
		"output":  "extend",
		"hostids": printer.HostData.Hostid,
		"search": map[string]string{
			"key_": "contador",
		},
	}

	resp, err := z.Call("item.get", params)
	if err != nil {
		printer.HostData.Error = true
		return
	}

	var items []struct {
		Itemid    string `json:"itemid"`
		Key       string `json:"key_"`
		Lastvalue string `json:"lastvalue"`
	}

	json.Unmarshal(resp, &items)

	if len(items) > 0 {
		for _, item := range items {
			switch item.Key {
			case "contador.colorido":
				c, _ := strconv.Atoi(item.Lastvalue)
				printer.ColorCounter = int64(c)
			case "contador.peb":
				c, _ := strconv.Atoi(item.Lastvalue)
				printer.BlackCounter = int64(c)
			case "contador.total":
				c, _ := strconv.Atoi(item.Lastvalue)
				printer.TotalCounter = int64(c)
			}
		}
	}
}

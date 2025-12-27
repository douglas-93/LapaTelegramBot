package file_handler

import (
	"LapaTelegramBot/monitor"
	"LapaTelegramBot/zabbix"
	"fmt"

	"github.com/xuri/excelize/v2"
)

func GenerateSheet() string {
	fileName := "contadores.xlsx"
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Cabeçalho
	f.SetCellValue("Sheet1", "A1", "REP")
	f.SetCellValue("Sheet1", "B1", "Preto e Branco")
	f.SetCellValue("Sheet1", "C1", "Colorido")

	z := zabbix.NewClient()
	c, _ := monitor.GetPrintersCounter(z)

	line := 2
	for _, printer := range c {
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", line), printer.HostData.Host)
		if printer.ColorCounter > 0 {
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", line), printer.BlackCounter)
			f.SetCellValue("Sheet1", fmt.Sprintf("C%d", line), printer.ColorCounter)
		} else {
			f.SetCellValue("Sheet1", fmt.Sprintf("B%d", line), printer.TotalCounter)
		}
		line++
	}

	// Save spreadsheet by the given path.
	if err := f.SaveAs(fileName); err != nil {
		fmt.Println(err)
	}
	return fileName
}

package client

import (
	"bufio"
	"bytes"
	"fmt"
	ui "github.com/gizak/termui"
	"strings"
	"text/tabwriter"
	"time"
)

func DoHtop(cli interface{}, args ...string) error {

	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	list := ui.NewList()
	list.Items = []string{
		"[0] abcdef",
		"[1] abcdef",
		"[2] abcdef",
		"[3] abcdef",
	}
	list.ItemFgColor = ui.ColorYellow
	list.Border.Label = "Containers"
	list.Height = 24
	list.Width = 80
	list.Y = 0

	header := ui.NewPar(" htop: Gattai for Docker ")
	header.HasBorder = false
	header.Width = 25
	header.Height = 1
	header.TextFgColor = ui.ColorBlack
	header.TextBgColor = ui.ColorWhite
	header.X = 53
	header.Y = 0

	statusBar := ui.NewPar(" Quit: Q")
	statusBar.HasBorder = false
	statusBar.Width = 25
	statusBar.Height = 1
	statusBar.TextFgColor = ui.ColorMagenta
	statusBar.X = 0
	statusBar.Y = 24

	evt := ui.EventCh()

	for {
		select {
		case e := <-evt:
			if e.Type == ui.EventKey && e.Ch == 'q' {
				return nil
			}
		default:
			var b bytes.Buffer
			buf := bufio.NewWriter(&b)
			w := tabwriter.NewWriter(buf, 5, 1, 3, ' ', 0)
			fmt.Fprintln(w, "NODE\tNAME\tID\tCPU\tMEM\tTX\tRX")

			w.Flush()
			buf.Flush()

			list.Items = strings.Split(b.String(), "\n")
			ui.Render(list, header, statusBar)
			time.Sleep(time.Second / 4)
		}
	}
	return nil
}

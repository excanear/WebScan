package ui

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"webscan/internal/scanner"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Theme is a small set of colors used by the TUI.
type Theme struct {
	HeaderColor   tcell.Color
	OpenColor     tcell.Color
	ClosedColor   tcell.Color
	FilteredColor tcell.Color
	TitleColor    tcell.Color
	HeaderArt     string
}

func getTheme(name string) Theme {
	switch strings.ToLower(name) {
	case "vaporwave":
		return Theme{
			HeaderColor:   tcell.NewHexColor(0xFF77FF),
			OpenColor:     tcell.NewHexColor(0x66FFCC),
			ClosedColor:   tcell.NewHexColor(0xFF6B6B),
			FilteredColor: tcell.NewHexColor(0xFFD166),
			TitleColor:    tcell.NewHexColor(0xA78BFA),
			HeaderArt: `
 __      __   _  __
 \ \    / /  / |/ /  _   _  ___
  \ \/\/ /   | ' /  | | | |/ _ \
   \_/\_/    |_|\_\  |_| |_|\___/
`,
		}
	case "neo":
		return Theme{
			HeaderColor:   tcell.NewHexColor(0x00FFFF),
			OpenColor:     tcell.NewHexColor(0x00FF99),
			ClosedColor:   tcell.NewHexColor(0xFF5370),
			FilteredColor: tcell.NewHexColor(0xFFD166),
			TitleColor:    tcell.NewHexColor(0x7C4DFF),
			HeaderArt: `
  _   _  ____  ___  ____
 | \ | |/ __ \| \ \/ / _ \
 |  \| | |  | | |\  / (_) |
 |_|\__|\____/|_| \/ \___/
`,
		}
	default:
		return Theme{
			HeaderColor:   tcell.ColorWhite,
			OpenColor:     tcell.ColorGreen,
			ClosedColor:   tcell.ColorRed,
			FilteredColor: tcell.ColorOrange,
			TitleColor:    tcell.ColorLightBlue,
			HeaderArt:     "",
		}
	}
}

// RunTUI runs a simple interactive TUI that streams scan results from the
// provided Scanner. cfg is the scanner configuration and style is a preset
// name (e.g., "vaporwave").
func RunTUI(ctx context.Context, sc *scanner.Scanner, cfg scanner.Config, style string) error {
	app := tview.NewApplication()

	header := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	header.SetBorder(true).SetTitle(" WebScan ")

	table := tview.NewTable().SetFixed(1, 0).SetSelectable(true, false)
	table.SetBorder(true).SetTitle(" Ports ")

	// set header row
	table.SetCell(0, 0, tview.NewTableCell("STATUS").SetSelectable(false).SetAlign(tview.AlignLeft))
	table.SetCell(0, 1, tview.NewTableCell("PORT").SetSelectable(false))
	table.SetCell(0, 2, tview.NewTableCell("PROTO").SetSelectable(false))
	table.SetCell(0, 3, tview.NewTableCell("SERVER").SetSelectable(false))
	table.SetCell(0, 4, tview.NewTableCell("TITLE").SetSelectable(false))

	progress := tview.NewTextView().SetDynamicColors(true)
	progress.SetBorder(true).SetTitle(" Progress ")

	// UI state
	var mu sync.Mutex
	resultsMap := map[int]scanner.PortResult{}
	scanned := 0
	openCount := 0
	filteredCount := 0
	closedCount := 0
	start := time.Now()
	currentFilter := "" // "open", "filtered", "closed", or empty
	searchQuery := ""

	// load persisted TUI config (style, sort)
	tuiCfg, _ := LoadConfig()
	if tuiCfg.Style != "" && style == "default" {
		style = tuiCfg.Style
	}
	currentSort := "port"
	if tuiCfg.Sort != "" {
		currentSort = tuiCfg.Sort
	}
	theme := getTheme(style)
	helpShown := false

	// mapping: port -> table row, and next available row index
	portRow := map[int]int{}
	nextRow := 1

	// apply header art/color if available
	if theme.HeaderArt != "" {
		header.SetText(theme.HeaderArt)
	}
	header.SetTextColor(theme.HeaderColor)

	// detail panel
	detail := tview.NewTextView().SetDynamicColors(true)
	detail.SetBorder(true).SetTitle(" Details ")

	// content layout: table (left) + detail (right)
	contentFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(table, 0, 3, true).
		AddItem(detail, 0, 2, false)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 3, 0, false).
		AddItem(contentFlex, 0, 1, true).
		AddItem(progress, 3, 0, false)

	// use Pages to allow overlaying input modal for search
	pages := tview.NewPages()
	pages.AddPage("main", flex, true, true)

	// helper to update header/progress
	updateStats := func() {
		mu.Lock()
		defer mu.Unlock()
		elapsed := time.Since(start).Round(time.Millisecond)
		headerText := fmt.Sprintf("[::b]TARGET:[-:-:-] %s    [::b]THREADS:[-:-:-] %d    [::b]STYLE:[-:-:-] %s",
			cfg.Target, cfg.Threads, style)
		// if header art present, keep it and append stats
		if theme.HeaderArt != "" {
			header.SetText(theme.HeaderArt + "\n" + headerText)
		} else {
			header.SetText(headerText)
		}

		progressText := fmt.Sprintf("Scanned: %d   Open: %d   Filtered: %d   Closed: %d   Time: %s   Filter: %s",
			scanned, openCount, filteredCount, closedCount, elapsed, currentFilter)
		progress.SetText(progressText)
	}

	// apply sorting to results and rebuild table rows
	applySort := func() {
		mu.Lock()
		defer mu.Unlock()
		// collect results
		prs := make([]scanner.PortResult, 0, len(resultsMap))
		for _, pr := range resultsMap {
			prs = append(prs, pr)
		}
		// sort based on currentSort
		switch currentSort {
		case "port":
			sort.Slice(prs, func(i, j int) bool { return prs[i].Port < prs[j].Port })
		case "status":
			value := func(p scanner.PortResult) int {
				if p.Open {
					return 2
				}
				if p.Filtered {
					return 1
				}
				return 0
			}
			sort.Slice(prs, func(i, j int) bool {
				vi := value(prs[i])
				vj := value(prs[j])
				if vi == vj {
					return prs[i].Port < prs[j].Port
				}
				return vi > vj
			})
		case "title":
			sort.Slice(prs, func(i, j int) bool { return strings.ToLower(prs[i].Title) < strings.ToLower(prs[j].Title) })
		default:
			sort.Slice(prs, func(i, j int) bool { return prs[i].Port < prs[j].Port })
		}

		// clear previous rows
		// find previous max row
		maxRow := 0
		for _, r := range portRow {
			if r > maxRow {
				maxRow = r
			}
		}
		// rebuild rows
		portRow = map[int]int{}
		for i, pr := range prs {
			row := i + 1
			portRow[pr.Port] = row
			status := "CLOSED"
			statusColor := theme.ClosedColor
			if pr.Open {
				status = "OPEN"
				statusColor = theme.OpenColor
			} else if pr.Filtered {
				status = "FILTERED"
				statusColor = theme.FilteredColor
			}
			table.SetCell(row, 0, tview.NewTableCell(status).SetTextColor(statusColor))
			table.SetCell(row, 1, tview.NewTableCell(strconv.Itoa(pr.Port)))
			table.SetCell(row, 2, tview.NewTableCell(pr.Protocol))
			table.SetCell(row, 3, tview.NewTableCell(pr.Server))
			table.SetCell(row, 4, tview.NewTableCell(pr.Title))
		}
		// clear leftover rows
		for r := len(prs) + 1; r <= maxRow; r++ {
			for c := 0; c <= 4; c++ {
				table.SetCell(r, c, tview.NewTableCell(""))
			}
		}
		nextRow = len(prs) + 1
	}

	// apply current filter to table rows (hide/show rows)
	applyFilter := func() {
		mu.Lock()
		defer mu.Unlock()
		for port, row := range portRow {
			pr, ok := resultsMap[port]
			if !ok {
				continue
			}
			show := false
			switch currentFilter {
			case "open":
				show = pr.Open
			case "filtered":
				show = pr.Filtered
			case "closed":
				show = !pr.Open && !pr.Filtered
			default:
				show = true
			}

			if !show {
				for c := 0; c <= 4; c++ {
					table.SetCell(row, c, tview.NewTableCell(""))
				}
				continue
			}

			// apply search query if present
			if searchQuery != "" {
				q := strings.ToLower(searchQuery)
				matched := false
				if strings.Contains(strings.ToLower(strconv.Itoa(pr.Port)), q) {
					matched = true
				}
				if !matched && strings.Contains(strings.ToLower(pr.Server), q) {
					matched = true
				}
				if !matched && strings.Contains(strings.ToLower(pr.Title), q) {
					matched = true
				}
				if !matched && len(pr.Headers) > 0 {
					for _, vals := range pr.Headers {
						for _, v := range vals {
							if strings.Contains(strings.ToLower(v), q) {
								matched = true
								break
							}
						}
						if matched {
							break
						}
					}
				}
				if !matched {
					for c := 0; c <= 4; c++ {
						table.SetCell(row, c, tview.NewTableCell(""))
					}
					continue
				}
			}

			// restore visible row cells from resultsMap
			status := "CLOSED"
			statusColor := theme.ClosedColor
			if pr.Open {
				status = "OPEN"
				statusColor = theme.OpenColor
			} else if pr.Filtered {
				status = "FILTERED"
				statusColor = theme.FilteredColor
			}
			proto := pr.Protocol
			if proto == "" {
				proto = "-"
			}
			server := pr.Server
			if server == "" {
				server = "-"
			}
			title := pr.Title
			if title == "" {
				title = "-"
			}
			table.SetCell(row, 0, tview.NewTableCell(status).SetTextColor(statusColor))
			table.SetCell(row, 1, tview.NewTableCell(strconv.Itoa(pr.Port)))
			table.SetCell(row, 2, tview.NewTableCell(proto))
			table.SetCell(row, 3, tview.NewTableCell(server))
			table.SetCell(row, 4, tview.NewTableCell(title))
		}
	}

	// channel for streaming results
	ch := make(chan scanner.PortResult)

	// start scan in goroutine
	go func() {
		_ = sc.StartStream(ctx, ch)
	}()

	// consume results and update UI
	go func() {
		for pr := range ch {
			mu.Lock()
			// store latest result
			resultsMap[pr.Port] = pr
			scanned++
			if pr.Open {
				openCount++
			} else if pr.Filtered {
				filteredCount++
			} else {
				closedCount++
			}

			row, ok := portRow[pr.Port]
			if !ok {
				row = nextRow
				portRow[pr.Port] = row
				nextRow++
			}

			// prepare cells
			status := "CLOSED"
			statusColor := theme.ClosedColor
			if pr.Open {
				status = "OPEN"
				statusColor = theme.OpenColor
			} else if pr.Filtered {
				status = "FILTERED"
				statusColor = theme.FilteredColor
			}

			proto := pr.Protocol
			if proto == "" {
				proto = "-"
			}
			server := pr.Server
			if server == "" {
				server = "-"
			}
			title := pr.Title
			if title == "" {
				title = "-"
			}

			// update table on main goroutine and apply filter
			app.QueueUpdateDraw(func() {
				table.SetCell(row, 0, tview.NewTableCell(status).SetTextColor(statusColor))
				table.SetCell(row, 1, tview.NewTableCell(strconv.Itoa(pr.Port)))
				table.SetCell(row, 2, tview.NewTableCell(proto))
				table.SetCell(row, 3, tview.NewTableCell(server))
				table.SetCell(row, 4, tview.NewTableCell(title))
				updateStats()
				applyFilter()
			})
			mu.Unlock()
		}

		// when channel closes, show complete message and sort table rows by port
		app.QueueUpdateDraw(func() {
			updateStats()
			// sort rows by port number
			ports := make([]int, 0, len(portRow))
			for p := range portRow {
				ports = append(ports, p)
			}
			sort.Ints(ports)
			// rebuild table body
			for i, p := range ports {
				srcRow := portRow[p]
				tgtRow := i + 1
				if srcRow != tgtRow {
					for c := 0; c <= 4; c++ {
						cell := table.GetCell(srcRow, c)
						table.SetCell(tgtRow, c, cell)
						table.SetCell(srcRow, c, tview.NewTableCell(""))
					}
					portRow[p] = tgtRow
				}
			}
			// re-apply any active filter after reordering
			applyFilter()
			// append footer
			header.SetText(header.GetText(true) + "    [::b]SCAN COMPLETE[-:-:-]")
		})
	}()

	// interactive keys: quit, filters and details
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// quit
		switch event.Key() {
		case tcell.KeyEsc:
			app.Stop()
			return nil
		case tcell.KeyEnter:
			// show details for selected row
			r, _ := table.GetSelection()
			// find port for row
			var portForRow int
			for p, rr := range portRow {
				if rr == r {
					portForRow = p
					break
				}
			}
			if portForRow != 0 {
				if pr, ok := resultsMap[portForRow]; ok {
					var b strings.Builder
					b.WriteString(fmt.Sprintf("Port: %d\nOpen: %t\nProtocol: %s\nStatus: %d\nServer: %s\nTitle: %s\nSize: %d\n", pr.Port, pr.Open, pr.Protocol, pr.Status, pr.Server, pr.Title, pr.Size))
					if pr.TLSVersion != "" {
						b.WriteString(fmt.Sprintf("TLS: %s  Cipher: %s\n", pr.TLSVersion, pr.CipherSuite))
					}
					if pr.CDN != "" || pr.WAF != "" || len(pr.Technologies) > 0 {
						b.WriteString("Fingerprint:\n")
						if pr.CDN != "" {
							b.WriteString(fmt.Sprintf("  CDN: %s\n", pr.CDN))
						}
						if pr.WAF != "" {
							b.WriteString(fmt.Sprintf("  WAF: %s (%s)\n", pr.WAF, pr.WAFReason))
						}
						if len(pr.Technologies) > 0 {
							b.WriteString(fmt.Sprintf("  Tech: %s\n", strings.Join(pr.Technologies, ", ")))
						}
					}
					if len(pr.Headers) > 0 {
						b.WriteString("\nHeaders:\n")
						for k, v := range pr.Headers {
							b.WriteString(fmt.Sprintf("  %s: %s\n", k, strings.Join(v, "; ")))
						}
					}
					detail.SetText(b.String())
				}
			}
			return nil
		}

		// if help overlay visible, close it on any key
		if helpShown {
			pages.RemovePage("help")
			helpShown = false
			return nil
		}

		// rune-based actions
		if r := event.Rune(); r != 0 {
			switch r {
			case 'q', 'Q':
				app.Stop()
				return nil
			case '?', 'h':
				// show help overlay
				helpText := "" +
					"Keys:\n" +
					" q / Esc  — quit\n" +
					" Enter    — show details for selected port\n" +
					" /        — search (type and Enter)\n" +
					" s        — cycle sort (port → status → title)\n" +
					" o        — filter open ports\n" +
					" f        — filter filtered ports\n" +
					" c        — filter closed ports\n" +
					" a        — show all (clear filter)\n" +
					" h / ?    — show this help\n" +
					" Press any key to close\n"
				helpView := tview.NewTextView().SetDynamicColors(true).SetText(helpText)
				helpView.SetBorder(true).SetTitle(" Help ")
				modal := tview.NewFlex().SetDirection(tview.FlexRow).
					AddItem(nil, 0, 1, false).
					AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
						AddItem(nil, 0, 1, false).
						AddItem(helpView, 40, 0, true).
						AddItem(nil, 0, 1, false), 10, 0, true).
					AddItem(nil, 0, 1, false)
				pages.AddPage("help", modal, true, true)
				helpShown = true
				app.SetFocus(helpView)
				return nil
			case '/':
				// show search input overlay
				input := tview.NewInputField().SetLabel("/ ").SetText(searchQuery)
				input.SetDoneFunc(func(key tcell.Key) {
					if key == tcell.KeyEnter {
						searchQuery = input.GetText()
						pages.RemovePage("search")
						app.SetFocus(table)
						app.QueueUpdateDraw(func() {
							updateStats()
							applyFilter()
						})
					} else if key == tcell.KeyEsc {
						pages.RemovePage("search")
						app.SetFocus(table)
					}
				})
				// center the input in a small box
				modal := tview.NewFlex().SetDirection(tview.FlexRow).
					AddItem(nil, 0, 1, false).
					AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
						AddItem(nil, 0, 1, false).
						AddItem(input, 3, 0, true).
						AddItem(nil, 0, 1, false), 3, 0, true).
					AddItem(nil, 0, 1, false)
				pages.AddPage("search", modal, true, true)
				app.SetFocus(input)
			case 's', 'S':
				// cycle sort: port -> status -> title -> port
				switch currentSort {
				case "port":
					currentSort = "status"
				case "status":
					currentSort = "title"
				default:
					currentSort = "port"
				}
				applySort()
				// persist choice
				_ = SaveConfig(TUIConfig{Style: style, Sort: currentSort})
			case 'o':
				currentFilter = "open"
			case 'f':
				currentFilter = "filtered"
			case 'c':
				currentFilter = "closed"
			case 'a':
				currentFilter = ""
			}
			// update status/progress and apply filter
			app.QueueUpdateDraw(func() {
				updateStats()
				applyFilter()
			})
		}
		return event
	})

	// run app
	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		return err
	}
	return nil
}

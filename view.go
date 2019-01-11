package main

import (
  "fmt"
  "github.com/gdamore/tcell"
  "github.com/rivo/tview"
  "strconv"
)

func ticketPreview(tick ticket) string {
  return fmt.Sprintf(`date: %s
-----------------------------------
%s`, tick.CreatedAt.Format("2006-01-02 15:04:05"), tick.Description)
}

func loadTickets(table *tview.Table, app *tview.Application, flex *tview.Flex, preview *tview.TextView, tickFunc func() ([]ticket, error)) error {
  tickets, err := tickFunc()
  if err != nil {
    return err
  }

  table.
        SetCell(0, 0, tview.NewTableCell("Number").SetSelectable(false)).
        SetCell(0, 1, tview.NewTableCell("Subject").SetSelectable(false).SetExpansion(1)).
        SetCell(0, 2, tview.NewTableCell("Name").SetSelectable(false)).
        SetCell(0, 3, tview.NewTableCell("Org").SetSelectable(false))

  table.SetSelectionChangedFunc(func(row, column int) {
    app.QueueUpdateDraw(func() {
      text := ""
      preview.SetTitle(" preview ")
      if row > 0 && row <= len(tickets) {
        text = ticketPreview(tickets[row-1])
      }
      preview.SetText(text)
      preview.SetWordWrap(true)
      preview.ScrollToBeginning()
    })
  })

  for i,ticket := range tickets {
    j := i+1

    table.
          SetCell(j, 0, tview.NewTableCell(strconv.Itoa(ticket.Id))).
          SetCell(j, 1, tview.NewTableCell(ticket.Subject)).
          SetCell(j, 2, tview.NewTableCell(ticket.Requester)).
          SetCell(j, 3, tview.NewTableCell(ticket.Org))
  }
  table.SetFixed(1, 0)
  if len(tickets) > 0 {
    preview.SetText(ticketPreview(tickets[0]))
  }
  app.QueueUpdateDraw(func() {
    app.SetRoot(flex, true)
  })
  return nil
}

func loading(modal *tview.Modal, app *tview.Application, loadingChan chan string, killCh chan struct{}) {
  i := 0
  foreverMsg := []string{"Loading ", "tickets ", "takes ", "F", "O", "R", "E", "V", "E", "R"}
  output := ""
  for {
    select {
    case <-killCh:
      return
    case s := <-loadingChan:
      if s != "loading" {
        return
      }
      if len(foreverMsg) > i {
        output = output + foreverMsg[i]
      } else {
        output = output + "."
      }
      i++
      app.QueueUpdateDraw(func() {
        modal.SetText(output)
      })
    }

  }
}

func runUI(loadingChan chan string, tickFunc func() ([]ticket, error)) error {
  app := tview.NewApplication()
  table := tview.NewTable().
            SetBorders(false).
            SetSelectable(true, false)
  modal := tview.NewModal()

  flex := tview.NewFlex().SetDirection(tview.FlexRow)
  preview := tview.NewTextView()
  flex.AddItem(table, 0, 1, true)
  flex.AddItem(preview, 0, 2, false)

  preview.SetBorder(true).SetTitle("preview")

  table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
    if event.Key() == tcell.KeyTab {
      app.QueueUpdateDraw(func() {
        app.SetFocus(preview)
      })
    }
    return event
  })

  preview.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
    if event.Key() == tcell.KeyTab {
      app.QueueUpdateDraw(func() {
        app.SetFocus(table)
      })

    }
    return event
  })

  killLoading := make(chan struct{})

  go func() {
    loading(modal, app, loadingChan, killLoading)
  }()

  go func() {
    err := loadTickets(table, app, flex, preview, tickFunc)

    if err != nil {
      close(killLoading)
      app.QueueUpdateDraw(func() {
        modal.SetText("error:\n\n" + err.Error() + "\n\nctrl-c to exit")
      })
    }
  }()

  return app.SetRoot(modal, true).SetFocus(modal).Run()
}

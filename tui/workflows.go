package tui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/lupinelab/circlog/circleci"
	"github.com/rivo/tview"
)

type workflowsTable struct {
	table *tview.Table
}

func (cTui *CirclogTui) newWorkflowsTable() workflowsTable {
	table := tview.NewTable().SetSelectable(true, false).SetFixed(1, 0).SetSeparator(tview.Borders.Vertical)
	table.SetTitle(" WORKFLOWS ").SetBorder(true)

	for column, header := range []string{"Name", "Duration"} {
		table.SetCell(0, column, tview.NewTableCell(header).SetStyle(tcell.StyleDefault.Attributes(tcell.AttrBold)).SetSelectable(false))
	}

	table.SetSelectedFunc(func(row int, col int) {
		cell := table.GetCell(row, 0)
		cellRef := cell.GetReference()
		switch cellRef := cellRef.(type) {
		case circleci.Workflow:
			cTui.tuiState.workflow = cellRef
			jobs, nextPageToken, _ := circleci.GetWorkflowJobs(cTui.config, cTui.tuiState.workflow.Id, 1, "")
			cTui.jobs.populateTable(jobs, nextPageToken)
			cTui.app.SetFocus(cTui.jobs.table)
		case string:
			if cell.Text == "..." {
				nextPageToken := cell.GetReference().(string)
				newWorkflows, nextPageToken, _ := circleci.GetPipelineWorkflows(cTui.config, cTui.tuiState.pipeline.Id, 1, nextPageToken)
				cTui.workflows.addWorkflowsToTable(newWorkflows, table.GetRowCount(), nextPageToken)
			}
		}
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			cTui.workflows.clear()
			cTui.app.SetFocus(cTui.pipelines.table)
		}

		if event.Rune() == 'b' {
			cTui.app.SetFocus(cTui.branchSelect)
		}

		if event.Rune() == 'd' {
			cTui.app.Stop()
			fmt.Printf("circlog workflows %s -l %s\n", cTui.config.Project, cTui.tuiState.pipeline.Id)
		}

		return event
	})

	table.SetFocusFunc(func() {
		cTui.controls.SetText(cTui.controlBindings)
	})

	return workflowsTable{
		table: table,
	}
}

func (w workflowsTable) populateWorkflowsTable(workflows []circleci.Workflow, nextPageToken string) {
	w.clear()
	w.addWorkflowsToTable(workflows, 1, nextPageToken)
}

func (w workflowsTable) addWorkflowsToTable(workflows []circleci.Workflow, startRow int, nextPageToken string) {
	if len(workflows) != 0 {
		for row, workflow := range workflows {
			var workflowDuration string
			if workflow.Status == circleci.RUNNING {
				workflowDuration = time.Since(workflow.CreatedAt).Round(time.Millisecond).String()
			} else {
				workflowDuration = workflow.StoppedAt.Sub(workflow.CreatedAt).Round(time.Millisecond).String()
			}

			for column, attr := range []string{workflow.Name, workflowDuration} {
				cell := tview.NewTableCell(attr).SetStyle(styleForStatus(workflow.Status))
				cell.SetReference(workflow)
				w.table.SetCell(row+1, column, cell)
			}
		}

		if nextPageToken != "" {
			cell := tview.NewTableCell("...")
			cell.SetReference(nextPageToken)
			w.table.SetCell(w.table.GetRowCount(), 0, cell)
		}

	} else {
		cell := tview.NewTableCell("None").SetStyle(tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDarkGray))
		w.table.SetCell(1, 0, cell)
	}
}

func (w workflowsTable) clear() {
	row := 1
	for row < w.table.GetRowCount() {
		w.table.RemoveRow(row)
	}
}
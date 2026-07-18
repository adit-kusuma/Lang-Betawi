//go:build windows

package installer

import (
	"github.com/lxn/walk"
	dec "github.com/lxn/walk/declarative"
)

type SetupError struct{ Message string }

func (e *SetupError) Error() string { return e.Message }

func Run() error {
	installDir, dirErr := InstallDir()
	existingFound := dirErr == nil && DetectExistingInstall(installDir)

	var mw *walk.MainWindow
	var pageWelcome, pageExisting, pageProgress, pageFinish *walk.Composite
	var btnBack, btnNext, btnCancel *walk.PushButton
	var lblExistingBody *walk.Label
	var rbOverwrite, rbRepair *walk.RadioButton
	var lblProgressStatus *walk.Label
	var progressBar *walk.ProgressBar
	var lblFinishTitle, lblFinishBody *walk.Label

	err := dec.MainWindow{
		AssignTo: &mw,
		Title:    "Language Betawi Setup",
		MinSize:  dec.Size{Width: 520, Height: 420},
		Size:     dec.Size{Width: 520, Height: 420},
		Layout:   dec.VBox{},
		Children: []dec.Widget{

			dec.Composite{
				AssignTo: &pageWelcome,
				Layout:   dec.VBox{},
				Children: []dec.Widget{
					dec.Label{Text: "Setup - Language Betawi", Font: dec.Font{PointSize: 12, Bold: true}},
					dec.VSpacer{Size: 12},
					dec.Label{
						Text: "Welcome to the Language Betawi Setup Wizard.\r\n\r\n" +
							"This will install the Language Betawi compiler on your computer " +
							"and register it in your System PATH so you can run 'betawi' from " +
							"any Command Prompt or PowerShell window.\r\n\r\n" +
							"Click Next to continue, or Cancel to exit Setup.",
					},
				},
			},

			dec.Composite{
				AssignTo: &pageExisting,
				Visible:  false,
				Layout:   dec.VBox{},
				Children: []dec.Widget{
					dec.Label{
						Text: "Existing installation",
						Font: dec.Font{PointSize: 12, Bold: true},
					},
					dec.VSpacer{Size: 12},
					dec.Label{AssignTo: &lblExistingBody},
					dec.VSpacer{Size: 12},

					dec.RadioButton{
						AssignTo: &rbOverwrite,
						Text:     "Overwrite the existing installation",
					},
					dec.RadioButton{
						AssignTo: &rbRepair,
						Text:     "Repair the existing installation",
					},
				},
			},

			dec.Composite{
				AssignTo: &pageProgress,
				Visible:  false,
				Layout:   dec.VBox{},
				Children: []dec.Widget{
					dec.Label{Text: "Installing", Font: dec.Font{PointSize: 12, Bold: true}},
					dec.VSpacer{Size: 12},
					dec.Label{AssignTo: &lblProgressStatus, Text: "Preparing..."},
					dec.VSpacer{Size: 8},
					dec.ProgressBar{AssignTo: &progressBar, MinValue: 0, MaxValue: 100},
				},
			},

			dec.Composite{
				AssignTo: &pageFinish,
				Visible:  false,
				Layout:   dec.VBox{},
				Children: []dec.Widget{
					dec.Label{AssignTo: &lblFinishTitle, Font: dec.Font{PointSize: 12, Bold: true}},
					dec.VSpacer{Size: 12},
					dec.Label{AssignTo: &lblFinishBody},
				},
			},

			dec.Composite{
				Layout: dec.HBox{},
				Children: []dec.Widget{
					dec.HSpacer{},
					dec.PushButton{AssignTo: &btnBack, Text: "< Back", Enabled: false},
					dec.PushButton{AssignTo: &btnNext, Text: "Next >"},
					dec.PushButton{AssignTo: &btnCancel, Text: "Cancel"},
				},
			},
		},
	}.Create()
	if err != nil {
		return err
	}

	pages := []*walk.Composite{pageWelcome}
	if existingFound {
		lblExistingBody.SetText(
			"An existing Language Betawi installation has been found at " + installDir + ".\r\n\r\n" +
				"Choose how you'd like to proceed:",
		)
		pages = append(pages, pageExisting)
	}
	pages = append(pages, pageProgress, pageFinish)

	progressPageIdx := len(pages) - 2
	finishPageIdx := len(pages) - 1
	currentPageIdx := 0

	showPage := func(idx int) {
		for i, p := range pages {
			p.SetVisible(i == idx)
		}
	}
	showPage(0)

	runInstall := func() {
		repairMode := existingFound && rbRepair.Checked()

		go func() {
			setupErr := RunSetup(repairMode, func(progress float64, status string) {
				mw.Synchronize(func() {
					progressBar.SetValue(int(progress * 100))
					lblProgressStatus.SetText(status)
				})
			})

			mw.Synchronize(func() {
				currentPageIdx = finishPageIdx
				showPage(currentPageIdx)

				if setupErr != nil {
					lblFinishTitle.SetText("Setup Failed")
					lblFinishBody.SetText(setupErr.Error())
				} else {
					lblFinishTitle.SetText("Completing the Language Betawi Setup Wizard")
					lblFinishBody.SetText(
						"Betawi language successfully downloaded.\r\n\r\n" +
							"You can now use 'betawi' from any Command Prompt or PowerShell window.",
					)
				}

				btnNext.SetText("Finish")
				btnNext.SetEnabled(true)
				btnBack.SetEnabled(false)
				btnCancel.SetEnabled(false)
			})
		}()
	}

	btnNext.Clicked().Attach(func() {
		if currentPageIdx == finishPageIdx {
			mw.Close()
			return
		}
		if currentPageIdx == progressPageIdx {
			return
		}

		currentPageIdx++
		showPage(currentPageIdx)
		btnBack.SetEnabled(currentPageIdx != 0 && currentPageIdx != progressPageIdx)

		if currentPageIdx == progressPageIdx {
			btnNext.SetEnabled(false)
			btnBack.SetEnabled(false)
			btnCancel.SetEnabled(false)
			runInstall()
		}
	})

	btnBack.Clicked().Attach(func() {
		if currentPageIdx <= 0 || currentPageIdx == progressPageIdx || currentPageIdx == finishPageIdx {
			return
		}
		currentPageIdx--
		showPage(currentPageIdx)
		btnBack.SetEnabled(currentPageIdx != 0)
	})

	btnCancel.Clicked().Attach(func() {
		mw.Close()
	})

	mw.Run()
	return nil
}

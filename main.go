package main

import (
	"log"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/micmonay/keybd_event"
)

func main() {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		panic(err)
	}
	kb.SetKeys(keybd_event.VK_SCROLLLOCK)

	var running int32 = 0
	stopChan := make(chan struct{})

	// UI elements we update from callbacks
	statusLabel := widget.NewLabel("Status: Stopped")
	toggleBtn := widget.NewButton("Start", nil)
	interval := widget.NewEntry()
	interval.SetText("60") // default seconds

	startLoop := func() {
		if !atomic.CompareAndSwapInt32(&running, 0, 1) {
			return
		}

		statusLabel.SetText("Status: Running")
		toggleBtn.SetText("Stop")

		go func() {
			log.Println("Loop started")

			for {
				select {
				case <-stopChan:
					log.Println("Loop stopped")
					return

				default:
					log.Println("Poke")

					if err := kb.Press(); err != nil {
						panic(err)
					}
					time.Sleep(100 * time.Millisecond)

					if err := kb.Release(); err != nil {
						panic(err)
					}

					// Parse interval
					secs, err := time.ParseDuration(interval.Text + "s")
					if err != nil {
						secs = 60 * time.Second
					}

					time.Sleep(secs)
				}
			}
		}()
	}

	stopLoop := func() {
		if !atomic.CompareAndSwapInt32(&running, 1, 0) {
			return
		}

		stopChan <- struct{}{}
		statusLabel.SetText("Status: Stopped")
		toggleBtn.SetText("Start")
	}

	// Toggle button behaviour
	toggleBtn.OnTapped = func() {
		if atomic.LoadInt32(&running) == 0 {
			startLoop()
		} else {
			stopLoop()
		}
	}

	// --- UI ---
	a := app.New()
	w := a.NewWindow("Poker")

	// System tray icon
	if desk, ok := a.(desktop.App); ok {
		desk.SetSystemTrayMenu(fyne.NewMenu("Poker",
			fyne.NewMenuItem("Show", func() {
				w.Show()
				w.RequestFocus()
			}),
			fyne.NewMenuItem("Start", func() { startLoop() }),
			fyne.NewMenuItem("Stop", func() { stopLoop() }),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() { a.Quit() }),
		))
	}

	// Minimise to tray instead of hiding
	w.SetCloseIntercept(func() {
		w.Hide()
	})

	intervalBox := container.NewHBox(
		widget.NewLabel("Interval (seconds):"),
		interval,
	)

	w.SetContent(container.NewVBox(
		toggleBtn,
		statusLabel,
		intervalBox,
	))

	w.Resize(fyne.NewSize(250, 150))
	startLoop()
	w.ShowAndRun()
}

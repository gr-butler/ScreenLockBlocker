package main

import (
	"log"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/micmonay/keybd_event"
)

func main() {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		panic(err)
	}

	kb.SetKeys(keybd_event.VK_SCROLLLOCK)

	var running int32 = 0 // atomic flag
	stopChan := make(chan struct{})

	startLoop := func() {
		if !atomic.CompareAndSwapInt32(&running, 0, 1) {
			return // already running
		}

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
					time.Sleep(60 * time.Second)
				}
			}
		}()
	}

	stopLoop := func() {
		if !atomic.CompareAndSwapInt32(&running, 1, 0) {
			return // already stopped
		}
		stopChan <- struct{}{}
	}

	// --- UI ---
	a := app.New()
	w := a.NewWindow("Poke Controller")

	startBtn := widget.NewButton("Start", func() {
		startLoop()
	})

	stopBtn := widget.NewButton("Stop", func() {
		stopLoop()
	})

	w.SetContent(container.NewVBox(startBtn, stopBtn))
	w.ShowAndRun()
}

package main

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/micmonay/keybd_event"
)

var (
	user32                      = syscall.NewLazyDLL("user32.dll")
	procGetKeyState             = user32.NewProc("GetKeyState")
	kernel32                    = syscall.NewLazyDLL("kernel32.dll")
	procSetThreadExecutionState = kernel32.NewProc("SetThreadExecutionState")
)

const (
	esContinuous      = 0x80000000
	esSystemRequired  = 0x00000001
	esDisplayRequired = 0x00000002
)

// keyOption describes a togglable lock key. A keybdKey of 0 means no key is
// toggled (rely solely on SetThreadExecutionState).
type keyOption struct {
	name     string
	vkCode   int // virtual key code used by GetKeyState
	keybdKey int // key code used by keybd_event
}

var keyOptions = []keyOption{
	{name: "None", vkCode: 0, keybdKey: 0},
	{name: "Scroll Lock", vkCode: 0x91, keybdKey: keybd_event.VK_SCROLLLOCK},
	{name: "Num Lock", vkCode: 0x90, keybdKey: keybd_event.VK_NUMLOCK},
	{name: "Caps Lock", vkCode: 0x14, keybdKey: keybd_event.VK_CAPSLOCK},
}

func keyOptionNames() []string {
	names := make([]string, len(keyOptions))
	for i, opt := range keyOptions {
		names[i] = opt.name
	}
	return names
}

func lookupKeyOption(name string) keyOption {
	for _, opt := range keyOptions {
		if opt.name == name {
			return opt
		}
	}
	return keyOptions[0]
}

// this is the documented way to prevent the system from sleeping, but it doesn't work when the session is locked. So we also toggle a lock key to keep the session awake.
func setKeepAwake(enable bool) {
	if enable {
		procSetThreadExecutionState.Call(uintptr(esDisplayRequired | esSystemRequired | esContinuous))
	} else {
		procSetThreadExecutionState.Call(uintptr(esContinuous))
	}
}

func pressAndReleaseKey(kb keybd_event.KeyBonding, keybdKey int) error {
	kb.SetKeys(keybdKey)

	if err := kb.Press(); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	if err := kb.Release(); err != nil {
		return err
	}

	return nil
}

func isKeyOn(vkCode int) bool {
	state, _, _ := procGetKeyState.Call(uintptr(vkCode))
	return state&0x1 == 1
}

func ensureKeyOn(kb keybd_event.KeyBonding, opt keyOption) error {
	if isKeyOn(opt.vkCode) {
		return nil
	}

	if err := pressAndReleaseKey(kb, opt.keybdKey); err != nil {
		return err
	}

	if !isKeyOn(opt.vkCode) {
		return fmt.Errorf("failed to enable %s", opt.name)
	}

	return nil
}

func main() {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		panic(err)
	}

	var running int32 = 0
	stopChan := make(chan struct{})

	// Currently selected key to toggle, protected by keyMu.
	var keyMu sync.Mutex
	selectedKey := keyOptions[1] // default: Scroll Lock

	// UI elements we update from callbacks
	statusLabel := widget.NewLabel("Status: Stopped")
	toggleBtn := widget.NewButton("Start", nil)
	interval := widget.NewEntry()
	interval.SetText("60") // default seconds

	keySelect := widget.NewSelect(keyOptionNames(), func(name string) {
		keyMu.Lock()
		selectedKey = lookupKeyOption(name)
		keyMu.Unlock()
	})
	keySelect.SetSelected(selectedKey.name)

	startLoop := func() {
		if !atomic.CompareAndSwapInt32(&running, 0, 1) {
			return
		}

		statusLabel.SetText("Status: Running")
		toggleBtn.SetText("Stop")
		setKeepAwake(true)

		go func() {
			log.Println("Loop started")

			for {
				select {
				case <-stopChan:
					log.Println("Loop stopped")
					return

				default:
					log.Println("Poke")

					keyMu.Lock()
					opt := selectedKey
					keyMu.Unlock()

					if opt.keybdKey != 0 {
						if err := pressAndReleaseKey(kb, opt.keybdKey); err != nil {
							log.Printf("pressAndReleaseKey error (session may be locked): %v", err)
							time.Sleep(5 * time.Second)
							continue
						}

						if err := ensureKeyOn(kb, opt); err != nil {
							log.Printf("ensureKeyOn error (session may be locked): %v", err)
							time.Sleep(5 * time.Second)
							continue
						}
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
		setKeepAwake(false)
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

	keyBox := container.NewHBox(
		widget.NewLabel("Toggle key:"),
		keySelect,
	)

	w.SetContent(container.NewVBox(
		toggleBtn,
		statusLabel,
		keyBox,
		intervalBox,
	))

	w.Resize(fyne.NewSize(250, 180))
	startLoop()
	w.ShowAndRun()
}

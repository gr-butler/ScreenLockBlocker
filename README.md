# ScreenLockBlocker (Wobbler)

A small Windows tray application that prevents the screen from locking or going
to sleep. It keeps the system awake using the Windows `SetThreadExecutionState`
API and can optionally toggle a lock key (Scroll Lock, Num Lock, or Caps Lock)
on a configurable interval.

## Features

- Keeps the display and system awake while running (`SetThreadExecutionState`).
- Optional key toggling — choose **None**, **Scroll Lock**, **Num Lock**, or
  **Caps Lock** from the UI.
- Configurable interval (in seconds).
- Runs in the system tray with Start/Stop/Show/Quit controls.

## Prerequisites

### 1. Go

Install Go (1.25.1 or newer) from <https://go.dev/dl/> and make sure `go` is on
your `PATH`:

```sh
go version
```

### 2. A C compiler (GCC)

Fyne on Windows requires **CGO** and a C compiler (GCC). The recommended way to
get one is **MSYS2 with MinGW-w64**.

1. Download and install MSYS2 from <https://www.msys2.org/> and follow the
   installer instructions.

2. Open the **MSYS2 MINGW64** shell and update the package database:

   ```sh
   pacman -Syu
   ```

   (Close and reopen the shell if it asks you to, then run `pacman -Syu` again.)

3. Install the MinGW-w64 GCC toolchain:

   ```sh
   pacman -S mingw-w64-x86_64-gcc
   ```

4. Add the MinGW-w64 `bin` directory to your Windows `PATH` so `gcc` is found:

   ```
   C:\msys64\mingw64\bin
   ```

5. Verify GCC is available from a normal terminal:

   ```sh
   gcc --version
   ```

### 3. Enable CGO

Fyne needs CGO turned on when building:

```sh
go env -w CGO_ENABLED=1
```

## Building

Clone the repository and fetch dependencies:

```sh
git clone https://github.com/gr-butler/ScreenLockBlocker.git
cd ScreenLockBlocker
go mod download
```

### Build an executable

Using the Makefile:

```sh
make build
```

Or directly with Go:

```sh
go build -o wobbles.grbutler.exe .
```

### Package as a Windows GUI app (optional)

To produce a packaged GUI application with an icon, install the Fyne CLI tool
and run the `gui` target:

```sh
go install fyne.io/fyne/v2/cmd/fyne@latest
make gui
```

This runs `fyne package --name wobbles.grbutler --icon Icon.png`.

## Running

Run the built executable:

```sh
./wobbles.grbutler.exe
```

The app starts automatically and appears in the system tray. Use the window or
tray menu to Start/Stop, pick a toggle key, and set the interval.

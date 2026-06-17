APP_NAME := wobbles.grbutler

.PHONY: build gui

build:
	go build -o $(APP_NAME).exe .

gui:
	fyne package --name $(APP_NAME) --icon Icon.png

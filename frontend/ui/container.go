package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

// BackgroundContainer cria um container com cor de fundo sólida
func BackgroundContainer(content fyne.CanvasObject, bgColor color.Color) fyne.CanvasObject {
	rect := canvas.NewRectangle(bgColor)
	return container.NewStack(rect, content)
}

// ChatScrollContainer gerencia a área de scroll das mensagens
func ChatScrollContainer(messages *fyne.Container) *container.Scroll {
	scroll := container.NewVScroll(messages)
	return scroll
}

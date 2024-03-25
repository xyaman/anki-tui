package image

import (
	"image"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/disintegration/imaging"
	"github.com/lucasb-eyer/go-colorful"
)

type Model struct {
	image       image.Image
	imageString string

	width      int
	height     int
	prevWidth  int
	prevHeight int
}

// New creates a new image model
func New() Model {
	return Model{}
}

// ToString converts an image to a string representation of an image.
func ToString(width int, img image.Image) string {
	img = imaging.Resize(img, width, 0, imaging.Lanczos)
	b := img.Bounds()
	imageWidth := b.Max.X
	h := b.Max.Y
	str := strings.Builder{}

	for heightCounter := 0; heightCounter < h; heightCounter += 2 {
		for x := imageWidth; x < width; x += 2 {
			str.WriteString(" ")
		}

		for x := 0; x < imageWidth; x++ {
			c1, _ := colorful.MakeColor(img.At(x, heightCounter))
			color1 := lipgloss.Color(c1.Hex())
			c2, _ := colorful.MakeColor(img.At(x, heightCounter+1))
			color2 := lipgloss.Color(c2.Hex())
			str.WriteString(lipgloss.NewStyle().Foreground(color1).
				Background(color2).Render("▀"))
		}

		str.WriteString("\n")
	}

	return str.String()
}

func (m *Model) SetImage(img image.Image) {
	if img == nil {
		m.image = nil
		m.imageString = lipgloss.Place(40, 20, lipgloss.Center, lipgloss.Center, "no image")
		return
	}

	imageString := ToString(m.width, img)
	m.image = img
	m.imageString = imageString
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.prevWidth = m.width
	m.prevHeight = m.height
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {

	if (m.width != m.prevWidth) || (m.height != m.prevHeight) {
		m.SetImage(m.image)
		m.prevWidth = m.width
		m.prevHeight = m.height
	}

	return m, nil
}

func (m Model) View() string {
	return m.imageString
}

package main

import (
	"fmt"
	"strconv"

	"github.com/google/gxui/math"
	"github.com/hajimehoshi/oto"

	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/gxfont"
	"github.com/google/gxui/samples/flags"
)

const (
	title              = "keyphone v0.2"
	ext                = ".wav"
	model      float32 = 440.0
	sampleRate         = 44100
	width              = 1024
	height             = 320
	maxVolume          = 128
)

var (
	keys = map[int]int{
		1:  66,  // Space
		8:  -21, // 1
		9:  -20, // 2
		10: -19, // 3
		11: -18, // 4
		12: -17, // 5
		13: -16, // 6
		14: -15, // 7
		15: -14, // 8
		16: -13, // 9
		7:  -12, // 0
		4:  -11, // Minus
		18: -10, // Equals
		35: -9,  // Q
		41: -8,  // W
		23: -7,  // E
		36: -6,  // R
		38: -5,  // T
		43: -4,  // Y
		39: -3,  // U
		27: -2,  // I
		33: -1,  // O
		34: 0,   // P
		45: 1,   // Left Bracket
		47: 2,   // Right Bracket
		19: 3,   // A
		37: 4,   // S
		22: 5,   // D
		24: 6,   // F
		25: 7,   // G
		26: 8,   // H
		28: 9,   // J
		29: 10,  // K
		30: 11,  // L
		17: 12,  // Semicolon
		2:  13,  // Apostrophe
		44: 14,  // Z
		42: 15,  // X
		21: 16,  // C
		40: 17,  // V
		20: 18,  // B
		32: 19,  // N
		31: 20,  // M
		3:  21,  // Colon
		5:  22,  // Period
		6:  24,  // Slash
		53: 67,  // Tab
	}

	driver gxui.Driver
	image  gxui.Image
	label  gxui.Label

	volumeStep float32 = 8
	volume     float32 = 64

	out   = make([]byte, 512*8)
	bases = []string{
		"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B",
	}
)

func appMain(d gxui.Driver) {
	driver = d
	theme := flags.CreateTheme(driver)
	window := theme.CreateWindow(width, height, title)
	window.SetScale(flags.DefaultScaleFactor)

	image = theme.CreateImage()

	layout := theme.CreateLinearLayout()
	layout.SetSizeMode(gxui.Fill)

	font32, _ := driver.CreateFont(gxfont.Default, 32)

	label = theme.CreateLabel()
	label.SetFont(font32)
	label.SetColor(gxui.White)

	signal := make(chan int)
	chunk := make(chan []byte)

	player, _ := oto.NewPlayer(sampleRate, 1, 1, sampleRate/20)

	go func() {
		for {
			select {
			case c := <-chunk:
				var b bool
				for !b {
					select {
					case <-signal:
						b = true
					default:
						player.Write(c)
						// fmt.Println(c)
					}
				}
			}
		}
	}()

	window.OnKeyUp(func(ev gxui.KeyboardEvent) {
		if tone, ok := keys[int(ev.Key)]; ok {
			canvas := driver.CreateCanvas(math.Size{W: width, H: 256})
			canvas.Clear(gxui.White)
			drawAxis(canvas)
			canvas.Complete()
			image.SetCanvas(canvas)
			label.SetText("Tab: коричневая нота, осторожно")
			signal <- tone
		}
	})

	window.OnKeyDown(func(ev gxui.KeyboardEvent) {
		if tone, ok := keys[int(ev.Key)]; ok {
			freq := model * math.Powf(2, float32(tone)/12)
			var a1 float32
			var phase float32
			var sample int
			var period int
			for sample = 0; sample < len(out); sample++ {
				a0 := math.Sinf(2 * math.Pi * phase)
				if a1 < 0 && a0 >= 0 {
					period++
					if period > 10 {
						outChunk := out[:sample]
						fmt.Println(len(outChunk))
						canvas := driver.CreateCanvas(math.Size{W: width, H: 256})
						canvas.Clear(gxui.White)
						drawSine(canvas, outChunk, 1)
						drawAxis(canvas)
						canvas.Complete()
						image.SetCanvas(canvas)
						chunk <- outChunk
						break
					}
				}
				out[sample] = byte((a0 + 1) * volume)
				phase += freq / sampleRate
				a1 = a0
			}
			tone += 45
			base := bases[tone%12]
			octave := tone/12 + 1
			noteText := base + strconv.Itoa(octave)
			freqText := strconv.FormatFloat(float64(freq), 'f', 3, 32) + " Гц"
			label.SetText(noteText + " / " + freqText)
		}

		if ev.Key == gxui.KeyUp {
			if volume < maxVolume {
				volume += volumeStep
			}
		} else if ev.Key == gxui.KeyDown {
			if volume > 0 {
				volume -= volumeStep
			}
		}
	})

	window.OnClose(driver.Terminate)

	window.AddChild(layout)
	layout.AddChild(image)
	layout.AddChild(label)
}

func drawSine(canvas gxui.Canvas, chunk []byte, t int) {
	pen := gxui.CreatePen(1, gxui.Red)
	c := len(chunk)
	p := make(gxui.Polygon, c*t)
	for j := 0; j < t; j++ {
		for i := 0; i < c; i++ {
			p[i+j*c] = gxui.PolygonVertex{
				Position: math.Point{
					X: i + j*c,
					Y: int(chunk[i]) + maxVolume - int(volume)},
				RoundedRadius: 0,
			}
		}
	}
	canvas.DrawLines(p, pen)
}

func drawAxis(canvas gxui.Canvas) {
	penHorizontal := gxui.CreatePen(1, gxui.Black)
	line := make(gxui.Polygon, 2)
	line[0] = gxui.PolygonVertex{Position: math.Point{X: 0, Y: 128}}
	line[1] = gxui.PolygonVertex{Position: math.Point{X: width, Y: 128}}
	canvas.DrawLines(line, penHorizontal)
}

func main() {
	gl.StartDriver(appMain)
}

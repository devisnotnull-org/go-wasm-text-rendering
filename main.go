package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"syscall/js"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func getFont(valueDomFont string) string {
	switch valueDomFont {
	case "freedom":
		//fmt.Println("freedom")
		return "http://localhost:8080/assets/freedom.ttf"
	case "shortbaby":
		//fmt.Println("shortbaby")
		return "http://localhost:8080/assets/shortbaby.ttf"
	case "crustyrock":
		//fmt.Println("crustyrock")
		return "http://localhost:8080/assets/crustyrock.ttf"

	default:
		// freebsd, openbsd,
		// plan9, windows...
		//fmt.Printf("%s.\n", valueDomFont)
		return "http://localhost:8080/assets/luximr.ttf"
	}
}

func renderText() js.Func {

	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) any {

		if len(args) != 1 {
			return errors.New("Invalid no of arguments passed")
		}

		jsDoc := js.Global().Get("document")

		if !jsDoc.Truthy() {
			return errors.New("Unable to get document object")
		}

		textinpiut := args[0].String()
		text := strings.Split(textinpiut, " ")

		go func() {
			res, err := http.Get(getFont("luximr"))
			if err != nil {
				fmt.Printf("error making http request: %s\n", err)
			}

			fmt.Printf("client: got response!\n")
			fmt.Printf("client: status code: %d\n", res.StatusCode)

			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)

			f, err := truetype.Parse(body)
			if err != nil {
				log.Println(err)
				return
			}

			fg, bg := image.Black, image.Transparent

			const imgW, imgH = 640, 480
			rgba := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
			draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)

			// Draw the text.
			h := font.HintingNone
			switch "full" {
			case "full":
				h = font.HintingFull
			}

			size := 12.0
			dpi := 72.0
			spacing := 1.5
			title := "Hello, World!"

			d := &font.Drawer{
				Dst: rgba,
				Src: fg,
				Face: truetype.NewFace(f, &truetype.Options{
					Size:    size,
					DPI:     dpi,
					Hinting: h,
				}),
			}

			y := 10 + int(math.Ceil(size*dpi/72))

			dy := int(math.Ceil(size * spacing * dpi / 72))

			d.Dot = fixed.Point26_6{
				X: (fixed.I(imgW) - d.MeasureString(title)) / 2,
				Y: fixed.I(y),
			}

			d.DrawString(title)

			y += dy
			x := 10.0

			dx := int(math.Ceil(size * dpi / dpi))

			for _, s := range text {
				measureString := d.MeasureString(s).Ceil()

				if (x + float64(measureString)) > 640 {
					x = 10.0
					y += dy
				}

				d.Dot = fixed.P(int(x), y)

				d.DrawString(s)

				x += (float64(measureString) + float64(dx))
			}

			buffer := new(bytes.Buffer)
			png.Encode(buffer, rgba)

			bufferPart := buffer.Bytes()

			imageDom := jsDoc.Call("getElementById", "textRender")
			imageDom.Set("src", "data:image/png;base64,"+base64.StdEncoding.EncodeToString(bufferPart))
		}()

		return nil
	})
	return jsonFunc
}

func main() {
	fmt.Println("Go Web Assembly")
	js.Global().Set("renderText", renderText())

	// We need to keep the application running so we keep waiting on a channel
	<-make(chan bool)
}

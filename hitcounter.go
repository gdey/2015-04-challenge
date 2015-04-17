package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"net/http"
	"os"

	"github.com/gdey/hitcounter/counter"
)

var c counter.Value
var mask *image.RGBA
var maskCoords map[rune]image.Point

func loadMask(path string) (m *image.RGBA, err error) {
	f, err := os.Open("images/numbers.png")
	if err != nil {
		return nil, err
	}
	p, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	mask := image.NewRGBA(p.Bounds())
	draw.Draw(mask, p.Bounds(), p, image.ZP, draw.Src)
	return mask, nil
}

func initMask() {
	var err error
	mask, err = loadMask("image/numbersgbtp.png")
	if err != nil {
		panic(nil)
	}
	maskCoords = make(map[rune]image.Point)
	maskCoords['1'] = image.Pt(0, 0)
	maskCoords['2'] = image.Pt(100, 0)
	maskCoords['3'] = image.Pt(200, 0)
	maskCoords['4'] = image.Pt(0, 100)
	maskCoords['5'] = image.Pt(100, 100)
	maskCoords['6'] = image.Pt(200, 100)
	maskCoords['7'] = image.Pt(0, 200)
	maskCoords['8'] = image.Pt(100, 200)
	maskCoords['9'] = image.Pt(200, 200)
	maskCoords['0'] = image.Pt(100, 300)
	maskCoords[','] = image.Pt(0, 300)
	maskCoords['.'] = image.Pt(200, 300)
}

func humanizeString(i int) string {
	stri := []rune(fmt.Sprintf("%d", i))

	l := len(stri)
	if l > 3 {
		n := int(l / 3)
		st := make([]rune, n+l)
		j := 0
		for i := l - 1; i >= 0; i-- {
			if j == 3 {
				j = 0
				st[i+n] = ','
				n--
			}
			j++
			st[i+n] = stri[i]
		}
		stri = st
	}
	return string(stri)
}

func counterHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[9:] // Let's get the id.
		if id == "" {
			http.Error(w, "Identifier Needed", http.StatusNotFound)
			return
		}
		switch r.Method {
		case "GET":
			i := c.Get(id)
			stri := humanizeString(i)
			len := int(len(stri) * 100)
			rec := image.Rect(0, 0, len, 100)
			destimag := image.NewRGBA(rec)
			//blue := color.RGBA{0, 0, 255, 255}
			for j, d := range stri {
				var coord image.Point
				var f bool

				if coord, f = maskCoords[d]; !f {
					fmt.Println("We could not find ", d)
				}
				draw.Draw(destimag, image.Rect(j*100, 0, (j*100)+100, 100), mask, coord, draw.Src)
				//draw.DrawMask(destimag, image.Rect(j*100, 0, (j*100)+100, 100), &image.Uniform{blue}, coord, mask, coord, draw.Over)
			}

			err := png.Encode(w, destimag)
			if err != nil {
				http.Error(w, fmt.Sprintf("Internal error %s", err), http.StatusInternalServerError)
				return
			}
		case "DELETE":
			c.Reset(id)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
}

func main() {

	c = counter.New()
	initMask()

	http.Handle("/counter/", counterHandler())
	log.Fatal(http.ListenAndServe(":8000", nil))
}

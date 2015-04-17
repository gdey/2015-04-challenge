package main

import (
	"fmt"
	"image"
	"image/color"
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

const charwidth = 55

func loadMask(path string) (m *image.RGBA, err error) {
	f, err := os.Open(path)
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
	mask, err = loadMask("images/numberstp.png")
	if err != nil {
		panic(nil)
	}
	maskCoords = make(map[rune]image.Point)
	maskCoords['1'] = image.Pt(25, 0)
	maskCoords['2'] = image.Pt(125, 0)
	maskCoords['3'] = image.Pt(225, 0)
	maskCoords['4'] = image.Pt(25, 100)
	maskCoords['5'] = image.Pt(125, 100)
	maskCoords['6'] = image.Pt(225, 100)
	maskCoords['7'] = image.Pt(25, 200)
	maskCoords['8'] = image.Pt(125, 200)
	maskCoords['9'] = image.Pt(225, 200)
	maskCoords['0'] = image.Pt(125, 300)
	maskCoords[','] = image.Pt(25, 300)
	maskCoords['.'] = image.Pt(225, 300)
}

func humanizeString(i int) string {
	stri := []rune(fmt.Sprintf("%d", i))

	l := len(stri)
	if l <= 3 {
		return string(stri)
	}
	n := int(l / 3)
	st := make([]rune, n+l)
	j := 0
	for i := l - 1; i >= 0; i-- {
		if j == 3 {
			st[i+n] = ','
			j = 0
			n--
		}
		j++
		st[i+n] = stri[i]
	}
	stri = st
	return string(stri)
}
func imageForNumber(i int, fg, bg color.Color) image.Image {
	stri := humanizeString(i)
	len := int(len(stri) * charwidth)
	rec := image.Rect(0, 0, len, 100)
	dest := image.NewRGBA(rec)
	draw.Draw(dest, dest.Bounds(), &image.Uniform{bg}, image.ZP, draw.Src)
	for j, d := range stri {
		var coord image.Point
		var f bool

		if coord, f = maskCoords[d]; !f {
			fmt.Println("We could not find ", d)
		}
		draw.DrawMask(
			dest,
			image.Rect(j*charwidth, 0, (j*charwidth)+charwidth, 100),
			&image.Uniform{fg},
			image.ZP,
			mask,
			coord,
			draw.Over)
	}
	return dest
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
			dest := imageForNumber(i, color.RGBA{0, 0, 255, 255}, color.RGBA{0, 255, 0, 255})
			err := png.Encode(w, dest)
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

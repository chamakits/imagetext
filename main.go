package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"code.google.com/p/freetype-go/freetype"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	fontSize := 12
	r.HandleFunc("/hello/{words}", acceptCors(helloHandler(initFreetypeContext(fontSize))))
	http.Handle("/", r)
	http.ListenAndServe(fmt.Sprintf(":%v", 3111), nil)
}

func corsEnable(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	//fmt.Println("corsEnable")
}

func acceptCors(handlerFunction http.HandlerFunc) http.HandlerFunc {
	return func(response http.ResponseWriter, req *http.Request) {
		log.Printf("Method called:%v\n", req.Method)
		corsEnable(&response)
		if "OPTIONS" == req.Method {
			return
		}
		handlerFunction(response, req)
	}
}

func initFreetypeContext(fontSize int) (*freetype.Context, *image.RGBA) {

	fontfile := "luxisr.ttf"
	fontBytes, err := ioutil.ReadFile(fontfile)
	if err != nil {
		log.Println(err)
		log.Fatalln("Failing out.")
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		log.Fatalln("Failing out.")
	}
	// Initialize the context.
	c := freetype.NewContext()
	c.SetDPI(300)
	c.SetFont(font)
	c.SetFontSize(float64(fontSize))

	rgba := image.NewRGBA(image.Rect(0, 0, 640, 480))
	bg := image.White
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
	c.SetClip(rgba.Bounds())
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)

	fg := image.Black
	c.SetSrc(fg)

	// Draw the guidelines.
	ruler := color.RGBA{0xdd, 0xdd, 0xdd, 0xff}
	for i := 0; i < 200; i++ {
		rgba.Set(10, 10+i, ruler)
		rgba.Set(10+i, 10, ruler)
	}

	return c, rgba
}

func writeText(c *freetype.Context, fontSize int, words string) {

	// Draw the text.
	pt := freetype.Pt(10, 10+int(c.PointToFix32(float64(fontSize))>>8))
	for _, s := range []string{words} {
		_, err := c.DrawString(s, pt)
		if err != nil {
			log.Println(err)
			return
		}
		spacing := 1.5
		pt.Y += c.PointToFix32(float64(fontSize) * spacing)
	}
}

func writeImage(rgba image.Image) {
	// Save that RGBA image to disk.
	f, err := os.Create("out.png")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	b := bufio.NewWriter(f)
	err = png.Encode(b, rgba)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = b.Flush()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	fmt.Println("Wrote out.png OK.")
}

func helloHandler(fc *freetype.Context, rgba *image.RGBA) http.HandlerFunc {
	return func(response http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		words := vars["words"]

		fontSize := 12
		writeText(fc, fontSize, words)

		var buf bytes.Buffer
		buffWriter := &buf
		png.Encode(buffWriter, rgba)
		// err := buffWriter.Flush()
		fmt.Fprintf(response, "%v", buffWriter)

		writeImage(rgba)

		// fmt.Fprintf(response, "Hello!")
		log.Printf("Hello handler!.")
	}
}

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
	"strconv"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"

	"github.com/gorilla/mux"
)

func main() {
	// Open 'http://localhost:3111/calculated-text/abc%5Cndef%20xyz%5Cn123' in your browser.
	r := mux.NewRouter()
	fontSize := 12
	r.HandleFunc("/text/{words}/{height}/{width}", acceptCors(textHandler(initFreetypeContext(fontSize))))
	r.HandleFunc("/text/{words}/{height}", acceptCors(textHandler(initFreetypeContext(fontSize))))
	r.HandleFunc("/text/{words}/", acceptCors(textHandler(initFreetypeContext(fontSize))))
	r.HandleFunc("/text/{words}", acceptCors(textHandler(initFreetypeContext(fontSize))))
	r.HandleFunc("/calculated-text/{words}", acceptCors(calcultedTextHandler(initFreetypeContext(fontSize))))
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

func FontPathToFont(fontfile string) (*truetype.Font, error) {
	fontBytes, err := ioutil.ReadFile(fontfile)
	if err != nil {
		log.Println(err)
		log.Fatalln("Failing out.")
	}
	font, err := freetype.ParseFont(fontBytes)
	return font, err
}

func InitContext(dpi float64, fontSize float64, fontPath string) (*freetype.Context, error) {

	font, error := FontPathToFont(fontPath)
	if error != nil {
		return nil, error
	}

	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(font)
	c.SetFontSize(fontSize)

	return c, nil
}

func CreateRGBA(pointX, pointY, width, height int) *image.RGBA {
	//*image.Rectangle
	//image.Rect(0, 0, 640, 480)
	rect := image.Rect(pointX, pointY, width, height)
	return image.NewRGBA(rect)
}

func DrawToContext(context *freetype.Context, img *image.RGBA, backGround image.Image,
	foreGround image.Image) {
	draw.Draw(img, img.Bounds(), backGround, image.ZP, draw.Src)
	context.SetClip(img.Bounds())
	// context.SetClip(img.Bounds())
	context.SetDst(img)
	context.SetSrc(foreGround)
}

func calcultedTextHandler(fc *freetype.Context, rgba *image.RGBA) http.HandlerFunc {
	return func(response http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		words := vars["words"]

		lines := strings.Split(words, "\\n")
		longestLine := int(0)
		for _, currLine := range lines {
			currLen := len(currLine)
			if currLen > longestLine {
				longestLine = currLen
			}
		}
		linesAmount := len(lines)

		// fontfile := "luxisr.ttf"
		fontfile := "mplus-1m-regular.ttf"

		context, err := InitContext(300, 12, fontfile)
		if err != nil {
			log.Fatalf("Error:%v\n", err)
		}

		width := 28
		// height := 50
		height := 70
		rgba := CreateRGBA(0, 0, longestLine*width, linesAmount*height)
		DrawToContext(context, rgba, image.White, image.Black)
		writeText(context, 12, lines)
		writeImage(rgba)

		bufWriter := GetImageByteBuffer(rgba)
		fmt.Fprintf(response, "%v", bufWriter)

	}
}

func textHandler(fc *freetype.Context, rgba *image.RGBA) http.HandlerFunc {
	return func(response http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		words := vars["words"]

		height, err := strconv.Atoi(vars["height"])
		if err != nil || height <= 0 {
			height = 100
		}

		width, err := strconv.Atoi(vars["width"])
		if err != nil || width <= 0 {
			width = 100
		}

		splitWords := strings.Split(words, " ")
		wordsList := make([]string, 0)
		currentLine := ""
		// for count, curr := range splitWords {
		// if count%2 == 0 && count != 0 {
		for _, curr := range splitWords {
			fmt.Println("Curr:" + curr)
			if curr == "\\n" {
				// currentLine = currentLine + curr + " "
				wordsList = append(wordsList, strings.Trim(currentLine, " "))
				currentLine = ""
			} else {
				currentLine = currentLine + curr + " "
			}
		}
		wordsList = append(wordsList, strings.Trim(currentLine, " "))

		fontfile := "luxisr.ttf"

		context, err := InitContext(300, 12, fontfile)
		if err != nil {
			log.Fatalf("Error:%v\n", err)
		}
		// rgba := CreateRGBA(0, 0, 640, 480)
		rgba := CreateRGBA(0, 0, width, height)
		DrawToContext(context, rgba, image.White, image.Black)
		writeText(context, 12, wordsList)
		writeImage(rgba)

		//Serve image
		bufWriter := GetImageByteBuffer(rgba)
		fmt.Fprintf(response, "%v", bufWriter)
	}
}

// func GetImageByteBuffer(img *image.RGBA) *bytes.Buffer {
func GetImageByteBuffer(img image.Image) *bytes.Buffer {
	var buf bytes.Buffer
	bufWriter := &buf
	png.Encode(bufWriter, img)
	return bufWriter
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
	// c.SetClip(rgba.Bounds())
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

func writeText(c *freetype.Context, fontSize int, words []string) {
	//experimentalY := int(c.PointToFixed(float64(fontSize))) >> 8
	//pt := freetype.Pt(0, experimentalY)

	// Draw the text.
	//fmt.Printf("X:%v, Y:%v\n", 10, 10+int(c.PointToFixed(float64(fontSize))>>8))
	//fmt.Printf("Experimental:X:%v, Y:%v\n", 0, experimentalY)
	pt := freetype.Pt(10, 40+int(c.PointToFixed(float64(fontSize))>>8))

	for _, s := range words {
		_, err := c.DrawString(s, pt)
		if err != nil {
			log.Println(err)
			return
		}
		spacing := 1.5
		pt.Y += c.PointToFixed(float64(fontSize) * spacing)
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

func textHandler_OLD(fc *freetype.Context, rgba *image.RGBA) http.HandlerFunc {
	return func(response http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		words := vars["words"]

		fontSize := 12
		writeText(fc, fontSize, []string{words})

		var buf bytes.Buffer
		buffWriter := &buf
		// c:
		// 	+1 + buffWriter
		png.Encode(buffWriter, rgba)
		// err := buffWriter.Flush()
		fmt.Fprintf(response, "%v", buffWriter)

		writeImage(rgba)

		// fmt.Fprintf(response, "Hello!")
		log.Printf("Hello handler!.")
	}
}

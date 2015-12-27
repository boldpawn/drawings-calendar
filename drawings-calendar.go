package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"strings"
	"log"
	"regexp"
	"sort"
	"image"
	"os"
	"image/png"
	"image/color"
	"golang.org/x/image/math/fixed"
	"time"
	"strconv"
	"path/filepath"
	"image/draw"
	"image/jpeg"
)

type ByDate []string

var dateExpression = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
var nameExpression = regexp.MustCompile(`\d{4}-\d{2}-\d{2}-\s*([^\n\r]*)`)

func (a ByDate) Len() int { return len(a) }
func (a ByDate) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool { return dateExpression.FindString(a[i]) < dateExpression.FindString(a[j]) }

var monthMap = map[time.Month]string{
	time.January : "januari",
	time.February : "februari",
	time.March : "maart",
	time.April : "april",
	time.May : "mei",
	time.June : "juni",
	time.July : "juli",
	time.August : "augustus",
	time.September : "september",
	time.October : "oktober",
	time.November : "november",
	time.December : "december",
}

var dayMap = map[time.Weekday]string{
	time.Sunday : "zondag",
	time.Monday : "maandag",
	time.Tuesday : "dinsdag",
	time.Wednesday : "woensdag",
	time.Thursday : "donderdag",
	time.Friday : "vrijdag",
	time.Saturday : "zaterdag",
}

type DrawingText struct {
	text string
	faceIndex int
}

type DrawingTexts struct {
	dayName   DrawingText
	dayNumber DrawingText
	monthName DrawingText
	year      DrawingText
	name      DrawingText
	website   DrawingText
}

var faces = [8]font.Face{}

var fileSeparator = fmt.Sprintf("%c", filepath.Separator)

func loadFaces(fontName string) {

	s := []string{"/Library/Fonts/", fontName, ".ttf"}
	fontFileName := strings.Join(s, "")

	fmt.Printf("Loading fontfile %q\n", fontFileName)
	b, err := ioutil.ReadFile(fontFileName)
	if err != nil {
		log.Fatalf("Error while loading font file %s. Error = %s", fontFileName, err)
	}
	f, err := truetype.Parse(b)
	if err != nil {
		log.Fatalf("Error while parsing font file %s. Error = %s", fontFileName, err)
	}
	fmt.Printf("Fontfile %q loaded\n", fontFileName)
	faces[0] = truetype.NewFace(f, &truetype.Options{Size: 40})
	faces[1] = truetype.NewFace(f, &truetype.Options{Size: 36})
	faces[2] = truetype.NewFace(f, &truetype.Options{Size: 32})
	faces[3] = truetype.NewFace(f, &truetype.Options{Size: 28})
	faces[4] = truetype.NewFace(f, &truetype.Options{Size: 24})
	faces[5] = truetype.NewFace(f, &truetype.Options{Size: 20})
	faces[6] = truetype.NewFace(f, &truetype.Options{Size: 16})
	faces[7] = truetype.NewFace(f, &truetype.Options{Size: 12})
}

func loadInputImage(fileName string) image.Image {
	reader, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Error while opening file %s. Error = %s", fileName, err)
	}
	defer reader.Close()
	m, err := jpeg.Decode(reader)
	if err != nil {
		log.Fatalf("Error while decoding file %s. Error = %s", fileName, err)
	}
	return m
}

func sortedListOfImages(imageFolder string) []string {
	fmt.Printf("Loading images from %q\n", imageFolder)
	files, err := ioutil.ReadDir(imageFolder)
	if (err != nil) {
		log.Fatalf("Error reading image folder %s. Error = %s", imageFolder, err)
	}
	nrOfImages := 0
	fileNames := make([]string, 0, len(files))
	for _, fileInfo := range files {
		if !fileInfo.IsDir() && strings.HasSuffix(strings.ToLower(fileInfo.Name()), ".jpg") {
			nrOfImages++
			fileNames = fileNames[0:nrOfImages]
			fileNames[nrOfImages - 1] = fileInfo.Name()
		}
	}
	sort.Sort(ByDate(fileNames))

	return fileNames
}

func drawStringInCenter(dest draw.Image, source image.Image, text DrawingText, y int) {
	if &text == nil {
		return
	}

	var drawingOk bool

	for i:= text.faceIndex; i<len(faces) && !drawingOk; i++ {
		drawer := font.Drawer{
			Dst:  dest,
			Src:  source,
			Face: faces[i]}
		width := fixed.I(240)
		stringLength := drawer.MeasureString(text.text)
		startPosition := fixed.Point26_6{X: ((width - stringLength) / 2), Y: fixed.I(y)}
		if startPosition.X > fixed.I(5) {
			drawer.Dot = startPosition
			drawer.DrawString(text.text)
			drawingOk = true
		}
	}

}

func createTextImage(texts DrawingTexts) image.Image {
	m := image.NewRGBA(image.Rect(0, 0, 1920, 1080))
	white := color.RGBA{255, 255, 255, 255}
	draw.Draw(m, m.Bounds(), &image.Uniform{white}, image.ZP, draw.Src)
	source := &image.Uniform{color.RGBA{0, 0, 0, 255}}
	drawStringInCenter(m, source, texts.dayName, 350)
	drawStringInCenter(m, source, texts.dayNumber, 450)
	drawStringInCenter(m, source, texts.monthName, 550)
	drawStringInCenter(m, source, texts.year, 650)
	drawStringInCenter(m, source, texts.name, 50)
	drawStringInCenter(m, source, texts.website, 1060)
	return m
}

func saveImage(img image.Image, destFolder string, fileName string) {
	fullPath := destFolder + fileSeparator + fileName
	outFile, err := os.Create(fullPath)
	if err != nil {
		log.Fatalf("Error while saving image. Error = %s", err)
	}
	defer outFile.Close()

	err = png.Encode(outFile, img)
	if (err != nil) {
		log.Fatalf("Error while writing image. Error = %s", err)
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

func buildTextFromFileName(fileName string) DrawingTexts {
	dateString := dateExpression.FindString(fileName)
	const dateFormat = "2006-01-02"
	t, _ := time.Parse(dateFormat, dateString)
	texts := new(DrawingTexts)
	texts.website = DrawingText{text:"www.piasprong.nl", faceIndex:6}
	texts.year = DrawingText{text:strconv.Itoa(t.Year()), faceIndex:0}
	texts.dayNumber = DrawingText{text:strconv.Itoa(t.Day()), faceIndex:0}
	texts.dayName = DrawingText{text:dayMap[t.Weekday()], faceIndex:0}
	texts.monthName = DrawingText{text:monthMap[t.Month()], faceIndex:0}
	matchparts := nameExpression.FindStringSubmatch(fileName)

	if (len(matchparts) > 1) {
		var extension = filepath.Ext(matchparts[1])
		s := matchparts[1][0:len(matchparts[1]) - len(extension)]
		texts.name = DrawingText{text:s, faceIndex:2}
	}
	return *texts
}

func createCalenderFile(imageFolder string, fileName string) {
	drawingTexts := buildTextFromFileName(fileName)
	m := createTextImage(drawingTexts)
	imageWithDrawing := loadInputImage(imageFolder + fileSeparator + fileName)
	newRect := image.NewRGBA(image.Rect(0, 0, 1920, 1080))
	draw.Draw(newRect, newRect.Bounds(), m, image.Point{X:0, Y:0}, draw.Src)
	draw.Draw(newRect, newRect.Bounds(), imageWithDrawing, image.Point{X:-240, Y:0}, draw.Src)
	saveImage(newRect, imageFolder + fileSeparator + "out", fileName)
	fmt.Printf("Image %s saved\n", fileName)
}

func main() {

	fontName := flag.String("font", "Verdana", "font name [Verdana]")
	imageFolder := flag.String("folder", ".", "source folder [.]")
	flag.Parse()

	fmt.Printf("font name = %s\n", *fontName)
	fmt.Printf("folder = %s\n", *imageFolder)

	loadFaces(*fontName)
	fileNames := sortedListOfImages(*imageFolder)

	outputFolder := *imageFolder + fileSeparator + "out"
	fileExists, _ := exists(outputFolder)
	if fileExists {
		os.RemoveAll(*imageFolder + string(fileSeparator) + "out")
	}

	os.Mkdir(*imageFolder + string(fileSeparator) + "out", 0755)

	for _, fileName := range fileNames {
		createCalenderFile(*imageFolder, fileName)
	}
}

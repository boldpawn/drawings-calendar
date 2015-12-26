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
)

type ByDate []string

var dateExpression = regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
var nameExpression = regexp.MustCompile(`\w+\.`)

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

type DrawingTexts struct {
	dayName   string
	dayNumber string
	monthName string
	year      string
	name      string
	website   string
}

var fileSeparator = fmt.Sprintf("%c", filepath.Separator)

func loadFace(fontName string) (font.Face) {

	fontOptions := truetype.Options{Size: 32}
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
	return truetype.NewFace(f, &fontOptions)
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

func drawStringInCenter(d font.Drawer, s string, y int) {
	width := fixed.I(240)
	stringLength := d.MeasureString(s)
	startPosition := fixed.Point26_6{X: ((width - stringLength) / 2), Y: fixed.I(y)}
	d.Dot = startPosition
	d.DrawString(s)
}

func createTextImage(face font.Face, texts DrawingTexts) image.Image {
	m := image.NewRGBA(image.Rect(0, 0, 1920, 1080))
	source := &image.Uniform{color.RGBA{0, 0, 0, 255}}
	drawer := font.Drawer{
		Dst:  m,
		Src:  source,
		Face: face}
	drawStringInCenter(drawer, texts.dayName, 200)
	drawStringInCenter(drawer, texts.dayNumber, 300)
	drawStringInCenter(drawer, texts.monthName, 400)
	drawStringInCenter(drawer, texts.year, 500)
	drawStringInCenter(drawer, texts.name, 800)
	drawStringInCenter(drawer, texts.website, 900)
	return m
}

func saveImage(img image.Image, destFolder string, fileName string) {
	fullPath := destFolder + fileSeparator + fileName
	outFile, err := os.Create(fullPath)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer outFile.Close()

	err = png.Encode(outFile, img)
	if (err != nil) {
		log.Fatalf("Error while writing image %s. Error = %s", fullPath, err)
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
	texts.website = "www.piasprong.nl"
	texts.year = strconv.Itoa(t.Year())
	texts.dayNumber = strconv.Itoa(t.Day())
	texts.dayName = dayMap[t.Weekday()]
	texts.monthName = monthMap[t.Month()]
	texts.name = nameExpression.FindString(fileName)
	return *texts
}

func main() {

	fontName := flag.String("font", "Verdana", "font name [Verdana]")
	imageFolder := flag.String("folder", ".", "source folder [.]")
	flag.Parse()

	fmt.Printf("font name = %s\n", *fontName)
	fmt.Printf("folder = %s\n", *imageFolder)

	face := loadFace(*fontName)
	fileNames := sortedListOfImages(*imageFolder)

	outputFolder := *imageFolder + fileSeparator + "out"
	fileExists, _ := exists(outputFolder)
	if fileExists {
		os.RemoveAll(*imageFolder + string(fileSeparator) + "out")
	}

	os.Mkdir(*imageFolder + string(fileSeparator) + "out", 0755)

	for _, fileName := range fileNames {
		drawingTexts := buildTextFromFileName(fileName)
		m := createTextImage(face, drawingTexts)
		saveImage(m, *imageFolder + fileSeparator + "out", fileName)
		fmt.Printf("Image %s saved\n", fileName)
	}
}

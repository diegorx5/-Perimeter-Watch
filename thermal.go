package detector

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"
)

type Detection struct {
	CenterX     int
	CenterY     int
	Width       int
	Height      int
	Temperature float64
}

type ThermalDetector struct {
	HeatThreshold uint8
	MinBlobSize   int
	FrameWidth    int
	FrameHeight   int
	lastHeatMap   [][]bool
}

func NewThermalDetector() *ThermalDetector {
	return &ThermalDetector{
		HeatThreshold: 200,
		MinBlobSize:   300,
		FrameWidth:    640,
		FrameHeight:   480,
	}
}

func (d *ThermalDetector) AnalyzeFrame(imagePath string) ([]Detection, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, fmt.Errorf("error abriendo imagen: %w", err)
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("error decodificando imagen: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	currentHeatMap := make([][]bool, height)
	for y := 0; y < height; y++ {
		currentHeatMap[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			gray := uint8((0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 256)
			currentHeatMap[y][x] = gray > d.HeatThreshold
		}
	}

	if d.lastHeatMap == nil {
		d.lastHeatMap = currentHeatMap
		fmt.Println("Primer frame - calibrando detector...")
		return []Detection{}, nil
	}

	movementMap := make([][]bool, height)
	for y := 0; y < height; y++ {
		movementMap[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			if currentHeatMap[y][x] != d.lastHeatMap[y][x] {
				movementMap[y][x] = true
			}
		}
	}

	d.lastHeatMap = currentHeatMap
	detections := d.findBlobs(movementMap, width, height)

	fmt.Printf("Analizando frame %dx%d - %d firmas en movimiento\n",
		width, height, len(detections))

	return detections, nil
}

func (d *ThermalDetector) findBlobs(heatMap [][]bool, width, height int) []Detection {
	visited := make([][]bool, height)
	for y := range visited {
		visited[y] = make([]bool, width)
	}

	marginX := width / 10
	marginY := height / 10

	var detections []Detection

	for y := marginY; y < height-marginY; y++ {
		for x := marginX; x < width-marginX; x++ {
			if heatMap[y][x] && !visited[y][x] {
				blob := d.expandBlob(heatMap, visited, x, y, width, height)
				if len(blob) >= d.MinBlobSize {
					det := d.blobToDetection(blob)
					detections = append(detections, det)
				}
			}
		}
	}

	return detections
}

type point struct{ x, y int }

func (d *ThermalDetector) expandBlob(heatMap [][]bool, visited [][]bool,
	startX, startY, width, height int) []point {

	var blob []point
	queue := []point{{startX, startY}}
	visited[startY][startX] = true
	dirs := []point{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		blob = append(blob, curr)

		for _, dir := range dirs {
			nx, ny := curr.x+dir.x, curr.y+dir.y
			if nx >= 0 && nx < width && ny >= 0 && ny < height &&
				!visited[ny][nx] && heatMap[ny][nx] {
				visited[ny][nx] = true
				queue = append(queue, point{nx, ny})
			}
		}
	}

	return blob
}

func (d *ThermalDetector) blobToDetection(blob []point) Detection {
	minX, minY := math.MaxInt32, math.MaxInt32
	maxX, maxY := 0, 0
	sumX, sumY := 0, 0

	for _, p := range blob {
		if p.x < minX {
			minX = p.x
		}
		if p.x > maxX {
			maxX = p.x
		}
		if p.y < minY {
			minY = p.y
		}
		if p.y > maxY {
			maxY = p.y
		}
		sumX += p.x
		sumY += p.y
	}

	centerX := sumX / len(blob)
	centerY := sumY / len(blob)
	tempEstimate := 35.0 + float64(len(blob))*0.01

	return Detection{
		CenterX:     centerX,
		CenterY:     centerY,
		Width:       maxX - minX,
		Height:      maxY - minY,
		Temperature: tempEstimate,
	}
}

func (d *ThermalDetector) DrawDetections(inputPath, outputPath string,
	detections []Detection) error {

	file, err := os.Open(inputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		return err
	}

	rgba := image.NewRGBA(img.Bounds())
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	for _, det := range detections {
		x0 := det.CenterX - det.Width/2
		y0 := det.CenterY - det.Height/2
		x1 := det.CenterX + det.Width/2
		y1 := det.CenterY + det.Height/2

		for x := x0; x <= x1; x++ {
			rgba.Set(x, y0, red)
			rgba.Set(x, y1, red)
		}
		for y := y0; y <= y1; y++ {
			rgba.Set(x0, y, red)
			rgba.Set(x1, y, red)
		}

		fmt.Printf("Firma en movimiento: centro(%d,%d) tamano(%dx%d) temp=%.1fC\n",
			det.CenterX, det.CenterY, det.Width, det.Height, det.Temperature)
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	return jpeg.Encode(out, rgba, &jpeg.Options{Quality: 90})
}

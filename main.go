package main

import (
	"fmt"
	"flir/camera"
	"flir/dashboard"
	"flir/detector"
	"log"
	"time"
)

const (
	CAMERA_IP   = "192.168.40.84"
	CAMERA_USER = "admin"
	CAMERA_PASS = "admin"
)

func clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func main() {
	fmt.Println("Iniciando Perimeter Watch...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	dashboard.Start()

	ptzCtrl, err := camera.NewPTZController(CAMERA_IP, CAMERA_USER, CAMERA_PASS)
	if err != nil {
		log.Fatal("Error conectando PTZ:", err)
	}

	stream := camera.NewStreamCapture(CAMERA_IP, CAMERA_USER, CAMERA_PASS)
	det := detector.NewThermalDetector()

	dashboard.State.SetStatus("Moviendo a posicion home...")
	fmt.Println("Moviendo a posicion home...")
	err = ptzCtrl.GoHome()
	if err != nil {
		log.Println("Error yendo a home:", err)
	}
	time.Sleep(1 * time.Second)

	for i := 1; ; i++ {
		dashboard.State.CycleCount = i
		fmt.Printf("\nCiclo %d\n", i)
		dashboard.State.SetStatus(fmt.Sprintf("Patrullando - ciclo %d", i))

		err = stream.CaptureSnapshot("thermal_raw.jpg")
		if err != nil {
			log.Println("Error capturando frame:", err)
			continue
		}

		detections, err := det.AnalyzeFrame("thermal_raw.jpg")
		if err != nil {
			log.Println("Error analizando frame:", err)
			continue
		}

		err = det.DrawDetections("thermal_raw.jpg", "thermal_detected.jpg", detections)
		if err != nil {
			log.Println("Error dibujando detecciones:", err)
			continue
		}

		dashboard.State.UpdateImage("thermal_detected.jpg")

		var validDetections []detector.Detection
		for _, d := range detections {
			if d.CenterY > 50 && d.CenterY < 430 &&
				d.CenterX > 80 && d.CenterX < 560 &&
				d.Width > 20 && d.Height > 20 {
				validDetections = append(validDetections, d)
			}
		}

		if len(validDetections) > 0 {
			fmt.Println("ALERTA - FIRMA TERMICA DETECTADA")
			biggest := validDetections[0]

			offsetX := float64(biggest.CenterX-320) / 320.0
			offsetY := float64(240-biggest.CenterY) / 240.0
			speedX := clamp(offsetX*0.2, -0.15, 0.15)
			speedY := clamp(offsetY*0.2, -0.15, 0.15)

			fmt.Printf("Objetivo: centro(%d,%d) temp=%.1fC\n",
				biggest.CenterX, biggest.CenterY, biggest.Temperature)
			fmt.Printf("PTZ X:%.2f Y:%.2f\n", speedX, speedY)

			dashboard.State.AddAlert(dashboard.AlertEvent{
				Timestamp:   dashboard.Timestamp(),
				CenterX:     biggest.CenterX,
				CenterY:     biggest.CenterY,
				Width:       biggest.Width,
				Height:      biggest.Height,
				Temperature: biggest.Temperature,
				SpeedX:      speedX,
				SpeedY:      speedY,
			})
			dashboard.State.SetStatus("FIRMA TERMICA DETECTADA - PTZ TRACKING")

			err = ptzCtrl.MoveRelative(speedX, speedY)
			if err != nil {
				log.Println("Error moviendo PTZ:", err)
			}
			time.Sleep(300 * time.Millisecond)
			err = ptzCtrl.Stop()
			if err != nil {
				log.Println("Error deteniendo PTZ:", err)
			}

		} else {
			fmt.Println("Zona despejada")
			dashboard.State.SetStatus("Zona despejada - patrullando")
		}

		time.Sleep(500 * time.Millisecond)
	}
}

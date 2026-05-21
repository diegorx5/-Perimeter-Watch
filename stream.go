package camera

import (
	"fmt"
	"os/exec"
)

type StreamCapture struct {
	IP       string
	User     string
	Password string
}

func NewStreamCapture(ip, user, password string) *StreamCapture {
	return &StreamCapture{IP: ip, User: user, Password: password}
}

func (s *StreamCapture) ThermalURL() string {
	return fmt.Sprintf("rtsp://%s:%s@%s/ch1", s.User, s.Password, s.IP)
}

func (s *StreamCapture) VisibleURL() string {
	return fmt.Sprintf("rtsp://%s:%s@%s/ch2", s.User, s.Password, s.IP)
}

func (s *StreamCapture) CaptureSnapshot(outputPath string) error {
	cmd := exec.Command("ffmpeg",
		"-rtsp_transport", "tcp",
		"-i", s.ThermalURL(),
		"-vframes", "1",
		"-q:v", "2",
		"-y",
		outputPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error capturando frame: %w\n%s", err, output)
	}
	fmt.Println("Snapshot guardado en:", outputPath)
	return nil
}

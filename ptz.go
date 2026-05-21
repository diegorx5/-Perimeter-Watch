package camera

import (
	"fmt"
	"github.com/use-go/onvif"
	"github.com/use-go/onvif/ptz"
	onvifxsd "github.com/use-go/onvif/xsd/onvif"
	"github.com/use-go/onvif/xsd"
)

type PTZController struct {
	device       *onvif.Device
	profileToken onvifxsd.ReferenceToken
}

func NewPTZController(ip, user, password string) (*PTZController, error) {
	dev, err := onvif.NewDevice(onvif.DeviceParams{
		Xaddr:    ip + ":80",
		Username: user,
		Password: password,
	})
	if err != nil {
		return nil, fmt.Errorf("error conectando a la camara: %w", err)
	}
	fmt.Println("Conectado a FLIR PT-Series:", ip)
	return &PTZController{
		device:       dev,
		profileToken: "MP1",
	}, nil
}

func (p *PTZController) MoveRelative(panSpeed, tiltSpeed float64) error {
	req := ptz.ContinuousMove{
		ProfileToken: p.profileToken,
		Velocity: onvifxsd.PTZSpeed{
			PanTilt: onvifxsd.Vector2D{
				X:     panSpeed,
				Y:     tiltSpeed,
				Space: "http://www.onvif.org/ver10/tptz/PanTiltSpaces/VelocityGenericSpace",
			},
		},
	}
	_, err := p.device.CallMethod(req)
	if err != nil {
		return fmt.Errorf("error moviendo PTZ: %w", err)
	}
	return nil
}

func (p *PTZController) Stop() error {
	req := ptz.Stop{
		ProfileToken: "MP1",
		PanTilt:      xsd.Boolean(true),
		Zoom:         xsd.Boolean(false),
	}
	_, err := p.device.CallMethod(req)
	if err != nil {
		return fmt.Errorf("error deteniendo PTZ: %w", err)
	}
	return nil
}

func (p *PTZController) GoToPreset(presetToken string) error {
	req := ptz.GotoPreset{
		ProfileToken: "MP1",
		PresetToken:  onvifxsd.ReferenceToken(presetToken),
	}
	_, err := p.device.CallMethod(req)
	if err != nil {
		return fmt.Errorf("error yendo a preset: %w", err)
	}
	return nil
}

func (p *PTZController) GoHome() error {
	req := ptz.GotoHomePosition{
		ProfileToken: "MP1",
	}
	_, err := p.device.CallMethod(req)
	if err != nil {
		return fmt.Errorf("error yendo a home: %w", err)
	}
	return nil
}

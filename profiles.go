package camera

import (
	"encoding/xml"
	"fmt"
	"github.com/use-go/onvif"
	"github.com/use-go/onvif/media"
)

type Profile struct {
	Token string
	Name  string
}

func GetProfiles(ip, user, password string) ([]Profile, error) {
	dev, err := onvif.NewDevice(onvif.DeviceParams{
		Xaddr:    ip + ":80",
		Username: user,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	resp, err := dev.CallMethod(media.GetProfiles{})
	if err != nil {
		return nil, fmt.Errorf("error obteniendo perfiles: %w", err)
	}

	buf := new(struct {
		Body struct {
			Inner []byte `xml:",innerxml"`
		}
	})

	decoder := xml.NewDecoder(resp.Body)
	err = decoder.Decode(buf)
	if err != nil {
		return nil, err
	}

	fmt.Println("Perfiles ONVIF:")
	fmt.Println(string(buf.Body.Inner))
	return nil, nil
}

package service

import (
	"encoding/json"

	"github.com/gotosky/gotosky/internal/domain"
	"github.com/gotosky/gotosky/internal/store"
)

type otaSpec struct {
	FocalMM    float64 `json:"focal_mm"`
	ApertureMM float64 `json:"aperture_mm"`
}

type camSpec struct {
	PixUM float64 `json:"pix_um"`
	W     float64 `json:"w"`
	H     float64 `json:"h"`
}

func AnnotateFOV(rigs []domain.Rig, eqs []domain.Equipment) []domain.Rig {
	byID := map[string]domain.Equipment{}
	for _, e := range eqs {
		byID[e.ID.String()] = e
	}
	for i := range rigs {
		var ota otaSpec
		var cam camSpec
		for _, c := range rigs[i].Components {
			e := byID[c.EquipmentID.String()]
			switch e.Kind {
			case "OTA":
				_ = json.Unmarshal(e.Specs, &ota)
			case "CAMERA":
				_ = json.Unmarshal(e.Specs, &cam)
			}
		}
		if ota.FocalMM > 0 && cam.PixUM > 0 && cam.W > 0 {
			rigs[i].ScalePPS = store.PlateScaleArcsec(cam.PixUM, ota.FocalMM)
			rigs[i].FOVArcmin = store.FOVArcmin(cam.PixUM, cam.W, ota.FocalMM)
		}
	}
	return rigs
}

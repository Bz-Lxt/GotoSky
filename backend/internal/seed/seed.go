package seed

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/gotosky/gotosky/internal/domain"
	"github.com/gotosky/gotosky/internal/engine"
	"github.com/gotosky/gotosky/internal/store"
	"golang.org/x/crypto/bcrypt"
)

var (
	DemoUser = uuid.MustParse("10000000-0000-4000-8000-000000000001")
	DemoSite = uuid.MustParse("20000000-0000-4000-8000-000000000001")
	DemoRig  = uuid.MustParse("30000000-0000-4000-8000-000000000001")
	DemoPlan = uuid.MustParse("40000000-0000-4000-8000-000000000001")
)

func Run(ctx context.Context, st *store.Store) error {
	hash, _ := bcrypt.GenerateFromPassword([]byte("gotosky"), bcrypt.DefaultCost)
	_ = st.CreateUser(ctx, domain.User{ID: DemoUser, Username: "skye", PasswordHash: string(hash), DisplayName: "打星人"})

	p := engine.DefaultProfile()
	w, _ := json.Marshal(map[string]float64{"C": p.WC, "S": p.WS, "M": p.WM, "A": p.WA, "T": p.WT, "L": p.WL, "N": p.WN})
	see, _ := json.Marshal(map[string]float64{"bias": p.SeeingBias, "v250": p.SeeingV250, "shear": p.SeeingShear, "v10": p.SeeingV10})
	if err := st.EnsureProfile(ctx, p.ID, "default", engine.EngineVersion, w, see); err != nil {
		return err
	}

	n, _ := st.CountTargets(ctx)
	if n < 120 {
		for _, t := range Catalog() {
			if err := st.UpsertTarget(ctx, t); err != nil {
				return err
			}
		}
	}

	site := domain.Site{
		ID: DemoSite, Name: "兴隆观测站", Latitude: 40.45, Longitude: 116.02, ElevationM: 960,
		Timezone: "Asia/Shanghai", Bortle: 4, SQM: engine.BortleToSQM(4), MinAltitude: 20,
		HorizonMask: domain.EmptyJSONArray(),
	}
	if err := st.UpsertSite(ctx, site); err != nil {
		return err
	}

	eqs := []domain.Equipment{
		{ID: uuid.MustParse("31000000-0000-4000-8000-000000000001"), Name: "EQ6-R Pro", Kind: "MOUNT", Specs: []byte(`{"payload_kg":20}`)},
		{ID: uuid.MustParse("31000000-0000-4000-8000-000000000002"), Name: "RedCat 51", Kind: "OTA", Specs: []byte(`{"focal_mm":250,"aperture_mm":51}`)},
		{ID: uuid.MustParse("31000000-0000-4000-8000-000000000003"), Name: "ASI2600MC", Kind: "CAMERA", Specs: []byte(`{"pix_um":3.76,"w":6248,"h":4176}`)},
		{ID: uuid.MustParse("31000000-0000-4000-8000-000000000004"), Name: "EFW 7x36", Kind: "FILTER_WHEEL", Specs: []byte(`{"slots":7}`)},
		{ID: uuid.MustParse("31000000-0000-4000-8000-000000000005"), Name: "30F4 Guide", Kind: "GUIDE_SCOPE", Specs: []byte(`{"focal_mm":120}`)},
		{ID: uuid.MustParse("31000000-0000-4000-8000-000000000006"), Name: "ASI120MM", Kind: "GUIDE_CAMERA", Specs: []byte(`{"pix_um":3.75}`)},
	}
	for _, e := range eqs {
		if err := st.UpsertEquipment(ctx, e); err != nil {
			return err
		}
	}
	rig := domain.Rig{ID: DemoRig, Name: "Widefield A", Notes: "RedCat + 2600MC", Components: []domain.RigComp{
		{EquipmentID: eqs[0].ID, Role: "MOUNT"},
		{EquipmentID: eqs[1].ID, Role: "OTA"},
		{EquipmentID: eqs[2].ID, Role: "CAMERA"},
		{EquipmentID: eqs[3].ID, Role: "FILTER_WHEEL"},
		{EquipmentID: eqs[4].ID, Role: "GUIDE_SCOPE"},
		{EquipmentID: eqs[5].ID, Role: "GUIDE_CAMERA"},
	}}
	if err := st.UpsertRig(ctx, rig); err != nil {
		return err
	}
	_ = st.UpsertPlan(ctx, domain.Plan{ID: DemoPlan, SiteID: DemoSite, Title: "本周兴隆深空", Notes: "优先 M31 / 北美"})
	return nil
}

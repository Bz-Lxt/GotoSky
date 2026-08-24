package telescope

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AlpacaDriver is the real HTTP path (UNVERIFIED without hardware).
type AlpacaDriver struct {
	Base   string
	Client *http.Client
	Dev    int
}

func NewAlpaca(base string) *AlpacaDriver {
	return &AlpacaDriver{Base: base, Client: &http.Client{Timeout: 8 * time.Second}, Dev: 0}
}

func (a *AlpacaDriver) Name() string   { return "alpaca" }
func (a *AlpacaDriver) Source() string { return "DEVICE" }
func (a *AlpacaDriver) Inject(_ string) {}

type alpacaResp struct {
	Value        any    `json:"Value"`
	ErrorNumber  int    `json:"ErrorNumber"`
	ErrorMessage string `json:"ErrorMessage"`
}

func (a *AlpacaDriver) put(ctx context.Context, path string, body map[string]any) error {
	if body == nil {
		body = map[string]any{}
	}
	body["ClientID"] = 1
	body["ClientTransactionID"] = time.Now().UnixNano() % 1e9
	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, a.Base+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.Client.Do(req)
	if err != nil {
		return &DeviceError{Code: "DeviceTimeout", Class: ClassTransient, Msg: err.Error()}
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return &DeviceError{Code: "InvalidCommand", Class: ClassPermanent, Msg: fmt.Sprintf("http %d", resp.StatusCode)}
	}
	var ar alpacaResp
	_ = json.Unmarshal(raw, &ar)
	if ar.ErrorNumber != 0 {
		return &DeviceError{Code: "HardwareFault", Class: ClassPermanent, Msg: ar.ErrorMessage}
	}
	return nil
}

func (a *AlpacaDriver) Connect(ctx context.Context) error {
	return a.put(ctx, fmt.Sprintf("/api/v1/telescope/%d/connected", a.Dev), map[string]any{"Connected": true})
}
func (a *AlpacaDriver) Disconnect(ctx context.Context) error {
	return a.put(ctx, fmt.Sprintf("/api/v1/telescope/%d/connected", a.Dev), map[string]any{"Connected": false})
}
func (a *AlpacaDriver) Slew(ctx context.Context, ra, dec float64) error {
	return a.put(ctx, fmt.Sprintf("/api/v1/telescope/%d/slewtocoordinates", a.Dev), map[string]any{
		"RightAscension": ra, "Declination": dec,
	})
}
func (a *AlpacaDriver) WaitSlew(ctx context.Context) error { return a.poll(ctx) }
func (a *AlpacaDriver) Settle(ctx context.Context) error   { return sleepCtx(ctx, 200*time.Millisecond) }
func (a *AlpacaDriver) LockGuide(ctx context.Context) error {
	return a.put(ctx, "/api/v1/guider/0/startguiding", map[string]any{})
}
func (a *AlpacaDriver) SetFilter(ctx context.Context, pos int) error {
	return a.put(ctx, "/api/v1/filterwheel/0/position", map[string]any{"Position": pos})
}
func (a *AlpacaDriver) Expose(ctx context.Context, seconds float64) error {
	return a.put(ctx, "/api/v1/camera/0/startexposure", map[string]any{"Duration": seconds, "Light": true})
}
func (a *AlpacaDriver) Dither(ctx context.Context) error { return sleepCtx(ctx, 200*time.Millisecond) }
func (a *AlpacaDriver) Park(ctx context.Context) error {
	return a.put(ctx, fmt.Sprintf("/api/v1/telescope/%d/park", a.Dev), map[string]any{})
}
func (a *AlpacaDriver) Heartbeat(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/telescope/%d/connected", a.Base, a.Dev), nil)
	if err != nil {
		return err
	}
	resp, err := a.Client.Do(req)
	if err != nil {
		return &DeviceError{Code: "ConnectionReset", Class: ClassTransient, Msg: err.Error()}
	}
	resp.Body.Close()
	return nil
}
func (a *AlpacaDriver) ReadSensors(_ context.Context) (Sensors, error) {
	return Sensors{Source: a.Source()}, nil
}
func (a *AlpacaDriver) poll(ctx context.Context) error { return sleepCtx(ctx, 100*time.Millisecond) }

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

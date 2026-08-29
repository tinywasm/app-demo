package devices

import (
	"github.com/tinywasm/input"
	"github.com/tinywasm/model"
	"github.com/tinywasm/view"
)

// deviceModel is a model built from real widgets (input.Text() is a model.Kind).
// All three fields are NotNull (a device with a blank id/name/ip isn't a
// valid record) and ip uses the dedicated input.IP() widget — not
// input.Text() — so a malformed address fails Form.Validate() instead of
// silently persisting.
var deviceDef = model.Definition{
	Name: "device",
	Fields: model.Fields{
		{Name: "id", Type: input.Text(), NotNull: true, DB: &model.FieldDB{PK: true}},
		{Name: "name", Type: input.Text(), NotNull: true},
		{Name: "ip", Type: input.IP(), NotNull: true},
	},
}

type Device struct {
	Id, Name, Ip string
}

func (d *Device) ModelName() string     { return "device" }
func (d *Device) Schema() []model.Field { return deviceDef.Fields }
func (d *Device) Pointers() []any       { return []any{&d.Id, &d.Name, &d.Ip} }
func (d *Device) IsNil() bool           { return d == nil }

func (d *Device) EncodeFields(w model.FieldWriter) {
	w.String("id", d.Id)
	w.String("name", d.Name)
	w.String("ip", d.Ip)
}

func (d *Device) DecodeFields(r model.FieldReader) {
	d.Id, _ = r.String("id")
	d.Name, _ = r.String("name")
	d.Ip, _ = r.String("ip")
}

func (d *Device) Item() view.Item {
	return view.Item{ID: d.Id, Label: d.Name, Description: d.Ip}
}

var _ model.Model = (*Device)(nil) // Verify compile-time implementation
var _ view.Itemizer = (*Device)(nil)

type deviceList struct {
	Items []*Device
}

func (l *deviceList) IsNil() bool                      { return l == nil }
func (l *deviceList) DecodeFields(r model.FieldReader) {}
func (l *deviceList) Schema() []model.Field            { return nil }
func (l *deviceList) Pointers() []any                  { return nil }
func (l *deviceList) Len() int                         { return len(l.Items) }
func (l *deviceList) At(i int) model.Fielder           { return l.Items[i] }
func (l *deviceList) Append() model.Fielder {
	d := &Device{}
	l.Items = append(l.Items, d)
	return d
}

var _ model.ModelSlice = (*deviceList)(nil)

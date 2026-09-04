package reservation

import (
	"github.com/tinywasm/input"
	"github.com/tinywasm/model"
	"github.com/tinywasm/view"
)

var reservationDef = model.Definition{
	Name: "reservation",
	Fields: model.Fields{
		{Name: "id", Type: input.Text(), NotNull: true, DB: &model.FieldDB{PK: true}},
		{Name: "patient_run", Type: input.Text(), NotNull: true},
		{Name: "patient_name", Type: input.Text(), NotNull: true},
		{Name: "patient_birthday", Type: input.Text()},
		{Name: "patient_contact", Type: input.Text()},
		{Name: "day", Type: input.Text(), NotNull: true},
		{Name: "hour", Type: input.Text(), NotNull: true},
		{Name: "detail", Type: input.Text()},
		{Name: "status", Type: input.Text()},
	},
}

type Reservation struct {
	Id, PatientRun, PatientName, PatientBirthday, PatientContact string
	Day, Hour, Detail, Status                                    string
}

func (r *Reservation) ModelName() string     { return "reservation" }
func (r *Reservation) Schema() []model.Field { return reservationDef.Fields }
func (r *Reservation) Pointers() []any {
	return []any{
		&r.Id, &r.PatientRun, &r.PatientName, &r.PatientBirthday, &r.PatientContact,
		&r.Day, &r.Hour, &r.Detail, &r.Status,
	}
}
func (r *Reservation) IsNil() bool { return r == nil }

func (r *Reservation) EncodeFields(w model.FieldWriter) {
	w.String("id", r.Id)
	w.String("patient_run", r.PatientRun)
	w.String("patient_name", r.PatientName)
	w.String("patient_birthday", r.PatientBirthday)
	w.String("patient_contact", r.PatientContact)
	w.String("day", r.Day)
	w.String("hour", r.Hour)
	w.String("detail", r.Detail)
	w.String("status", r.Status)
}

func (r *Reservation) DecodeFields(fr model.FieldReader) {
	r.Id, _ = fr.String("id")
	r.PatientRun, _ = fr.String("patient_run")
	r.PatientName, _ = fr.String("patient_name")
	r.PatientBirthday, _ = fr.String("patient_birthday")
	r.PatientContact, _ = fr.String("patient_contact")
	r.Day, _ = fr.String("day")
	r.Hour, _ = fr.String("hour")
	r.Detail, _ = fr.String("detail")
	r.Status, _ = fr.String("status")
}

func (r *Reservation) Item() view.Item {
	return view.Item{
		ID:          r.Id,
		LeadMain:    r.Hour,
		Label:       r.PatientName,
		Description: statusLabel(r.Status),
	}
}

func statusLabel(s string) string {
	switch s {
	case "confirmed":
		return "Confirmada"
	case "attended":
		return "Atendida"
	}
	return ""
}

var _ model.Model = (*Reservation)(nil)
var _ view.Itemizer = (*Reservation)(nil)

package medicalhistory

import (
	"github.com/tinywasm/input"
	"github.com/tinywasm/model"
	"github.com/tinywasm/time"
	"github.com/tinywasm/view"
)

// visitDef mirrors devices' deviceDef exactly: a model built from real
// widgets. Patient is a plain field, same tier as the other three — the
// picker above narrows the list, it does not own the record.
var visitDef = model.Definition{
	Name: "visit",
	Fields: model.Fields{
		{Name: "id", Type: input.Text(), NotNull: true, DB: &model.FieldDB{PK: true}},
		{Name: "patient", Type: input.Text(), NotNull: true},
		{Name: "doctor", Type: input.Text(), NotNull: true},
		{Name: "date", Type: input.Text(), NotNull: true},
		{Name: "reason", Type: input.Text(), NotNull: true},
		{Name: "diagnosis", Type: input.Text()},
	},
}

type Visit struct {
	Id, Patient, Doctor, Date, Reason, Diagnosis string
}

func (v *Visit) ModelName() string     { return "visit" }
func (v *Visit) Schema() []model.Field { return visitDef.Fields }
func (v *Visit) Pointers() []any {
	return []any{&v.Id, &v.Patient, &v.Doctor, &v.Date, &v.Reason, &v.Diagnosis}
}
func (v *Visit) IsNil() bool { return v == nil }

func (v *Visit) EncodeFields(w model.FieldWriter) {
	w.String("id", v.Id)
	w.String("patient", v.Patient)
	w.String("doctor", v.Doctor)
	w.String("date", v.Date)
	w.String("reason", v.Reason)
	w.String("diagnosis", v.Diagnosis)
}

func (v *Visit) DecodeFields(r model.FieldReader) {
	v.Id, _ = r.String("id")
	v.Patient, _ = r.String("patient")
	v.Doctor, _ = r.String("doctor")
	v.Date, _ = r.String("date")
	v.Reason, _ = r.String("reason")
	v.Diagnosis, _ = r.String("diagnosis")
}

// weekdayAbbr/monthAbbr back leadFromDate's LeadTop/LeadBottom — Spanish,
// matching this module's other user-facing strings.
var weekdayAbbr = [7]string{"Dom", "Lun", "Mar", "Mié", "Jue", "Vie", "Sáb"} // time.Weekday: 0=Sunday
var monthAbbr = [12]string{"Ene", "Feb", "Mar", "Abr", "May", "Jun", "Jul", "Ago", "Sep", "Oct", "Nov", "Dic"}

// leadFromDate turns a "YYYY-MM-DD" date into the three-line badge
// medicalhistory.go's targethour factory reads instead of a plain label —
// e.g. "Vie" / "20" / "Jul 26" (month + the year's last 2 digits — the
// reference this badge follows pairs the month with a short year, not the
// day again). A date that fails to parse degrades to an empty badge rather
// than a panic: Item() runs over whatever the store holds, and a malformed
// date must not take the whole list down with it.
func leadFromDate(dateStr string) (top, main, bottom string) {
	if len(dateStr) != len("YYYY-MM-DD") {
		return "", "", ""
	}
	nano, err := time.ParseDate(dateStr)
	if err != nil {
		return "", "", ""
	}
	weekday := time.Weekday(nano / 1e9)
	day := dateStr[8:10]
	if day[0] == '0' {
		day = day[1:]
	}
	month := dateStr[5:7]
	monthIdx := int(month[0]-'0')*10 + int(month[1]-'0') - 1
	if monthIdx < 0 || monthIdx > 11 {
		return "", "", ""
	}
	yy := dateStr[2:4]
	return weekdayAbbr[weekday], day, monthAbbr[monthIdx] + " " + yy
}

// Item projects a visit as a list row: the leading badge carries the date
// (see leadFromDate), Label the reason, Description the attending doctor —
// the list is always reached by picking a patient first (see
// medicalhistory.go's requirePatient), so every visible row already belongs
// to that one patient; what a row's own badge still needs to say is WHO saw
// them, not whose visit it was.
func (v *Visit) Item() view.Item {
	top, main, bottom := leadFromDate(v.Date)
	return view.Item{
		ID: v.Id, Label: v.Reason, Description: v.Doctor,
		LeadTop: top, LeadMain: main, LeadBottom: bottom,
	}
}

var _ model.Model = (*Visit)(nil)
var _ view.Itemizer = (*Visit)(nil)

type visitList struct {
	Items []*Visit
}

func (l *visitList) IsNil() bool                      { return l == nil }
func (l *visitList) DecodeFields(r model.FieldReader) {}
func (l *visitList) Schema() []model.Field            { return nil }
func (l *visitList) Pointers() []any                  { return nil }
func (l *visitList) Len() int                         { return len(l.Items) }
func (l *visitList) At(i int) model.Fielder           { return l.Items[i] }
func (l *visitList) Append() model.Fielder {
	v := &Visit{}
	l.Items = append(l.Items, v)
	return v
}

var _ model.ModelSlice = (*visitList)(nil)

package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// TODO: removal provision
type SkipAddressable interface {
	Str() string
	Toggle()
	// selfID has been nuked, use type assertions instead
	GetModifiable() *[]NamedReceiver
	Next() (bool, SkipAddressable)
	Prev() (bool, SkipAddressable)

	SetNext(SkipAddressable)
	SetPrev(SkipAddressable)
	IsDone() bool
	New() SkipAddressable
}

type SkipItem struct {
	Title   string
	Text    string
	DueDate time.Time
	Done    bool
	ID      uint8
	P       SkipAddressable // either points to another skipitem or a category
	N       SkipAddressable // same rules as above
	Parent  *SkipCategory   // such that Attach / Detach can be run as necessary, views can be peeked, etc.
}

func mkSItem(Parent *SkipCategory) *SkipItem {
	return &SkipItem{
		Title:  "",
		Text:   "(empty note)",
		Done:   false,
		P:      Parent,
		N:      nil,
		Parent: Parent,
	}
}

func (si *SkipItem) Str() string {
	template := "	%s -- %s 				(%s)\n"
	title := FaintStyle.Render(si.Title)
	text := si.Text
	var date string
	if !si.DueDate.IsZero() {
		date = si.DueDate.Format("02/01/2006")
	} else {
		date = ItalicStyle.Render(" no  date ")
	}
	if si.Done { // todo: fix
		return StrikeStyle.Render(fmt.Sprintf(template, title, text, date))
	}

	return fmt.Sprintf(template, title, text, date)
}

func (si *SkipItem) Toggle() {
	si.Done = !si.Done
}

func (si *SkipItem) IsDone() bool {
	return si.Done
}

func (si *SkipItem) SetNext(s SkipAddressable) {
	si.N = s
}

func (si *SkipItem) SetPrev(s SkipAddressable) {
	si.P = s
}

func (si *SkipItem) Next() (bool, SkipAddressable) {
	if si.N == nil {
		return false, nil
	}
	if si.Parent.View == Incomplete {
		var res = true
		s := si.N
		for s != nil && s.IsDone() {
			res, s = s.Next()
			if !res {
				break
			}
		}
		return res, s
	}
	return true, si.N
}

func (si *SkipItem) Prev() (bool, SkipAddressable) {
	if si.P == nil {
		return false, nil
	}
	if si.Parent.View == Incomplete {
		var res = true
		s := si.P
		for s != nil && s.IsDone() {
			res, s = s.Prev()
			if !res {
				break
			}
		}
		return res, s

	}
	return true, si.P
}

func (si *SkipItem) New() SkipAddressable {
	tmp := mkSItem(si.Parent)
	tmp.P = si
	if si.N != nil {
		si.N.SetPrev(tmp)
	}
	tmp.N = si.N
	si.N = tmp
	si.Parent.ItemCt++
	return tmp
}

func (si *SkipItem) ReceiveTitle(s string) bool {
	si.Title = s
	return true
}

func (si *SkipItem) ReceiveText(s string) bool {
	si.Text = s
	return true
}

func (si *SkipItem) ReceiveDate(s string) bool {
	/*
		date validation:
		dd: assumes same year and month as current
		mm/dd: assumes same year as current
		mm/dd/yy: no assumption

		+n: adds n days to current day
	*/
	switch len(s) {
	case 2:
		// days mode
		if s[0] == '+' {
			day, err := strconv.Atoi(s[1:])
			if err != nil {
				return false
			}
			si.DueDate = time.Now().AddDate(0, 0, day)
		} else {
			t := time.Now()
			day, err := strconv.Atoi(s)
			if err != nil {
				return false
			}
			dayDiff := day - t.Day()
			if dayDiff < 0 {
				return false
			}
			si.DueDate = t.AddDate(0, 0, dayDiff)
		}
	case 5:
		t, err := time.Parse("02/01", s)
		t = t.AddDate(time.Now().Year(), 0, 0)
		if err != nil {
			return false
		}
		si.DueDate = t
	case 8:
		t, err := time.Parse("02/01/06", s)
		if err != nil {
			return false
		}
		si.DueDate = t
	}
	return true
}

func (si *SkipItem) GetModifiable() *[]NamedReceiver {
	return &[]NamedReceiver{
		{Name: "title: ", Input: si.ReceiveTitle},
		{Name: "text: ", Input: si.ReceiveText},
		{Name: "date: ", Input: si.ReceiveDate},
	}
}

type SkipCategory struct {
	Color    lipgloss.Color
	Title    string
	LastItem *SkipItem

	N      SkipAddressable // should always be an item
	P      SkipAddressable
	NC     *SkipCategory
	PC     *SkipCategory
	ItemCt int

	View ViewMode
}

func mkSCategory(name string) *SkipCategory {
	res := &SkipCategory{
		Color:  lipgloss.Color(DefaultColor),
		Title:  name,
		NC:     nil,
		PC:     nil,
		P:      nil,
		ItemCt: 1,
		View:   All,
	}
	res.LastItem = mkSItem(res)
	res.N = res.LastItem
	return res
}

// todo: handle stringification of data as a whole in main rather than here, so that "current" is easier to handle

func (sc *SkipCategory) Str() string {
	var res string
	template := "%s					(%d items) %s\n"
	title := ItalicStyle.Render(sc.Title)
	indicator := "⌅"
	switch sc.View {
	case Incomplete:
		indicator = "⌃"
	case Compacted:
		indicator = "⌄"
	}
	res = fmt.Sprintf(template, title, sc.ItemCt, indicator)

	res = lipgloss.NewStyle().Foreground(sc.Color).Render(res) + "\n"

	return res

}

func (sc *SkipCategory) Toggle() {
	if sc.View < All {
		sc.View++
		return
	}
	sc.View = Incomplete
}

func (sc *SkipCategory) ReceiveTitle(s string) bool {
	sc.Title = s
	return true
}

func (sc *SkipCategory) ReceiveColour(s string) bool {
	/*
		colour validation:
		if the string is a 6-character sequence, the colour type is assumed to be hex, with RRGGBB format
		if the string is a 3-character sequence, the colour type is assumed to be 8-bit ANSI.
	*/
	start := ""
	switch len(s) {
	case 3:
		if _, err := strconv.Atoi(s); err != nil {
			return false
		}
	case 6:
		if _, err := strconv.ParseUint(s, 16, 32); err != nil {
			return false
		}
		start = "#"
	}
	sc.Color = lipgloss.Color(start + s)
	return true
}

func (sc *SkipCategory) GetModifiable() *[]NamedReceiver {
	return &[]NamedReceiver{
		{Name: "title: ", Input: sc.ReceiveTitle},
		{Name: "colour: ", Input: sc.ReceiveColour},
	}

}

func (sc *SkipCategory) SetPrev(s SkipAddressable) {
	sc.P = s
}
func (sc *SkipCategory) SetNext(s SkipAddressable) {
	sc.N = s
}

func (sc *SkipCategory) Next() (bool, SkipAddressable) {
	if sc.View == Compacted {
		return sc.NextCat()
	}
	if sc.N != nil {
		return true, sc.N
	}
	return false, nil
}

func (sc *SkipCategory) Prev() (bool, SkipAddressable) {
	if sc.PC != nil && sc.PC.View == Compacted {
		return sc.PrevCat()
	}
	if sc.P != nil {
		return true, sc.P
	}
	return false, nil

}

func (sc *SkipCategory) New() SkipAddressable {
	tmp := mkSCategory("(untitled)")
	// category-level pointers
	tmp.PC = sc
	tmp.NC = sc.NC
	sc.NC = tmp

	// lower-level pointers
	e, r := sc.N.Next()
	sc.End().SetNext(tmp)
	tmp.SetPrev(sc.End())

	if e {
		tmp.N.SetNext(r)
	}
	if r != nil {
		r.SetPrev(tmp.N)
	}

	return tmp

}

func (sc *SkipCategory) NextCat() (bool, SkipAddressable) {
	if sc.NC != nil {
		return true, sc.NC
	}
	return false, nil

}

func (sc *SkipCategory) PrevCat() (bool, SkipAddressable) {
	if sc.PC != nil {
		return true, sc.PC
	}
	return false, nil
}

func (sc *SkipCategory) End() SkipAddressable {
	return sc.LastItem
}

func (sc *SkipCategory) IsDone() bool {
	return false
}

// ! TODO: peekView

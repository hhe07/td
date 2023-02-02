package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// categories will be sorted alphabetically
type Category struct {
	Color       lipgloss.Color
	Title       string
	Items       []*Item // pointers to the relevant items
	View        ViewMode
	CurrentItem int
}

// would use int to track progress in a more complex case.

// categories are always date sorted

func mkCategory(name string) *Category {
	return &Category{
		Color:       lipgloss.Color(DefaultColor),
		Title:       name,
		Items:       []*Item{mkItem()},
		CurrentItem: -1,
	}
}
func (c *Category) AddItem(i *Item) {
	c.Items = append(c.Items, i)
}

func (c *Category) Str() string {
	var res string
	template := "%s					(%d items)\n"
	title := ItalicStyle.Render(c.Title)
	res = fmt.Sprintf(template, title, len(c.Items))
	switch c.View {
	case Incomplete:
		for idx, i := range c.Items {
			if !i.Done {
				if idx == c.CurrentItem {
					res += "> "
				}
				res += "  " + i.Str()
			}
		}
	case All:
		for idx, i := range c.Items {
			if idx == c.CurrentItem {
				res += "> "
			}
			res += i.Str()
		}
	}
	res = lipgloss.NewStyle().Foreground(c.Color).Render(res) + "\n"
	return res
}

func (c *Category) ReceiveTitle(s string) bool {
	c.Title = s
	return true
}

func (c *Category) ReceiveColour(s string) bool {
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
	c.Color = lipgloss.Color(start + s)
	return true
}

func (c *Category) GetModifiable() *[]NamedReceiver {
	return &[]NamedReceiver{
		{Name: "title: ", Input: c.ReceiveTitle},
		{Name: "colour: ", Input: c.ReceiveColour},
	}
}

func (c *Category) Toggle() {
	if c.View < Compacted {
		c.View++
		return
	}
	c.View = Incomplete
}

func (c *Category) SelfID() AddressableType {
	return Category_a
}

func (c *Category) Uncapture() {
	c.CurrentItem = -1
}

func (c *Category) Prev() (bool, Addressable) {
	if c.CurrentItem <= 0 || c.View == Compacted {
		c.Uncapture()
		return false, nil
	}
	switch c.View {
	case Incomplete:
		for i := c.CurrentItem - 1; i >= 0; i-- {
			if !c.Items[i].Done {
				c.CurrentItem = i
				break
			}
		}
	case All:
		c.CurrentItem--
	}
	return true, c.Items[c.CurrentItem]
}

func (c *Category) Next() (bool, Addressable) {
	if c.CurrentItem == len(c.Items)-1 || c.View == Compacted {
		c.Uncapture()
		return false, nil
	}
	if c.CurrentItem == -1 {
		c.CurrentItem = 0
		return true, c.Items[c.CurrentItem]
	}
	switch c.View {
	case Incomplete:
		for i := c.CurrentItem + 1; i < len(c.Items); i++ {
			if !c.Items[i].Done {
				c.CurrentItem = i
				break
			}
		}
	case All:
		c.CurrentItem++
	}
	return true, c.Items[c.CurrentItem]
}

func mkItem() *Item {
	return &Item{
		Title: "",
		Text:  "(empty note)",
		Done:  false,
	}
}

type Item struct {
	Title   string
	Text    string
	DueDate time.Time
	Done    bool
	ID      uint8 // todo: keep for search?
}

func (i *Item) Toggle() {
	i.Done = !i.Done
}

func (i *Item) Str() string {
	template := "	%s -- %s 				(%s)\n"
	title := FaintStyle.Render(i.Title)
	text := i.Text
	var date string
	if !i.DueDate.IsZero() {
		date = i.DueDate.Format("02/01/2006")
	} else {
		date = ItalicStyle.Render(" no  date ")
	}
	if i.Done { // todo: fix
		return StrikeStyle.Render(fmt.Sprintf(template, title, text, date))
	}

	return fmt.Sprintf(template, title, text, date)
}

func (i *Item) ReceiveTitle(s string) bool {
	i.Title = s
	return true
}

func (i *Item) ReceiveText(s string) bool {
	i.Text = s
	return true
}

func (i *Item) ReceiveDate(s string) bool {
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
			i.DueDate = time.Now().AddDate(0, 0, day)
		} else {
			day, err := strconv.Atoi(s)
			if err != nil {
				return false
			}
			dayDiff := day - time.Now().Day()
			if dayDiff < 0 {
				return false
			}
			i.DueDate = time.Now()
			i.DueDate = i.DueDate.AddDate(0, 0, dayDiff)
		}
	case 5:
		t, err := time.Parse("02/01", s)
		t = t.AddDate(time.Now().Year(), 0, 0)
		if err != nil {
			return false
		}
		i.DueDate = t
	case 8:
		t, err := time.Parse("02/01/06", s)
		if err != nil {
			return false
		}
		i.DueDate = t
	}
	return true
}

func (i *Item) SelfID() AddressableType {
	return Item_a
}

func (i *Item) GetModifiable() *[]NamedReceiver {
	return &[]NamedReceiver{
		{Name: "title: ", Input: i.ReceiveTitle},
		{Name: "text: ", Input: i.ReceiveText},
		{Name: "date: ", Input: i.ReceiveDate},
	}
}

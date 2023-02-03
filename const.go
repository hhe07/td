package main

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type ViewMode uint8

const (
	Incomplete ViewMode = iota
	Compacted
	All
)

type AddressableType uint8

const (
	Item_a AddressableType = iota
	Category_a
)

type KeyMap struct {
	// when selecting a Category title, switches between Incomplete, All, Compacted views
	// when selecting a Item, toggles completion status.
	Toggle key.Binding

	// when in an editing mode, these switch between the modifiable parameters for the selected category / item
	Prev key.Binding
	Next key.Binding

	// quick traverses to next/previous category
	UpCategory   key.Binding
	DownCategory key.Binding
	Exit         key.Binding
	Edit         key.Binding
	Confirm      key.Binding
	New          key.Binding
}

const DefaultColor string = "#42f595"

var DefaultKeyMap = KeyMap{
	Toggle:       key.NewBinding(key.WithKeys("ctrl+d")),
	Prev:         key.NewBinding(key.WithKeys("up")),
	Next:         key.NewBinding(key.WithKeys("down")),
	UpCategory:   key.NewBinding(key.WithKeys("ctrl+up")),
	DownCategory: key.NewBinding(key.WithKeys("ctrl+down")),
	Exit:         key.NewBinding(key.WithKeys("ctrl+c")),
	Edit:         key.NewBinding(key.WithKeys("ctrl+e")),
	Confirm:      key.NewBinding(key.WithKeys("enter")),
	New:          key.NewBinding(key.WithKeys("ctrl+n")),
}

var FaintStyle = lipgloss.NewStyle().Faint(true)
var BoldStyle = lipgloss.NewStyle().Bold(true)
var ItalicStyle = lipgloss.NewStyle().Italic(true)
var DateStyle = lipgloss.NewStyle().PaddingLeft(10)
var StrikeStyle = lipgloss.NewStyle().Strikethrough(true)

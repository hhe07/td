package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// todo:
// todo: file in/out
// todo: fix spacing
// todo: fix strikethrough
// todo: deletion provision
// todo: redo internal data structure
// todo: help menu

type InputReceiver func(s string) bool

type ReceiverPool struct {
	Current   int
	Receivers *[]NamedReceiver
}

type NamedReceiver struct {
	Name  string
	Hint  string
	Input InputReceiver
}

type Addressable interface {
	Str() string
	Toggle()
	SelfID() AddressableType
	GetModifiable() *[]NamedReceiver
}

func (rp *ReceiverPool) Next() {
	if rp.Current < len(*rp.Receivers)-1 {
		rp.Current++
	}
}

func (rp *ReceiverPool) Prev() {
	if rp.Current > 0 {
		rp.Current--
	}
}
func (rp *ReceiverPool) Input(s string) bool {
	return (*rp.Receivers)[rp.Current].Input(s)
}

func (rp *ReceiverPool) Str() string {
	res := "editing:	\n"
	for i, v := range *rp.Receivers {
		if i == rp.Current {
			res += BoldStyle.Render(v.Name + " ")
		} else {
			res += v.Name + " "
		}
	}
	return res
}

type Main struct {
	First   SkipAddressable
	Last    SkipAddressable
	Current SkipAddressable

	Input   textinput.Model
	Editing bool
	Status  string

	Receivers ReceiverPool
	ModStatus bool
}

func MakeMain() Main {
	res := Main{
		First:   mkSCategory("uncategorised"),
		Input:   textinput.New(),
		Editing: false,
		Status:  "ok",
	}
	res.Current = res.First
	if s, r := res.Current.(*SkipCategory); r {
		res.Last = s.N
	}
	res.Input.Width = 20
	res.Input.Prompt = ""
	return res
}

func (m Main) Init() tea.Cmd {
	return nil
}

func (m Main) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Next):
			if m.Editing {
				m.Receivers.Next()
			} else {
				e, res := m.Current.Next()
				if e {
					m.Current = res
				}
			}
		case key.Matches(msg, DefaultKeyMap.Prev):
			if m.Editing {
				m.Receivers.Prev()
			} else {
				e, res := m.Current.Prev()
				if e {
					m.Current = res
				}
			}
		case key.Matches(msg, DefaultKeyMap.Toggle):
			m.Current.Toggle()
		case key.Matches(msg, DefaultKeyMap.UpCategory):
			switch r := m.Current.(type) {
			case *SkipCategory:
				s, res := r.PrevCat()
				if s {
					m.Current = res
				}
			case *SkipItem:
				t := r.Parent
				s, res := t.PrevCat()
				if s {
					m.Current = res
				}
			}
		case key.Matches(msg, DefaultKeyMap.DownCategory):
			switch r := m.Current.(type) {
			case *SkipCategory:
				s, res := r.NextCat()
				if s {
					m.Current = res
				}
			case *SkipItem:
				t := r.Parent
				s, res := t.NextCat()
				if s {
					m.Current = res
				}
			}
		case key.Matches(msg, DefaultKeyMap.Exit):
			if m.Editing {
				m.Editing = !m.Editing
				m.Input.Reset()
				m.Input.Blur()
				m.Input.Prompt = ""
			} else {
				return m, tea.Quit
			}
		case key.Matches(msg, DefaultKeyMap.New):
			if m.Current == m.Last {
				m.Current = m.Current.New()
				m.Last = m.Current
			} else { // WARN: categories aren't set properly
				if r, e := m.Current.(*SkipCategory); e {
					if r.LastItem == m.Last {
						m.Current = m.Current.New()
						a, b := m.Current.Next()
						if a {
							m.Last = b
						}
					}
				} else {
					m.Current = m.Current.New()
				}
			}
		case key.Matches(msg, DefaultKeyMap.Confirm):
			if m.Editing {
				res := m.Receivers.Input(m.Input.Value())
				if res {
					m.Status = "ok"
				} else {
					m.Status = "invalid input"
				}
			}
		case key.Matches(msg, DefaultKeyMap.Edit):
			m.Editing = true
			m.Receivers.Receivers = m.Current.GetModifiable()
			m.Receivers.Current = 0
			m.Input.Prompt = ">"
			m.Input.Focus()
		default:
			if m.Editing {
				m.ModStatus = true
				var cmd tea.Cmd
				m.Input, cmd = m.Input.Update(msg)
				return m, cmd
			}

		}
	}
	return m, nil
}

func (m Main) View() string {
	var res string
	b := true
	s := m.First
	for b {
		if s == m.Current {
			res += "> "
		} else {
			res += "  "
		}
		res += s.Str()
		b, s = s.Next()
	}
	res += "\n"
	var statusbar string
	if m.Editing {
		statusbar = m.Receivers.Str()
	} else {
		statusbar += "\n"
	}

	statusbar += "\n"
	statusbar += m.Input.View()
	statusbar += "						status: " + m.Status
	statusbar += " | "
	if m.ModStatus {
		statusbar += "( changes made ) \n"
	} else {
		statusbar += "(no changes made) \n"
	}
	res += statusbar
	return res
}

func main() {
	m := MakeMain()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}

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
// todo: change status in statusbar
// todo: fix spacing
// todo: fix strikethrough
// todo: deletion provision
// todo: redo internal data structure
// todo: make display style explicit

type InputReceiver func(s string) bool

type NamedReceiver struct {
	Name  string
	Input InputReceiver
}

type Addressable interface {
	Str() string
	Toggle()
	SelfID() AddressableType
	GetModifiable() *[]NamedReceiver
}

type Main struct {
	Categories      []*Category
	Input           textinput.Model
	Editing         bool
	Status          string
	Current         Addressable
	CurrentReceiver NamedReceiver
	Modifiable      *[]NamedReceiver
	CurrentCat      int
}

func MakeMain() Main {
	res := Main{
		Categories: []*Category{mkCategory("uncategorised")},
		CurrentCat: 0,
		Input:      textinput.New(),
		Editing:    false,
	}
	res.Current = res.Categories[0]
	res.Input.Width = 10
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
				for k, v := range *m.Modifiable {
					if v.Name == m.CurrentReceiver.Name && k < len(*m.Modifiable)-1 {
						m.CurrentReceiver = (*m.Modifiable)[k+1]
						break
					}

				}
			} else if m.Current.SelfID() == Item_a {
				r, t := m.Categories[m.CurrentCat].Next()
				if !r {
					if m.CurrentCat < len(m.Categories)-1 {
						m.Current = m.Categories[m.CurrentCat+1]
						m.CurrentCat++
					}
				} else {
					m.Current = t
				}
			} else if m.Current.SelfID() == Category_a {
				r, t := m.Categories[m.CurrentCat].Next()
				if !r {
					if m.CurrentCat < len(m.Categories)-1 {
						m.CurrentCat++
						m.Current = m.Categories[m.CurrentCat]
					}
				} else {
					m.Current = t
				}
			}
		case key.Matches(msg, DefaultKeyMap.Prev):
			if m.Editing {
				for k, v := range *m.Modifiable {
					if v.Name == m.CurrentReceiver.Name && k > 0 {
						m.CurrentReceiver = (*m.Modifiable)[k-1]
						break
					}
				}
			} else if m.Current.SelfID() == Item_a {
				r, t := m.Categories[m.CurrentCat].Prev()
				if !r {
					m.Current = m.Categories[m.CurrentCat]
				} else {
					m.Current = t
				}
			} else if m.Current.SelfID() == Category_a {
				r, t := m.Categories[m.CurrentCat].Prev()
				if !r {
					if m.CurrentCat > 0 {
						m.CurrentCat--
						tmp := m.Categories[m.CurrentCat].Items
						l := len(tmp) - 1
						m.Current = tmp[l]
						m.Categories[m.CurrentCat].CurrentItem = l
					}
				} else {
					m.Current = t
				}
			}
		case key.Matches(msg, DefaultKeyMap.Toggle):
			m.Current.Toggle()
		case key.Matches(msg, DefaultKeyMap.UpCategory):
			if m.CurrentCat > 0 {
				m.Categories[m.CurrentCat].Uncapture()
				m.CurrentCat--
				m.Current = m.Categories[m.CurrentCat]
			}
		case key.Matches(msg, DefaultKeyMap.DownCategory):
			if m.CurrentCat < len(m.Categories)-1 {
				m.Categories[m.CurrentCat].Uncapture()
				m.CurrentCat++
				m.Current = m.Categories[m.CurrentCat]
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
			if m.Current.SelfID() == Item_a {
				tmp := m.Categories[m.CurrentCat].Items
				l := len(tmp)
				tmp = append(tmp, mkItem())
				m.Categories[m.CurrentCat].Items = tmp
				m.Categories[m.CurrentCat].CurrentItem = l
				m.Current = m.Categories[m.CurrentCat].Items[l]
			} else {
				m.Categories = append(m.Categories, mkCategory("new category"))
				m.Current = m.Categories[len(m.Categories)-1]
				m.CurrentCat++
			}
		case key.Matches(msg, DefaultKeyMap.Confirm):
			if m.Editing {
				res := m.CurrentReceiver.Input(m.Input.Value())
				if res {
					m.Status = "ok"
				} else {
					m.Status = "invalid input"
				}
			}
		case key.Matches(msg, DefaultKeyMap.Edit):
			m.Editing = true
			m.Modifiable = m.Current.GetModifiable()
			m.CurrentReceiver = (*m.Modifiable)[0]
			m.Input.Prompt = ">"
			m.Input.Focus()
		default:
			if m.Editing {
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
	for i := 0; i < len(m.Categories); i++ {
		if i == m.CurrentCat && m.Current.SelfID() == Category_a {
			res += "> "
		} else {
			res += "  "
		}
		res += m.Categories[i].Str()
	}
	res += "\n"
	var statusbar string
	if m.Editing {
		statusbar += "editing:	\n"
		for _, v := range *m.Modifiable {
			if v.Name == m.CurrentReceiver.Name {
				statusbar += BoldStyle.Render(v.Name + " ")
			} else {
				statusbar += v.Name + " "
			}
		}
	} else {
		statusbar += "\n"
	}

	statusbar += "\n"
	statusbar += m.Input.View()
	statusbar += "							status: " + m.Status
	res += statusbar
	return res
}

// todo: make sure to init a nocategory as a category
// todo: make sure to add arrow on display for selected category
func main() {
	m := MakeMain()
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}

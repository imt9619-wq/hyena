package form

import (
	"encoding/json"
	"iter"
)

type Form interface {
	ResponseJson() []byte
	FormID() uint32
	Title() string
}

const(
	FormCancel int = iota-1
	FormTrue
	FormFalse
)

type Menu struct {
	formId     uint32
    title      string
    body       string
    elems      []Element
    respButton int
}

func (m *Menu) PressButtonByIndex(ind int) bool{
	if ind < 0{
		return false
	}
	buttonInd := 0
	for _, elem := range m.elems{
		if _, ok := elem.(*Button); ok{
			if ind == buttonInd{
				m.respButton = buttonInd
				return true
			}
			buttonInd++
		}
	}
	return false
}

func (m *Menu) PressButton(buttonName string) bool{
	buttonInd := 0
	for _, elem := range m.elems{
		if button, ok := elem.(*Button); ok{
			if button.Text == buttonName{
				m.respButton = buttonInd
				return true
			}
			buttonInd++
		}
	}
	return false
}

func (m *Menu) FormID() uint32 {
	return m.formId
}

func (m *Menu) Title() string {
	return m.title
}

func (m *Menu) Body() string {
	return m.body
}

func (m *Menu) Elements() []Element {
	return m.elems
}

func (m *Menu) ResponseJson() []byte {
	if m.respButton == FormCancel{
		return nil
	}
	data, _ := json.Marshal(m.respButton) 
	return data
}

type Modal struct {
	formId           uint32
    title            string
    body             string
    button1, button2 *Button
    resp             int
}

func (m *Modal) Press(button bool){
	if button == false{
		m.resp = FormFalse
	}else{
		m.resp = FormTrue
	}
}

func (m *Modal) PressButton(buttName string) bool{
	if m.button1.Text == buttName{
		m.resp = FormTrue
	}else if m.button2.Text == buttName{
		m.resp = FormFalse
	}else{
		return false
	}
	return true
}

func (m *Modal) FormID() uint32 {
	return m.formId
}

func (m *Modal) Title() string {
	return m.title
}

func (m *Modal) Body() string {
	return m.body
}

func (m *Modal) Buttons() []*Button {
	return []*Button{m.button1, m.button2}
}

func (m *Modal) ResponseJson() []byte {
	resp := false
	if m.resp == FormCancel{
		return nil
	}
	if m.resp == FormTrue{
		resp = true
	}
	data, _ := json.Marshal(resp)
	return data
}

type Custom struct {
	formId uint32
	title  string
	elems  []Element
	resp   any
}

func (c *Custom) CustomElementWithName(name string) iter.Seq[CustomElement]{
	return func(yield func(CustomElement) bool) {
		for _, elem := range c.elems{
			_, ok := elem.(*Button)
			if elem.ReadOnly() == true || ok{
				continue
			}
			if cElem, ok := elem.(CustomElement); ok{
				if cElem.Name() != name{
					continue
				}
				if !yield(cElem){
					return 
				}
			}
		}
	}
	
}

func (c *Custom) CustomElementWithType(t CustomElement) iter.Seq[CustomElement]{
	return func(yield func(CustomElement) bool) {
		for _, elem := range c.elems{
			_, ok := elem.(*Button)
			if elem.ReadOnly() == true || ok{
				continue
			}
			cElem, ok := elem.(CustomElement)
			if !ok{
				continue
			}
			switch t.(type){
			case *Input:
				if _, ok := cElem.(*Input); !ok{
					continue
				}
			case *Dropdown:
				if _, ok := cElem.(*Dropdown); !ok{
					continue
				}
			case *Slider:
				if _, ok := cElem.(*Slider); !ok{
					continue
				}
			case *Toggle:
				if _, ok := cElem.(*Toggle); !ok{
					continue
				}
			case *StepSlider:
				if _, ok := cElem.(*StepSlider); !ok{
					continue
				}
			default:
				continue
			}
			if !yield(cElem){
				return 
			}
		}
	}
}


func (c *Custom) FormID() uint32 {
	return c.formId
}

func (c *Custom) Title() string {
	return c.title
}

func (c *Custom) Elements() []Element {
	return c.elems
}

func (c *Custom) ResponseJson() []byte {
	data, _ := json.Marshal(c.elems)
	return data
}

type rawForm struct {
	Type     string          `json:"type"`
	Title    string          `json:"title"`
	Content  json.RawMessage `json:"content"`
	Elements json.RawMessage `json:"elements"`
	Button1  string          `json:"button1"`
	Button2  string          `json:"button2"`
}

func UnmarshalForm(id uint32, data []byte) (f Form, ok bool) {
	f, ok = nil, false
	var rawForm rawForm
	err := json.Unmarshal(data, &rawForm)
	if err != nil {
		return
	}

	switch rawForm.Type {
	case "form":
		var body string
		err = json.Unmarshal(rawForm.Content, &body)
		if err != nil{
			return
		}
		e, alr := unMarshalElement(rawForm.Elements)
		if !alr{
			return
		}
		ok = true
		f = &Menu{
			formId: id,
			title:  rawForm.Title,
			body:   body,
			elems:  e,
			respButton: FormCancel,
		}
	case "custom_form":
		e, alr := unMarshalElement(rawForm.Content)
		if !alr{
			return
		}
		ok = true
		f = &Custom{
			formId: id,
			title:  rawForm.Title,
			elems:  e,
		}
	case "modal":
		var body string
		err = json.Unmarshal(rawForm.Content, &body)
		if err != nil{
			return
		}
		ok = true
		f = &Modal{
			formId:  id,
			title:   rawForm.Title,
			body:    body,
			button1: &Button{Text: rawForm.Button1},
			button2: &Button{Text: rawForm.Button2},
			resp: FormCancel,
		}
	}
	return 
}

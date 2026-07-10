// Element structs are copied from github.com/df-mc/server/player/form/element.go
package form

import (
	"encoding/json"
)

type Element interface{
	json.Marshaler
	ReadOnly() bool
}	

type CustomElement interface{
	Name() string
	SetValue(any)
	Element
}

type Button struct {
	Text  string `json:"text"`
	Image struct {
		Type string `json:"type"`
		Data string `json:"data"`
	} `json:"image"`
}

func (*Button) MarshalJSON() ([]byte, error){
	return json.Marshal(nil)
}

func (*Button) ReadOnly() bool{
	return false
}

type Divider struct{}

func (*Divider) MarshalJSON() ([]byte, error){
	return json.Marshal(nil)
}

func (*Divider) ReadOnly() bool{
	return true
}

type Header struct {
	Text string `json:"text"`
}

func (*Header) MarshalJSON() ([]byte, error){
	return json.Marshal(nil)
}

func (*Header) ReadOnly() bool{
	return true
}

type Label struct {
	Text string `json:"text"`
}

func (*Label) MarshalJSON() ([]byte, error){
	return json.Marshal(nil)
}

func (*Label) ReadOnly() bool{
	return true
}

type Input struct {
	Text        string `json:"text"`
	Default     string `json:"default"`
	Placeholder string `json:"placeholder"`
	Tooltip     string `json:"tooltip"`
	value string
}

func (i *Input) MarshalJSON() ([]byte, error){
	return json.Marshal(i.value)
}

func (i *Input) SetValue(val any){
	if v, ok := val.(string); ok{
		i.value = v
	}
}

func (i *Input) Name() string{
	return i.Text
}

func (*Input) ReadOnly() bool{
	return false
}

type Toggle struct {
	Text    string `json:"text"`
	Default bool   `json:"default"`
	Tooltip string `json:"tooltip"`
	value bool
}

func (t *Toggle) MarshalJSON() ([]byte, error){
	return json.Marshal(t.value)
}

func (t *Toggle) SetValue(val any){
	if v, ok := val.(bool); ok{
		t.value = v
	}
}

func (*Toggle) ReadOnly() bool{
	return false
}

func (t *Toggle) Name() string{
	return t.Text
}

type Slider struct {
	Text     string  `json:"text"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	StepSize float64 `json:"step"`
	Default  float64 `json:"default"`
	Tooltip  string  `json:"tooltip"`
	value float64
}

func (s *Slider) MarshalJSON() ([]byte, error){
	return json.Marshal(s.value)
}

func (s *Slider) Name() string{
	return s.Text
}

func (*Slider) ReadOnly() bool{
	return false
}

func (s *Slider) SetValue(val any){
	if v, ok := val.(float64); ok{
		s.value = v
	}
}

type Dropdown struct {
	Text         string   `json:"text"`
	Options      []string `json:"options"`
	DefaultIndex int      `json:"default"`
	Tooltip      string   `json:"tooltip"`
	value int
}

func (d *Dropdown) MarshalJSON() ([]byte, error){
	return json.Marshal(d.value)
}

func (*Dropdown) ReadOnly() bool{
	return false
}

func (d *Dropdown) Name() string{
	return d.Text
}

func (d *Dropdown) SetValue(val any){
	if v, ok := val.(int); ok{
		d.value = v
	}
}

type StepSlider struct {
	Text         string   `json:"text"`
	Step         []string `json:"steps"`
	DefaultIndex int      `json:"default"`
	Tooltip      string   `json:"tooltip"`
	value int
}

func (s *StepSlider) MarshalJSON() ([]byte, error){
	return json.Marshal(s.value)
}

func (s *StepSlider) SetValue(val any){
	if v, ok := val.(int); ok{
		s.value = v
	}
}

func (*StepSlider) ReadOnly() bool{
	return false
}

func (s *StepSlider) Name() string{
	return s.Text
}

type rawElement struct {
	Type string `json:"type"`
}

func unMarshalElement(data []byte) (e []Element, ok bool) {
	e, ok = nil, false
	var rawElems []json.RawMessage
	err := json.Unmarshal(data, &rawElems)
	if err != nil {
		return
	}

	elems := make([]Element, 0, len(rawElems))
	for _, rawElemMgs := range rawElems {
		var rawElem rawElement
		err := json.Unmarshal(rawElemMgs, &rawElem)
		if err != nil {
			return
		}
		switch rawElem.Type {
		case "input":
			var elem Input
			err = json.Unmarshal(rawElemMgs, &elem)
			if err != nil {
				return 
			}
			elem.value = elem.Default
			elems = append(elems, &elem)
		case "toggle":
			var elem Toggle
			err = json.Unmarshal(rawElemMgs, &elem)
			if err != nil {
				return 
			}
			elem.value = elem.Default
			elems = append(elems, &elem)
		case "slider":
			var elem Slider
			err = json.Unmarshal(rawElemMgs, &elem)
			if err != nil {
				return 
			}
			elem.value = elem.Default
			elems = append(elems, &elem)
		case "dropdown":
			var elem Dropdown
			err = json.Unmarshal(rawElemMgs, &elem)
			if err != nil {
				return 
			}
			elem.value = elem.DefaultIndex
			elems = append(elems, &elem)
		case "step_slider":
			var elem StepSlider
			err = json.Unmarshal(rawElemMgs, &elem)
			if err != nil {
				return 
			}
			elem.value = elem.DefaultIndex
			elems = append(elems, &elem)
		case "label":
			var elem Label
			err = json.Unmarshal(rawElemMgs, &elem)
			if err != nil {
				return 
			}
			elems = append(elems, &elem)
		case "header":
			var elem Header
			err = json.Unmarshal(rawElemMgs, &elem)
			if err != nil {
				return 
			}
			elems = append(elems, &elem)
		case "divider":
			elems = append(elems, &Divider{})
		default:
			if rawElem.Type != "" {
				return
			}
			var elem Button
			err = json.Unmarshal(rawElemMgs, &elem)
			if err != nil {
				return 
			}
			elems = append(elems, &elem)
		}
	}
	return elems, true
}

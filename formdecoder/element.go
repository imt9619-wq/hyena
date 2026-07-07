// Element structs fields are copied from github.com/df-mc/server/player/form/element.go
package formdecoder

import "encoding/json"

type ResponceElement interface {
	json.Marshaler
}

type Cancel struct{}

type Element any

type Button struct{
	Text string
	Image string
}

type Divider struct{}

type Header struct {
	Text string
}

type Label struct {
	Text string
}

type Input struct {
	Text string
	Default string
	Placeholder string
	Tooltip string
	value string
}

type Toggle struct {
	Text string
	Default bool
	Tooltip string
	value bool
}

type Slider struct {
	Text string
	Min, Max float64
	StepSize float64
	Default float64
	Tooltip string
	value float64
}

type Dropdown struct {
	Text string
	Options []string
	DefaultIndex int
	Tooltip string
	value int
}

type StepSlider Dropdown
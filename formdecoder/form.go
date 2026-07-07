package formdecoder

import (
	"encoding/json"
)

type Form interface{
	ResponceJson(ResponceElement, ...any) []byte
	FormID() uint32
}

type Menu struct{
	formId uint32
	title string
	body string
	elems []Element
}

func (m Menu) FormID() uint32{
	return m.formId
}

func (m Menu) ResponceJson(elem ResponceElement, value ...any) []byte{
	return nil
}

type Model struct{
	formId uint32
	title string
	body string
	button1, button2 Button
}

func (m Model) FormID() uint32{
	return m.formId
}

func (m Model) ResponceJson(elem ResponceElement, value ...any) []byte{
	return nil
}

type Custom struct{
	formId uint32
	title string
	elems []Element
}

func (c Custom) FormID() uint32{
	return c.formId
}

func (c Custom) ResponceJson(elem ResponceElement, value ...any) []byte{
	return nil
}

func UnmarshalForm(id uint32, data []byte) Form{
	var formData map[string]any
	err := json.Unmarshal(data, &formData)
	if err != nil{
		return nil
	}
	formType, ok := formData["type"]
	if !ok{
		return nil
	}
	title := formData["title"].(string)

	switch formType.(string){
	case "form":
		return Menu{
			formId: id,
			title: title,
			body: formData["content"].(string),
			elems: unMarshalElement(formData["elements"].([]byte)),
		}
	case "custom_form":
		return Custom{
			formId: id,
			title: title,
			elems: unMarshalElement(formData["content"].([]byte)),
		}
	case "modal":
		return Model{
			formId: id,
			title: title,
			body: formData["content"].(string),
			button1: unMarshalElement(formData["button1"].([]byte))[0].(Button),
			button2: unMarshalElement(formData["button2"].([]byte))[0].(Button),
		}
	}
	return nil
}

func unMarshalElement(data []byte) []Element{
	return nil
}

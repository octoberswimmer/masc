package prop

import (
	"github.com/octoberswimmer/masc"
)

type InputType string

const (
	TypeButton        InputType = "button"
	TypeCheckbox      InputType = "checkbox"
	TypeColor         InputType = "color"
	TypeDate          InputType = "date"
	TypeDatetime      InputType = "datetime"
	TypeDatetimeLocal InputType = "datetime-local"
	TypeEmail         InputType = "email"
	TypeFile          InputType = "file"
	TypeHidden        InputType = "hidden"
	TypeImage         InputType = "image"
	TypeMonth         InputType = "month"
	TypeNumber        InputType = "number"
	TypePassword      InputType = "password"
	TypeRadio         InputType = "radio"
	TypeRange         InputType = "range"
	TypeMin           InputType = "min"
	TypeMax           InputType = "max"
	TypeValue         InputType = "value"
	TypeStep          InputType = "step"
	TypeReset         InputType = "reset"
	TypeSearch        InputType = "search"
	TypeSubmit        InputType = "submit"
	TypeTel           InputType = "tel"
	TypeText          InputType = "text"
	TypeTime          InputType = "time"
	TypeURL           InputType = "url"
	TypeWeek          InputType = "week"
)

func Autofocus(autofocus bool) masc.Applyer {
	return masc.Property("autofocus", autofocus)
}

func Disabled(disabled bool) masc.Applyer {
	return masc.Property("disabled", disabled)
}

func Checked(checked bool) masc.Applyer {
	return masc.Property("checked", checked)
}

func For(id string) masc.Applyer {
	return masc.Property("htmlFor", id)
}

func Href(url string) masc.Applyer {
	return masc.Property("href", url)
}

func ID(id string) masc.Applyer {
	return masc.Property("id", id)
}

func Placeholder(text string) masc.Applyer {
	return masc.Property("placeholder", text)
}

func Src(url string) masc.Applyer {
	return masc.Property("src", url)
}

func Type(t InputType) masc.Applyer {
	return masc.Property("type", string(t))
}

func Value(v string) masc.Applyer {
	return masc.Property("value", v)
}

func Name(name string) masc.Applyer {
	return masc.Property("name", name)
}

func Alt(text string) masc.Applyer {
	return masc.Property("alt", text)
}

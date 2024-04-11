package style

import (
	"strconv"

	"github.com/octoberswimmer/masc"
)

type Size string

func Px(pixels int) Size {
	return Size(strconv.Itoa(pixels) + "px")
}

func Color(value string) masc.Applyer {
	return masc.Style("color", value)
}

func Width(size Size) masc.Applyer {
	return masc.Style("width", string(size))
}

func MinWidth(size Size) masc.Applyer {
	return masc.Style("min-width", string(size))
}

func MaxWidth(size Size) masc.Applyer {
	return masc.Style("max-width", string(size))
}

func Height(size Size) masc.Applyer {
	return masc.Style("height", string(size))
}

func MinHeight(size Size) masc.Applyer {
	return masc.Style("min-height", string(size))
}

func MaxHeight(size Size) masc.Applyer {
	return masc.Style("max-height", string(size))
}

func Margin(size Size) masc.Applyer {
	return masc.Style("margin", string(size))
}

func Position(pos string) masc.Applyer {
	return masc.Style("position", pos)
}

func Bottom(size Size) masc.Applyer {
	return masc.Style("bottom", string(size))
}

func Top(size Size) masc.Applyer {
	return masc.Style("top", string(size))
}

func Left(size Size) masc.Applyer {
	return masc.Style("left", string(size))
}

func Right(size Size) masc.Applyer {
	return masc.Style("right", string(size))
}

type OverflowOption string

const (
	OverflowVisible OverflowOption = "visible"
	OverflowHidden  OverflowOption = "hidden"
	OverflowScroll  OverflowOption = "scroll"
	OverflowAuto    OverflowOption = "auto"
)

func Overflow(option OverflowOption) masc.Applyer {
	return masc.Style("overflow", string(option))
}

func OverflowX(option OverflowOption) masc.Applyer {
	return masc.Style("overflow-x", string(option))
}

func OverflowY(option OverflowOption) masc.Applyer {
	return masc.Style("overflow-y", string(option))
}

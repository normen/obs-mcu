package mcu

import (
	"fmt"
	"gonum.org/v1/gonum/interp"
	"math"
	"regexp"
	"strings"
	"unicode/utf8"
)

var re *regexp.Regexp

var faderToObs interp.PiecewiseLinear
var obsToFader interp.PiecewiseLinear

// prepares the interpolation for the mackie fader to obs fader translation
func InitInterp() {
	// fader values for -inf, -60, -50, -40, -30, -20, -10, -6, 0, +6, +10
	// -8192, -7460, -6512, -5142, -3578, -2376, -252, 1815, 4190, 6482, 8191
	// 0, 0.000952, 0.002919, 0.009665, 0.031204, 0.096298, 0.315420, 0.488820, 1, 1, 1
	faderVals := []float64{-8192, -7460, -6512, -5142, -3578, -2376, -252, 1815, 4190}
	obsVals := []float64{0, 0.000952, 0.002919, 0.009665, 0.031204, 0.096298, 0.315420, 0.488820, 1}
	faderToObs.Fit(faderVals, obsVals)
	obsToFader.Fit(obsVals, faderVals)
}

func ShortenText(input string) string {
	input = strings.ReplaceAll(input, "Input", "In")
	input = strings.ReplaceAll(input, "Output", "Out")
	if re == nil {
		re = regexp.MustCompile(`([^-_ ]+)[AEIOUaeiou]([^-_ ]+)`)
	}
	ret := re.FindAllString(input, 1)
	length := utf8.RuneCountInString(input)
	for length > 6 && ret != nil {
		input = re.ReplaceAllString(input, `$1$2`)
		//log.Println("Found", input)
		ret = re.FindAllString(input, 1)
		length = utf8.RuneCountInString(input)
	}
	if length > 6 {
		input = strings.ReplaceAll(input, " ", "")
		input = strings.ReplaceAll(input, "-", "")
		input = strings.ReplaceAll(input, "_", "")
		length = utf8.RuneCountInString(input)
	}
	if length > 6 {
		if match, _ := regexp.MatchString(".*[0-9][0-9][/-_][0-9][0-9]$", input); match {
			input = input[:3] + input[length-3:]
		} else if match, _ := regexp.MatchString(".*[0-9][/-_][0-9][0-9]$", input); match {
			input = input[:3] + input[length-3:]
		} else if match, _ := regexp.MatchString(".*[0-9][/-_][0-9]$", input); match {
			input = input[:4] + input[length-2:]
		} else if match, _ := regexp.MatchString(".*[0-9][0-9]$", input); match {
			input = input[:4] + input[length-2:]
		} else if match, _ := regexp.MatchString(".*[0-9]$", input); match {
			input = input[:5] + input[length-1:]
		}
		length = utf8.RuneCountInString(input)
	}
	if length < 6 {
		input = fmt.Sprintf("%-6s", input)
	} else if length > 6 {
		input = input[0:6]
	}
	return input
}

func FaderFloatToInt(level float64) int16 {
	// why do I have to add 8191?..
	level = obsToFader.Predict(level) + 8191
	return int16(level)
}

func IntToFaderFloat(faderVal int16) float64 {
	level := faderToObs.Predict(float64(faderVal))
	return level
}

func MapToRange(value, fromMin, fromMax, toMin, toMax float64) float64 {
	if fromMax == fromMin {
		panic("Ursprünglicher Bereich darf nicht Null sein")
	}
	return ((value-fromMin)*(toMax-toMin)/(fromMax-fromMin) + toMin)
}

// LinearToDb wandelt einen linearen Faktor in Dezibel um.
func LinearToDb(linear float64) float64 {
	return 20 * math.Log10(linear)
}

// DbToLinear wandelt einen Dezibelwert in einen linearen Faktor um.
func DbToLinear(db float64) float64 {
	return math.Pow(10, db/20)
}

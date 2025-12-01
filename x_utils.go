package twitterscraper

import (
	"fmt"
	"math"
	"strings"
)

func interpolate(fromVal, toVal []float64, value float64) []float64 {
	res := make([]float64, len(fromVal))
	for i := range fromVal {
		res[i] = fromVal[i] + (toVal[i]-fromVal[i])*value
	}
	return res
}

func convertRotationToMatrix(degrees float64) []float64 {
	rad := degrees * math.Pi / 180
	cos := math.Cos(rad)
	sin := math.Sin(rad)
	return []float64{cos, -sin, sin, cos}
}

func isOdd(num float64) float64 {
	if int(num)%2 != 0 {
		return -1.0
	}
	return 0.0
}

func floatToHex(x float64) string {
	result := []string{}
	quotient := int(x)
	fraction := x - float64(quotient)

	for quotient > 0 {
		q := quotient / 16
		remainder := quotient - (q * 16)
		if remainder > 9 {
			result = append([]string{string(rune(remainder + 55))}, result...)
		} else {
			result = append([]string{fmt.Sprintf("%d", remainder)}, result...)
		}
		quotient = q
	}

	if fraction == 0 {
		if len(result) == 0 {
			return ""
		}
		return strings.Join(result, "")
	}

	if len(result) == 0 {
		result = append(result, "0")
	}
	result = append(result, ".")

	for fraction > 0 {
		fraction *= 16
		integer := int(fraction)
		fraction -= float64(integer)

		if integer > 9 {
			result = append(result, string(rune(integer+55)))
		} else {
			result = append(result, fmt.Sprintf("%d", integer))
		}
	}

	return strings.Join(result, "")
}

func round(val float64) float64 {
	return math.Round(val)
}

func round2(val float64) float64 {
	return math.Round(val*100) / 100
}

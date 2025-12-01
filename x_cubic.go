package twitterscraper

import "math"

type cubic struct {
	curves []float64
}

func newCubic(curves []float64) *cubic {
	return &cubic{curves: curves}
}

func (c *cubic) getValue(timeVal float64) float64 {
	startGradient := 0.0
	endGradient := 0.0
	start := 0.0
	mid := 0.0
	end := 1.0

	if timeVal <= 0.0 {
		if c.curves[0] > 0.0 {
			startGradient = c.curves[1] / c.curves[0]
		} else if c.curves[1] == 0.0 && c.curves[2] > 0.0 {
			startGradient = c.curves[3] / c.curves[2]
		}
		return startGradient * timeVal
	}

	if timeVal >= 1.0 {
		if c.curves[2] < 1.0 {
			endGradient = (c.curves[3] - 1.0) / (c.curves[2] - 1.0)
		} else if c.curves[2] == 1.0 && c.curves[0] < 1.0 {
			endGradient = (c.curves[1] - 1.0) / (c.curves[0] - 1.0)
		}
		return 1.0 + endGradient*(timeVal-1.0)
	}

	for start < end {
		mid = (start + end) / 2
		xEst := c.calculate(c.curves[0], c.curves[2], mid)
		if math.Abs(timeVal-xEst) < 0.00001 {
			return c.calculate(c.curves[1], c.curves[3], mid)
		}
		if xEst < timeVal {
			start = mid
		} else {
			end = mid
		}
	}
	return c.calculate(c.curves[1], c.curves[3], mid)
}

func (c *cubic) calculate(a, b, m float64) float64 {
	return 3.0*a*(1-m)*(1-m)*m + 3.0*b*(1-m)*m*m + m*m*m
}


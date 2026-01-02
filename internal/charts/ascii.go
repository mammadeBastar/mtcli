package charts

import (
	"fmt"
	"math"
	"strings"
)

// DataPoint represents a point in the chart
type DataPoint struct {
	TimeMs int64
	Value  float64
}

// ChartOptions configures the chart rendering
type ChartOptions struct {
	Width     int
	Height    int
	ShowAxis  bool
	Title     string
	ValueUnit string // e.g., "WPM"
}

// DefaultOptions returns sensible default chart options
func DefaultOptions() ChartOptions {
	return ChartOptions{
		Width:     60,
		Height:    10,
		ShowAxis:  true,
		ValueUnit: "WPM",
	}
}

// RenderChart renders a series of data points as an ASCII chart
func RenderChart(points []DataPoint, opts ChartOptions) string {
	if len(points) == 0 {
		return "No data"
	}

	// Ensure minimum dimensions
	if opts.Width < 20 {
		opts.Width = 20
	}
	if opts.Height < 5 {
		opts.Height = 5
	}

	// Find min/max values
	minVal, maxVal := findMinMax(points)

	// Add some padding to the range
	valRange := maxVal - minVal
	if valRange < 1 {
		valRange = 1
	}
	minVal = math.Max(0, minVal-valRange*0.1)
	maxVal = maxVal + valRange*0.1

	// Calculate actual chart area (accounting for axis labels)
	axisWidth := 6 // Width for Y-axis labels
	chartWidth := opts.Width - axisWidth
	if chartWidth < 10 {
		chartWidth = 10
	}

	// Create the chart grid
	grid := make([][]rune, opts.Height)
	for i := range grid {
		grid[i] = make([]rune, chartWidth)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Plot points
	for _, point := range points {
		// Map time to X coordinate
		x := mapToRange(float64(point.TimeMs), 0, float64(points[len(points)-1].TimeMs), 0, float64(chartWidth-1))
		xIdx := int(math.Round(x))
		if xIdx < 0 {
			xIdx = 0
		}
		if xIdx >= chartWidth {
			xIdx = chartWidth - 1
		}

		// Map value to Y coordinate (inverted because row 0 is top)
		y := mapToRange(point.Value, minVal, maxVal, float64(opts.Height-1), 0)
		yIdx := int(math.Round(y))
		if yIdx < 0 {
			yIdx = 0
		}
		if yIdx >= opts.Height {
			yIdx = opts.Height - 1
		}

		// Use different characters for different chart styles
		grid[yIdx][xIdx] = '█'
	}

	// Connect points with a line (optional, makes chart more readable)
	connectPoints(grid, points, minVal, maxVal, chartWidth, opts.Height)

	// Build output string
	var sb strings.Builder

	// Title
	if opts.Title != "" {
		sb.WriteString(opts.Title)
		sb.WriteRune('\n')
	}

	// Render grid with Y-axis labels
	for row := 0; row < opts.Height; row++ {
		if opts.ShowAxis {
			// Calculate value at this row
			val := mapToRange(float64(row), 0, float64(opts.Height-1), maxVal, minVal)
			if row == 0 || row == opts.Height-1 || row == opts.Height/2 {
				sb.WriteString(fmt.Sprintf("%5.0f│", val))
			} else {
				sb.WriteString("     │")
			}
		}
		sb.WriteString(string(grid[row]))
		sb.WriteRune('\n')
	}

	// X-axis
	if opts.ShowAxis {
		sb.WriteString("     └")
		sb.WriteString(strings.Repeat("─", chartWidth))
		sb.WriteRune('\n')

		// Time labels
		totalMs := points[len(points)-1].TimeMs
		sb.WriteString("     ")
		sb.WriteString(fmt.Sprintf("0s"))
		midPadding := chartWidth/2 - 2
		if midPadding > 0 {
			sb.WriteString(strings.Repeat(" ", midPadding))
			sb.WriteString(fmt.Sprintf("%ds", totalMs/2000))
		}
		endPadding := chartWidth - midPadding - 6
		if endPadding > 0 {
			sb.WriteString(strings.Repeat(" ", endPadding))
			sb.WriteString(fmt.Sprintf("%ds", totalMs/1000))
		}
		sb.WriteRune('\n')
	}

	return sb.String()
}

// RenderDualChart renders two data series on the same chart (e.g., WPM and Raw WPM)
func RenderDualChart(primary, secondary []DataPoint, opts ChartOptions) string {
	if len(primary) == 0 && len(secondary) == 0 {
		return "No data"
	}

	// Combine all points to find range
	allPoints := append(primary, secondary...)
	minVal, maxVal := findMinMax(allPoints)

	valRange := maxVal - minVal
	if valRange < 1 {
		valRange = 1
	}
	minVal = math.Max(0, minVal-valRange*0.1)
	maxVal = maxVal + valRange*0.1

	axisWidth := 6
	chartWidth := opts.Width - axisWidth
	if chartWidth < 10 {
		chartWidth = 10
	}

	// Create the chart grid
	grid := make([][]rune, opts.Height)
	for i := range grid {
		grid[i] = make([]rune, chartWidth)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Find max time
	var maxTime int64
	if len(primary) > 0 {
		maxTime = primary[len(primary)-1].TimeMs
	}
	if len(secondary) > 0 && secondary[len(secondary)-1].TimeMs > maxTime {
		maxTime = secondary[len(secondary)-1].TimeMs
	}
	if maxTime == 0 {
		maxTime = 1
	}

	// Plot secondary series (Raw WPM) with lighter character
	for _, point := range secondary {
		x := mapToRange(float64(point.TimeMs), 0, float64(maxTime), 0, float64(chartWidth-1))
		y := mapToRange(point.Value, minVal, maxVal, float64(opts.Height-1), 0)
		xIdx := clampInt(int(math.Round(x)), 0, chartWidth-1)
		yIdx := clampInt(int(math.Round(y)), 0, opts.Height-1)
		if grid[yIdx][xIdx] == ' ' {
			grid[yIdx][xIdx] = '░'
		}
	}

	// Plot primary series (Net WPM) with solid character (overwrites secondary)
	for _, point := range primary {
		x := mapToRange(float64(point.TimeMs), 0, float64(maxTime), 0, float64(chartWidth-1))
		y := mapToRange(point.Value, minVal, maxVal, float64(opts.Height-1), 0)
		xIdx := clampInt(int(math.Round(x)), 0, chartWidth-1)
		yIdx := clampInt(int(math.Round(y)), 0, opts.Height-1)
		grid[yIdx][xIdx] = '█'
	}

	// Build output
	var sb strings.Builder

	if opts.Title != "" {
		sb.WriteString(opts.Title)
		sb.WriteRune('\n')
	}

	// Legend
	sb.WriteString("      █ WPM  ░ Raw WPM\n")

	// Render grid
	for row := 0; row < opts.Height; row++ {
		if opts.ShowAxis {
			val := mapToRange(float64(row), 0, float64(opts.Height-1), maxVal, minVal)
			if row == 0 || row == opts.Height-1 || row == opts.Height/2 {
				sb.WriteString(fmt.Sprintf("%5.0f│", val))
			} else {
				sb.WriteString("     │")
			}
		}
		sb.WriteString(string(grid[row]))
		sb.WriteRune('\n')
	}

	// X-axis
	if opts.ShowAxis {
		sb.WriteString("     └")
		sb.WriteString(strings.Repeat("─", chartWidth))
		sb.WriteRune('\n')

		sb.WriteString("     ")
		sb.WriteString("0s")
		midPadding := chartWidth/2 - 2
		if midPadding > 0 {
			sb.WriteString(strings.Repeat(" ", midPadding))
			sb.WriteString(fmt.Sprintf("%ds", maxTime/2000))
		}
		endPadding := chartWidth - midPadding - 6
		if endPadding > 0 {
			sb.WriteString(strings.Repeat(" ", endPadding))
			sb.WriteString(fmt.Sprintf("%ds", maxTime/1000))
		}
		sb.WriteRune('\n')
	}

	return sb.String()
}

// Helper functions

func findMinMax(points []DataPoint) (min, max float64) {
	if len(points) == 0 {
		return 0, 100
	}

	min = points[0].Value
	max = points[0].Value

	for _, p := range points {
		if p.Value < min {
			min = p.Value
		}
		if p.Value > max {
			max = p.Value
		}
	}

	return min, max
}

func mapToRange(value, inMin, inMax, outMin, outMax float64) float64 {
	if inMax-inMin == 0 {
		return outMin
	}
	return (value-inMin)/(inMax-inMin)*(outMax-outMin) + outMin
}

func clampInt(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// connectPoints draws lines between consecutive points
func connectPoints(grid [][]rune, points []DataPoint, minVal, maxVal float64, width, height int) {
	if len(points) < 2 {
		return
	}

	maxTime := float64(points[len(points)-1].TimeMs)
	if maxTime == 0 {
		maxTime = 1
	}

	for i := 0; i < len(points)-1; i++ {
		x1 := mapToRange(float64(points[i].TimeMs), 0, maxTime, 0, float64(width-1))
		y1 := mapToRange(points[i].Value, minVal, maxVal, float64(height-1), 0)
		x2 := mapToRange(float64(points[i+1].TimeMs), 0, maxTime, 0, float64(width-1))
		y2 := mapToRange(points[i+1].Value, minVal, maxVal, float64(height-1), 0)

		// Draw line between points using Bresenham-style iteration
		steps := int(math.Max(math.Abs(x2-x1), math.Abs(y2-y1)))
		if steps == 0 {
			continue
		}

		for s := 0; s <= steps; s++ {
			t := float64(s) / float64(steps)
			x := x1 + t*(x2-x1)
			y := y1 + t*(y2-y1)

			xIdx := clampInt(int(math.Round(x)), 0, width-1)
			yIdx := clampInt(int(math.Round(y)), 0, height-1)

			if grid[yIdx][xIdx] == ' ' {
				grid[yIdx][xIdx] = '·'
			}
		}
	}
}

// SparklineFromSamples creates a simple sparkline from samples
func SparklineFromSamples(samples []DataPoint, width int) string {
	if len(samples) == 0 {
		return ""
	}

	chars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	minVal, maxVal := findMinMax(samples)
	valRange := maxVal - minVal
	if valRange < 1 {
		valRange = 1
	}

	// Sample down to width if needed
	var result strings.Builder
	step := float64(len(samples)) / float64(width)
	if step < 1 {
		step = 1
	}

	for i := 0; i < width && int(float64(i)*step) < len(samples); i++ {
		idx := int(float64(i) * step)
		normalized := (samples[idx].Value - minVal) / valRange
		charIdx := int(normalized * float64(len(chars)-1))
		if charIdx < 0 {
			charIdx = 0
		}
		if charIdx >= len(chars) {
			charIdx = len(chars) - 1
		}
		result.WriteRune(chars[charIdx])
	}

	return result.String()
}


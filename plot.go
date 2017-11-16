// The package used here to produce plots can produce 'picture' files.
// Here, intermediate PNG files have been used and then embedded in a PDF.
package main

import (
	"flag"
	"fmt"
	"strconv"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"

	"github.com/jung-kurt/gofpdf"
	"github.com/montanaflynn/stats"
	"github.com/phil0lucas/GoForCP/CPUtils"
	"github.com/phil0lucas/GoForCP/SC"
	"github.com/phil0lucas/GoForCP/VS"
)

var infile1 = flag.String("s", "../CreateData/sc3.csv", "Name of SC input file")
var infile2 = flag.String("v", "../CreateData/vs3.csv", "Name of VS input file")
var outfile = flag.String("o", "plot01.pdf", "Name of output file")

// The graphics dimensions in the same ratio as an A4 landscape sheet
// allowing for title and footnote space.
var imgX = 207.0 // Image X size in mm
var imgY = 138.0 // Image Y size in mm

// Header and Footer structs
type headers struct {
	head1Left   string
	head1Right  string
	head2Left   string
	head2Right  string
	head3Left   string
	head4Centre string
	head5Centre string
	head6Centre string
}

type footers struct {
	foot1Left   string
	foot2Left   string
	foot3Left   string
	foot4Left   string
	foot4Centre string
	foot4Right  string
}

// Header and Footer text
// The output variables are pointers to structs holding the text.
func titles() *headers {
	h := &headers{
		head1Left:   "Acme Corp",
		head1Right:  "CONFIDENTIAL",
		head2Left:   "XYZ123 / Anti-Hypertensive",
		head2Right:  "Draft",
		head3Left:   "Protocol XYZ123",
		head4Centre: "Study XYZ123",
		head5Centre: "Blood Pressures by Visit and Treatment Arm",
		head6Centre: "All Randomized Subjects",
	}
	return h
}

// func footnotes(screened string, failures string) *footers{
func footnotes() *footers {
	f := &footers{
		foot1Left:   "Created with Go 1.8 for linux/amd64.",
		foot2Left:   "",
		foot3Left:   "Measurements were taken at 14 day intervals.",
		foot4Left:   "Page %d of {nb}",
		foot4Right:  "Run: " + CPUtils.TimeStamp(),
		foot4Centre: CPUtils.GetCurrentProgram(),
	}
	return f
}

// This provides a structure for summarizing the BPs per Arm (i.e teatment),
// Test and Visit
// One object per Arm-Vstestcd-Visitnum
type perAVV struct {
	Arm      string
	Vstestcd string
	Visitnum int
	Vsstresn float64
}

// Create a slice of perAVV objects.
// perAVV objects have a compound key Arm-Vstestcd-Visitnum and Vsstresn values to
// summarize into plottable points.
// The Arm (i.e. Treatment Group) is looked up from the input map, simulating the 'merge'
// 	of the data
func sMerge(vs []*VS.Vsrec, subjArm map[string]string) []perAVV {
	// 	Output slice of structs
	var vsp []perAVV
	for _, v := range vs {
		var p perAVV
		p.Arm = subjArm[v.Usubjid]
		p.Vstestcd = v.Vstestcd
		p.Visitnum = v.Visitnum

		// Note v.Vsstresn is a pointer; if nil the value is missing.
		// else p.Vsstresn will take its default zero value i.e. 0
		if v.Vsstresn != nil {
			p.Vsstresn = *v.Vsstresn
		}

		// Include those measures with non-missing Arm, Test and result and,
		// for this plot, only take the BP measures
		if p.Arm != "" && p.Vstestcd != "" && p.Vstestcd != "HR" && p.Vsstresn != 0 {
			vsp = append(vsp, p)
		}
	}
	return vsp
}

// In order to be able to calculate the mean BP for each Arm/Vstest/Visit the
// results values are placed in a slice attached to each key in a map.
type Key struct {
	Arm      string
	Vstestcd string
	Visitnum int
}

// For each value of the 'Key' create a slice of results.
// This is like transposing the data from long and thin to short and wide.
func transp(p []perAVV) map[Key][]float64 {
	// Output map
	m := make(map[Key][]float64)

	for _, v := range p {
		var k Key
		k.Arm = v.Arm
		k.Vstestcd = v.Vstestcd
		k.Visitnum = v.Visitnum
		m[k] = append(m[k], v.Vsstresn)
	}
	return m
}

// Use the structure from the transpose above to calculate the mean
type results struct {
	key  Key
	mean float64
}

func calcMean(m map[Key][]float64) []results {
	var sr []results
	for k, v := range m {
		var r results
		r.key = k
		r.mean, _ = stats.Mean(v)
		sr = append(sr, r)
	}
	return sr
}

// In the final output we need a graph per value of Arm
// ie one for Active and one for Placebo
type Graph struct {
	Arm string
}

// Each graph will have 2 lines - one each for SysBP & DiaBP
type Line struct {
	graph Graph
	line  string
}

// The plottable points need to be added to a plotter.XYs object which
// is itself a slice
func createPoints(sr []results) map[Line]plotter.XYs {
	m := make(map[Line]plotter.XYs)
	for _, v := range sr {
		var p Line
		p.graph.Arm = v.key.Arm
		p.line = v.key.Vstestcd
		m[p] = nil
	}

	// We know there are 15 visits to plot
	for k, _ := range m {
		m[k] = make(plotter.XYs, 15)
	}

	for _, v := range sr {
		var p Line
		p.graph.Arm = v.key.Arm
		p.line = v.key.Vstestcd
		for k, _ := range m {
			if k == p {
				index := v.key.Visitnum
				x := float64(v.key.Visitnum)
				y := v.mean
				m[k][index].X = x
				m[k][index].Y = y
			}
		}
	}
	return m
}

// Generate a slice of tick mark values to be added to each plot
func genTicks(min, max, interval float64) []plot.Tick {
	var t []plot.Tick
	for i := min; i < max; i += interval {
		position := i
		label := strconv.Itoa(int(i))
		s := plot.Tick{position, label}
		t = append(t, s)
	}
	return t
}

// Creation of each internal plot
func plotBP(pp map[Line]plotter.XYs, group string, n int, minY, maxY float64) string {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Treatment Group: " + group
	p.Y.Min = minY
	p.Y.Max = maxY
	p.X.Label.Text = "Visit Number"
	p.Y.Label.Text = "Blood Pressure (mmHg)"

	p.X.Tick.Marker = plot.ConstantTicks(genTicks(0, 14, 1))
	p.Y.Tick.Marker = plot.ConstantTicks(genTicks(minY, maxY, 10))

	err = plotutil.AddLinePoints(p,
		"Systolic BP", pp[Line{Graph{group}, "SBP"}],
		"Diastolic BP", pp[Line{Graph{group}, "DBP"}])
	if err != nil {
		panic(err)
	}

	// A temporary output PNG file for each plot
	t_out := "temp" + strconv.Itoa(n) + ".png"
	// Save the plot to a PNG file.
	if err := p.Save(vg.Length(imgX)*vg.Millimeter, vg.Length(imgY)*vg.Millimeter, t_out); err != nil {
		panic(err)
	}
	return t_out
}

// The overarching PDF that will include the PNG files.
func WriteReport(outputFile *string, h *headers, f *footers, g1 string, g2 string) error {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetHeaderFunc(func() {
		pdf.SetFont("Courier", "", 10)
		pdf.CellFormat(0, 10, (*h).head1Left, "0", 0, "L", false, 0, "")
		pdf.CellFormat(0, 10, (*h).head1Right, "0", 0, "R", false, 0, "")
		pdf.Ln(4)
		pdf.CellFormat(0, 10, (*h).head2Left, "0", 0, "L", false, 0, "")
		pdf.CellFormat(0, 10, (*h).head2Right, "0", 0, "R", false, 0, "")
		pdf.Ln(4)
		pdf.CellFormat(0, 10, (*h).head3Left, "0", 0, "L", false, 0, "")
		pdf.Ln(4)
		pdf.CellFormat(0, 10, (*h).head4Centre, "0", 0, "C", false, 0, "")
		pdf.Ln(4)
		pdf.CellFormat(0, 10, (*h).head5Centre, "0", 0, "C", false, 0, "")
		pdf.Ln(4)
		pdf.CellFormat(0, 10, (*h).head6Centre, "0", 0, "C", false, 0, "")
		pdf.Ln(10)
	})

	pdf.SetFooterFunc(func() {
		pdf.SetY(-30)
		pdf.SetFont("Courier", "", 10)
		pdf.CellFormat(0, 10, (*f).foot1Left, "0", 0, "L", false, 0, "")
		pdf.Ln(4)
		pdf.CellFormat(0, 10, (*f).foot2Left, "0", 0, "L", false, 0, "")
		pdf.Ln(4)
		pdf.CellFormat(0, 10, (*f).foot3Left, "0", 0, "L", false, 0, "")
		pdf.Ln(4)
		pdf.CellFormat(0, 10, fmt.Sprintf((*f).foot4Left, pdf.PageNo()), "", 0, "L", false, 0, "")
		pdf.SetX(40)
		pdf.CellFormat(0, 10, (*f).foot4Centre, "", 0, "L", false, 0, "")
		pdf.CellFormat(0, 10, (*f).foot4Right, "", 0, "R", false, 0, "")
	})
	pdf.AliasNbPages("")

	// 	AddPage() executes the generated Header and Footer functions
	y := pdf.GetY()
	pdf.AddPage()
	y = pdf.GetY()
	// Include the first graph
	pdf.Image(g1, 30, 40, imgX, imgY, false, "", 0, "")
	pdf.AddPage()
	// Include the second graph
	pdf.Image(g2, 30, 40, imgX, imgY, false, "", 0, "")

	// 	Output
	err := pdf.OutputFileAndClose(*outputFile)
	fmt.Println(err)
	return err
}

// Define the extreme values of the blood pressures in order to
// bound the y axis by the data range.
func MinMax(vsp []perAVV) (float64, float64) {
	var minY, maxY float64
	for _, v := range vsp {
		if v.Vsstresn > maxY {
			maxY = v.Vsstresn
		}
		if v.Vsstresn < minY || minY == 0 {
			minY = v.Vsstresn
		}
	}
	// Round the min down to the nearest mutiple of 10
	m := minY / 10
	m2 := int64(m)
	minY = float64(m2) * 10

	// Round the max up to the nearest mutiple of 10
	m = maxY / 10
	m2 = int64(m) + 1
	maxY = float64(m2) * 10

	return minY, maxY
}

func main() {
	// Read the 'SC' data and dump into the slice of structs
	sc := SC.ReadSC(infile1)

	// Create a map of Usubjid as key and Arm as value
	// Screening failures are eliminated from the lookup map.
	subjArm := make(map[string]string)
	for _, v := range sc {
		if v.Arm != nil {
			subjArm[v.Usubjid] = *v.Arm
		}
	}

	// Read the VS data
	vs := VS.ReadVS(infile2)

	// Create a slice of Point objects.
	// Point objects have a compound key Arm-Vstestcd-Visitnum and Vsstresn values to
	// summarize into plottable points
	vsp := sMerge(vs, subjArm)

	// Determine minimum and maximum BP measures for setting the Y axis
	minY, maxY := MinMax(vsp)

	// Transpose from a 'record' per Arm-Vstestcd-Visitnum-Vsstresn to one
	// per Arm-Vstestcd-Visitnum, aligning the Vsstresn into slices in
	// order to calculate the mean Vsstresn.
	pMap := transp(vsp)

	// Calculate the mean of the Vsstresn for each Arm-Vstestcd-Visitnum
	meanVals := calcMean(pMap)

	// Add the data into the objects needed to plot the points
	pp := createPoints(meanVals)

	// Create a plot for each treatment group (Arm)
	g1 := plotBP(pp, "Placebo", 1, minY, maxY)
	g2 := plotBP(pp, "Active", 2, minY, maxY)

	// 	Report
	h := titles()
	f := footnotes()

	err := WriteReport(outfile, h, f, g1, g2)
	if err != nil {
		fmt.Println(err)
	}
}

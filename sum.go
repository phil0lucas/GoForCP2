// A simple summary of demographic data for randomized subjects.
// The gofpdf package is used to create the output PDF.
package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/montanaflynn/stats"
	"github.com/phil0lucas/GoForCP/CPUtils"
	"github.com/phil0lucas/GoForCP/DM"
)

// Input and output files
var infile = flag.String("i", "../CreateData/dm3.csv", "Name of input file")
var outfile = flag.String("o", "summary15.pdf", "Name of output file")

// Define header structure
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

// Footer structure
type footers struct {
	foot1Left   string
	foot2Left   string
	foot3Left   string
	foot4Left   string
	foot4Centre string
	foot4Right  string
}

// Add values to the header struct and create a pointer to them
func titles() *headers {
	h := &headers{
		head1Left:   "Acme Corp",
		head1Right:  "CONFIDENTIAL",
		head2Left:   "XYZ123 / Anti-Hypertensive",
		head2Right:  "Draft",
		head3Left:   "Protocol XYZ123",
		head4Centre: "Study XYZ123",
		head5Centre: "Summary of Demographic Data by Treatment Arm",
		head6Centre: "All Randomized Subjects",
	}
	return h
}

// Footer as per header with added substituted values
func footnotes(screened string, failures string) *footers {
	f2 := "Of the original " + screened + " screened subjects, " +
		failures + " were excluded at Screening and are not counted."
	f := &footers{
		foot1Left:   "Created with Go 1.8 for linux/amd64.",
		foot2Left:   f2,
		foot3Left:   "All measurements were taken at the screening visit.",
		foot4Left:   "Page %d of {nb}",
		foot4Right:  "Run: " + CPUtils.TimeStamp(),
		foot4Centre: CPUtils.GetCurrentProgram(),
	}
	return f
}

// Define a slice of strings being the treatment group column headers i.e. Active, Placebo, Overall
func selectTGs(m map[string]int) []string {
	var s []string
	for k, _ := range m {
		if k != "SF" && k != "Screened" {
			s = append(s, k)
		}
	}
	return s
}

func pad(i int, width int) string {
	c_num := strconv.Itoa(i)
	spaces := width - len(c_num)
	c_num = strings.Repeat(" ", spaces) + c_num
	return c_num
}

func cpad(s string, width int) string {
	spaces := width - len(s)
	outstr := strings.Repeat(" ", spaces) + s
	return outstr
}

// Report
func WriteReport(outputFile *string, h *headers, f *footers,
	nTG map[string]int,
	nAge map[string]string,
	meansd map[string]string,
	median map[string]string,
	min map[string]string,
	max map[string]string,
	sexPct map[Key]string,
	racePct map[KeyR]string) error {

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
	pdf.AddPage()

	// 	Column headers
	colHeaderSlice := []string{"Characteristic", "Statistic", "Placebo", "Active", "Overall"}
	colWidthSlice := []float64{60, 60, 50, 50, 50}
	colJustSlice := []string{"L", "L", "L", "L", "L"}
	for i, str := range colHeaderSlice {
		pdf.CellFormat(colWidthSlice[i], 8, str, "TB", 0, colJustSlice[i], false, 0, "")
	}
	pdf.Ln(8)

	//	Number of Subjects By TG
	textSlice := []string{"Number of Subjects", "N", pad(nTG["Placebo"], 3),
		pad(nTG["Active"], 3), pad(nTG["Overall"], 3)}
	for i, str := range textSlice {
		pdf.CellFormat(colWidthSlice[i], 8, str, "", 0, colJustSlice[i], false, 0, "")
	}
	pdf.Ln(8)

	//	Number of non-missing Ages By TG
	textSlice2 := []string{"Age (years)", "Number of Non-Missing", nAge["Placebo"], nAge["Active"], nAge["Overall"]}
	for i, str := range textSlice2 {
		pdf.CellFormat(colWidthSlice[i], 8, str, "", 0, colJustSlice[i], false, 0, "")
	}
	pdf.Ln(4)

	// 	Mean and Standard Deviation by TG
	textSlice3 := []string{" ", "Mean (SD)",
		meansd["Placebo"],
		meansd["Active"],
		meansd["Overall"]}
	for i, str := range textSlice3 {
		pdf.CellFormat(colWidthSlice[i], 8, str, "", 0, colJustSlice[i], false, 0, "")
	}
	pdf.Ln(4)

	//  Median
	textSlice4 := []string{" ", "Median",
		median["Placebo"],
		median["Active"],
		median["Overall"]}
	for i, str := range textSlice4 {
		pdf.CellFormat(colWidthSlice[i], 8, str, "", 0, colJustSlice[i], false, 0, "")
	}
	pdf.Ln(4)

	//  Minimum
	textSlice5 := []string{" ", "Minimum",
		min["Placebo"],
		min["Active"],
		min["Overall"]}
	for i, str := range textSlice5 {
		pdf.CellFormat(colWidthSlice[i], 8, str, "", 0, colJustSlice[i], false, 0, "")
	}
	pdf.Ln(4)

	//  Maximum
	textSlice6 := []string{" ", "Maximum",
		max["Placebo"],
		max["Active"],
		max["Overall"]}
	for i, str := range textSlice6 {
		pdf.CellFormat(colWidthSlice[i], 8, str, "", 0, colJustSlice[i], false, 0, "")
	}
	pdf.Ln(8)

	//	Gender
	uV := uniqueValues(sexPct)
	var iter int
	var col1text string
	sexFmt := map[string]string{"F": "Female", "M": "Male"}
	for _, v := range uV {
		if iter == 0 {
			col1text = "Gender"
		} else {
			col1text = ""
		}
		textSlice7 := []string{col1text, sexFmt[v],
			sexPct[Key{v, "Placebo"}],
			sexPct[Key{v, "Active"}],
			sexPct[Key{v, "Overall"}]}
		for i, str := range textSlice7 {
			pdf.CellFormat(colWidthSlice[i], 8, str, "", 0, colJustSlice[i], false, 0, "")
		}
		pdf.Ln(4)
		iter++
	}

	//	Race
	pdf.Ln(4)
	uVr := uniqueValuesR(racePct)
	var iterR int
	var col1textR string
	for _, v := range uVr {
		if iterR == 0 {
			col1textR = "Race"
		} else {
			col1textR = ""
		}
		textSlice8 := []string{col1textR, v,
			racePct[KeyR{v, "Placebo"}],
			racePct[KeyR{v, "Active"}],
			racePct[KeyR{v, "Overall"}]}
		for i, str := range textSlice8 {
			pdf.CellFormat(colWidthSlice[i], 8, str, "", 0, colJustSlice[i], false, 0, "")
		}
		pdf.Ln(4)
		iterR++
	}

	//	Underline
	pdf.SetY(-36)
	colUnderSlice := []string{" ", " ", " ", " ", " "}
	for i, str := range colUnderSlice {
		pdf.CellFormat(colWidthSlice[i], 8, str, "B", 0, colJustSlice[i], false, 0, "")
	}

	// 	Output
	err := pdf.OutputFileAndClose(*outputFile)
	return err
}

// Count the number of non-missing Age values per treatment group
func nMiss(dm []*DM.Dmrec) map[string]string {
	m := make(map[string]int)
	for _, v := range dm {
		if v.Age != nil {
			m[*v.Arm]++
			m["Overall"]++
		} else {
			m["Missing"]++
		}
	}

	mm := make(map[string]string)
	for k, v := range m {
		mm[k] = pad(v, 3)
	}
	return mm
}

// Here, the input is 'transposed' into a map where the keys are the treatment group values
// and the values are arrays of non-missing Age values. In this way the stats can easily
// be calculated for each treatment group
func prepareData(dm []*DM.Dmrec, tg []string) map[string][]float64 {
	m := make(map[string][]float64)
	for _, s := range tg {
		var out []float64
		for _, v := range dm {
			if v.Age != nil {
				if s == *v.Arm {
					out = append(out, float64(*v.Age))
				} else if s == "Overall" {
					out = append(out, float64(*v.Age))
				}
			}
		}
		m[s] = out
	}
	return m
}

// Calculate stats on each input slice in the map from the function above
func mStat(indata map[string][]float64, stat string, dec int, width int) map[string]string {
	m := make(map[string]string)
	for i, v := range indata {
		var result float64
		if stat == "Mean" {
			result, _ = stats.Mean(v)
		} else if stat == "SD" {
			result, _ = stats.StandardDeviationPopulation(v)
		} else if stat == "Median" {
			result, _ = stats.Median(v)
		} else if stat == "Min" {
			result, _ = stats.Min(v)
		} else if stat == "Max" {
			result, _ = stats.Max(v)
		}
		c_stat := strconv.FormatFloat(result, 'f', dec, 64)
		if len(c_stat) < width {
			c_stat = cpad(c_stat, width)
		}
		m[i] = c_stat
	}
	return m
}

// Using an object like this as a map key lets it act like a compound key
type Key struct {
	sex string
	arm string
}

// Count each non-missing value of Sex by Treatment Group (Arm)
func countSexByTG(dm []*DM.Dmrec) map[Key]int {
	var r []Key
	for _, v := range dm {
		var k Key
		if v.Sex != nil {
			k.sex = *v.Sex
			k.arm = *v.Arm
			r = append(r, k)
			k.arm = "Overall"
			r = append(r, k)
		}
	}

	// calculate sum:
	m := make(map[Key]int)
	for _, v := range r {
		m[v]++
	}

	return m
}

// Determine percentage of the N values of the Treatment Group (Arm)
func pctSexByTG(m map[Key]int, tg map[string]int) map[Key]string {
	outmap := make(map[Key]string)
	for k, v := range m {
		var pct float64
		pct = (float64(v) / float64(tg[k.arm])) * 100
		c_pct := strconv.FormatFloat(pct, 'f', 2, 64)
		c_stat := strconv.FormatInt(int64(v), 10)
		spaces := 3 - len(c_stat)
		c_stat = strings.Repeat(" ", spaces) + c_stat + " (" + c_pct + "%)"
		outmap[k] = c_stat
	}
	return outmap
}

//	Determine the unique values of the non-TG key
func uniqueValues(m map[Key]string) []string {
	var uValues []string
	for k, _ := range m {
		if !CPUtils.StringInSlice(k.sex, uValues) {
			uValues = append(uValues, k.sex)
		}
	}
	return uValues
}

// Define the composite key to do counts of each Race value by Treatment Group (Arm)
type KeyR struct {
	race string
	arm  string
}

func countRaceByTG(dm []*DM.Dmrec) map[KeyR]int {
	var r []KeyR
	for _, v := range dm {
		var k KeyR
		if v.Race != nil {
			k.race = *v.Race
			k.arm = *v.Arm
			r = append(r, k)
			k.arm = "Overall"
			r = append(r, k)
		}
	}
	// calculate sum:
	m := make(map[KeyR]int)
	for _, v := range r {
		m[v]++
	}
	return m
}

//	Determine the unique values of the non-TG key
func uniqueValuesR(m map[KeyR]string) []string {
	var uValues []string
	for k, _ := range m {
		if !CPUtils.StringInSlice(k.race, uValues) {
			uValues = append(uValues, k.race)
		}
	}
	return uValues
}

// Each count of Race Value as percentage of the N for the Treatment Group (Arm)
func pctRaceByTG(m map[KeyR]int, tg map[string]int) map[KeyR]string {
	outmap := make(map[KeyR]string)
	for k, v := range m {
		var pct float64
		pct = (float64(v) / float64(tg[k.arm])) * 100
		c_pct := strconv.FormatFloat(pct, 'f', 2, 64)
		c_stat := strconv.FormatInt(int64(v), 10)
		spaces := 3 - len(c_stat)

		c_stat = strings.Repeat(" ", spaces) + c_stat
		c_stat = c_stat + " (" + c_pct + "%)"
		outmap[k] = c_stat
	}
	return outmap
}

func main() {
	// Read the file and dump into the slice of structs
	dm := DM.ReadDM(infile)

	// 	Compute number of subjects by treatment group
	nTG := DM.CountByTG(dm)

	// Select treatment groups to display i.e. Placebo, Active, Overall
	TGs := selectTGs(nTG)

	// Create version of dm without the SFs
	dm2 := DM.RemoveSF(dm)

	// 	Compute number of non-missing Age values by TG
	nAge := nMiss(dm2)

	//	Prepare the data for passing to the stats functions
	//	Select only the TGs to display
	//	Remove the missing values.
	rMiss := prepareData(dm2, TGs)

	// 	Compute stats of age by TG
	mean := mStat(rMiss, "Mean", 2, 6)

	sd := mStat(rMiss, "SD", 2, 5)

	//	Concatenate mean and SD values into a display string
	meansd := make(map[string]string)
	for k, _ := range mean {
		meansd[k] = mean[k] + " (" + sd[k] + ")"
	}

	median := mStat(rMiss, "Median", 0, 3)
	min := mStat(rMiss, "Min", 0, 3)
	max := mStat(rMiss, "Max", 0, 3)

	//  N and % of subjects by gender and TG
	keyValues := countSexByTG(dm2)
	pctMap := pctSexByTG(keyValues, nTG)

	//  N and % of subjects by race and TG
	raceValues := countRaceByTG(dm2)
	pctRace := pctRaceByTG(raceValues, nTG)

	// 	Report
	h := titles()
	f_scr := strconv.Itoa(nTG["Screened"])
	f_sf := strconv.Itoa(nTG["SF"])
	f := footnotes(f_scr, f_sf)
	err := WriteReport(outfile, h, f, nTG, nAge, meansd, median, min, max, pctMap, pctRace)
	if err != nil {
		fmt.Println(err)
	}
}

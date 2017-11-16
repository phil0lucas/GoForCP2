// Program to generate SDTM data for a fictitious study.
// Domain VS
// Metadata :
// - STUDYID 	Char 6  (constant) Study Identifier
// - DOMAIN  	Char 2  (constant) Domain abbreviation
// - USUBJID 	Char 18 STUDYID-SITEID-SUBJID Unique Subject Identifier (Key variable 1)
// - SUBJID  	Char 6  Subject Identifier
// - SITEID  	Char 4  Site Identifier
// - VSSEQ   	Num	 	Sequence number (Key variable 2)
// - VISITNUM	Num     Visit number (0=Screening, 1-14=Dosing visits and assessments)
// - VSTESTCD	Char 3  Test code
// - VSTEST		Char 30 Test description
// - VSORRES	Num 	Original recorded result
// - VSORRESU   Char    Units of original result
// - VSSTRESC   Char	Standardized result in char form
// - VSSTRESN   Num     Standardized result in numeric form
// - VSSTRESU   Char    Units of result in standardized form
// - VSBLFL     Char    Flags baseline visit
// - VSDTC      Date    Date of visit in ISO8601
// - VSDY    	Num     Study Day of collection

package VS

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/phil0lucas/GoForCP/CPUtils"
)

// This will mirror the metadata above with more natural types.
// The elements modelled as pointers may have missing values i.e.
// a nil pointer
type Vsrec struct {
	Studyid  string
	Domain   string
	Usubjid  string
	Subjid   string
	Siteid   string
	Vsseq    int
	Visitnum int
	Vstestcd string
	Vstest   string
	Vsorres  *float64
	Vsorresu *string
	Vsstresc *string
	Vsstresn *float64
	Vsstresu *string
	Vsblfl   bool
	Vsdtc    time.Time
	Vsdy     int
}

//	The type vsrecs models a 'data set' as a slice of
//	pointers to vsrec structs.
type vsrecs []*Vsrec

// The program will be run with flags to specify the input & output files
var testcodes = []string{"SBP", "DBP", "HR"}
var testnames = []string{"Systolic Blood Pressure", "Diastolic Blood Pressure", "Heart Rate"}

const (
	domain = "VS"
)

// Return a random integer in the specified range
func randValue(max, min int) float64 {
	rand.Seed(time.Now().UTC().UnixNano())
	return float64(rand.Intn(max-min) + min)
}

// Generate random baseline values for each test performed
func genBaseline(tcode string) float64 {
	switch tcode {
	// return rand.Intn(max - min) + min
	case "HR":
		return randValue(120, 70)
	case "SBP":
		return randValue(160, 120)
	case "DBP":
		return randValue(120, 90)
	}
	return 0.0
}

// Generate random resukts based on the subjects randomized treatment (Arm),
// baseline value and how far they are into the trial (visit number)
func getOrigRes(baseline float64, visitnum int, armcd *int) *float64 {
	if armcd == nil {
		return nil
	} else {
		if *armcd == 0 {
			if visitnum == 0 {
				return &baseline
			} else {
				v := baseline + randValue(5, -5)
				return &v
			}
		} else {
			if visitnum == 0 {
				return &baseline
			} else if visitnum < 5 {
				v := baseline*0.975 + randValue(2, -3)
				return &v
			} else if visitnum < 8 {
				v := baseline*0.95 + randValue(1, -5)
				return &v
			} else if visitnum < 11 {
				v := baseline*0.925 + randValue(0, -7)
				return &v
			} else {
				v := baseline*0.9 + randValue(-3, -10)
				return &v
			}
		}
	}
}

// Blood pressures in mm of Mercury, heart rate in beats per minute
func getUnits(testcode string) (string, string) {
	if testcode == "HR" {
		return "bpm", "bpm"
	} else if testcode == "SBP" || testcode == "DBP" {
		return "mmHg", "mmHg"
	} else {
		return "", ""
	}
}

// Len, Swap and Less are required for the Sort Interface. An interface
// defines a 'contract' i.e. a set of methods a type will possess.
// In this case it allows the structs to be re-arranged in the slice according
// the compound key of Usubjid-Vstestcd-Visitnum and thus generate VSSEQ as per SDTM
// principles.
func (t vsrecs) Len() int {
	return len(t)
}

func (t vsrecs) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t vsrecs) Less(i, j int) bool {
	if t[i].Usubjid < t[j].Usubjid {
		return true
	}
	if t[i].Usubjid > t[j].Usubjid {
		return false
	}
	// If USUBJIDs are equal
	if t[i].Vstestcd < t[j].Vstestcd {
		return true
	}
	if t[i].Vstestcd > t[j].Vstestcd {
		return false
	}
	return t[i].Visitnum < t[j].Visitnum
}

// Allocates test codes and their description
func tcodes(vstcd []string, vstdesc []string, index int, rtype int) (string, string) {
	if rtype > 0 {
		return vstcd[index], vstdesc[index]
	} else {
		return "", ""
	}
}

// Flags the baseline visit.
func flagBline(visit int) bool {
	if visit == 1 {
		return true
	} else {
		return false
	}
}

// Writes the generated data to a CSV correctky sorted by Usubjid-Vstestcd-Visitnum
func WriteVS(infile, outfile *string) {
	// open the file and pass it to a Scanner object
	file, err := os.Open(*infile)
	if err != nil {
		panic(fmt.Sprintf("error opening %s: %v", infile, err))
	}
	defer file.Close()

	// Output slice of pointers to structs
	var vs vsrecs

	// Pass the opened file to a scanner
	scanner := bufio.NewScanner(file)

	// For each subject
	for i := 0; scanner.Scan(); i++ {
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "error reading from file:", err)
			os.Exit(3)
		}
		str := scanner.Text()
		studyid := strings.Split(str, ",")[0]
		usubjid := strings.Split(str, ",")[3]
		subjid := strings.Split(str, ",")[1]
		siteid := strings.Split(str, ",")[2]
		rectype, _ := strconv.Atoi(strings.Split(str, ",")[4])
		dmdtc, _ := time.Parse("2006-01-02", strings.Split(str, ",")[5])
		endv := strings.Split(str, ",")[6]
		endvn, _ := strconv.Atoi(endv)

		// 		The ARMCD will be needed to create the data but will
		// 		not be included in the final data set. Recall this will
		// 		be a pointer to an int
		armcd := CPUtils.Str2IntP(strings.Split(str, ",")[9])
		// 		fmt.Printf("Study=%s Subject=%s Subjid=%s Siteid=%s\n", studyid, usubjid, subjid, siteid)
		// 		fmt.Printf("%v %v %v \n", dmdtc, endv, endvn)
		// 		CPUtils.PrintIntP(armcd)

		// Add in the visits up to the generated end-visit
		// Subjects with just visit 0 are screening failures.
		// Subjects with a final visit number < 14 are withdrawers.

		// Test codes
		for j := 0; j < len(testcodes); j++ {
			vstestcd, vstest := tcodes(testcodes, testnames, j, rectype)
			// 			fmt.Printf("Testcode=%s Test=%s\n", vstestcd, vstest)

			baseline := genBaseline(testcodes[j])
			// 			fmt.Printf("Test code %s value %v\n", testcodes[j], baseline)

			vsorresu, vsstresu := getUnits(vstestcd)
			// 			fmt.Printf("   Test code units %s, %s\n", vsorresu, vsstresu)

			// Visits
			for k := 0; k <= endvn; k++ {
				vsblfl := flagBline(k)
				// 				fmt.Println(vsblfl)
				// Recall ARMCD is now a pointer to an int.
				// VSORRES is a pointer to a float64, nil being a missing value
				vsorres := getOrigRes(baseline, k, armcd)
				// 				CPUtils.PrintFloatP(vsorres)
				vsdtc := dmdtc.AddDate(0, 0, (k * 14))
				vsdy := k * 14

				vs = append(vs, &Vsrec{
					Studyid:  studyid,
					Domain:   domain,
					Usubjid:  usubjid,
					Subjid:   subjid,
					Siteid:   siteid,
					Visitnum: k,
					Vstestcd: vstestcd,
					Vstest:   vstest,
					Vsorres:  vsorres,
					Vsstresn: vsorres,
					Vsstresc: CPUtils.FloatP2StrP(vsorres, 2),
					Vsorresu: &vsorresu,
					Vsstresu: &vsstresu,
					Vsblfl:   vsblfl,
					Vsdtc:    vsdtc,
					Vsdy:     vsdy,
				})

			} // End k loop
		} //	End j loop
	} // End i loop

	// Sort the struct of VS 'records'
	// Note the usage of the Sort interface
	sort.Sort(vsrecs(vs))

	// Define VSSEQ as key as running int within each subject
	// Need to define a variable external to the loop otherwise
	// scope will make each value 1
	var count int
	for ii := 0; ii < len(vs); ii++ {
		if ii == 0 || (vs[ii].Usubjid != vs[ii-1].Usubjid) {
			count = 0
		}
		count++
		vs[ii].Vsseq = count
	}

	// Write to external file.
	fo, err := os.Create(*outfile)
	if err != nil {
		log.Fatal(err)
	}
	defer fo.Close()

	// Create a buffered writer from the file
	w := bufio.NewWriter(fo)

	for ii, _ := range vs {
		bytesWritten, err := w.WriteString(
			vs[ii].Studyid + "," +
				vs[ii].Domain + "," +
				vs[ii].Subjid + "," +
				vs[ii].Siteid + "," +
				vs[ii].Usubjid + "," +
				strconv.Itoa(vs[ii].Vsseq) + "," +
				strconv.Itoa(vs[ii].Visitnum) + "," +
				vs[ii].Vstestcd + "," +
				vs[ii].Vstest + "," +
				CPUtils.FloatP2Str(vs[ii].Vsorres, 1) + "," +
				CPUtils.FloatP2Str(vs[ii].Vsstresn, 1) + "," +
				CPUtils.StrP2Str(vs[ii].Vsstresc) + "," +
				CPUtils.StrP2Str(vs[ii].Vsorresu) + "," +
				CPUtils.StrP2Str(vs[ii].Vsstresu) + "," +
				strconv.FormatBool(vs[ii].Vsblfl) + "," +
				vs[ii].Vsdtc.Format("2006-01-02") + "," +
				strconv.Itoa(vs[ii].Vsdy) +
				"\n")

		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Bytes written: %d\n", bytesWritten)
	}

	// Write to disk
	w.Flush()
}

// This reads the CSV into the same slice of pointers to structs
func ReadVS(infile *string) []*Vsrec {
	// open the file and pass it to a Scanner object
	file, err := os.Open(*infile)
	if err != nil {
		panic(fmt.Sprintf("error opening %s: %v", *infile, err))
	}
	defer file.Close()

	// Pass the opened file to a scanner
	scanner := bufio.NewScanner(file)

	var vsx []*Vsrec
	for i := 0; scanner.Scan(); i++ {
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "error reading from file:", err)
			os.Exit(3)
		}
		str := scanner.Text()
		studyid := strings.Split(str, ",")[0]
		domain := strings.Split(str, ",")[1]
		subjid := strings.Split(str, ",")[2]
		siteid := strings.Split(str, ",")[3]
		usubjid := strings.Split(str, ",")[4]
		vsseq, _ := strconv.Atoi(strings.Split(str, ",")[5])
		vnum, _ := strconv.Atoi(strings.Split(str, ",")[6])
		vstestcd := strings.Split(str, ",")[7]
		vstest := strings.Split(str, ",")[8]
		vsorres := CPUtils.Str2FloatP(strings.Split(str, ",")[9])
		vsstresn := CPUtils.Str2FloatP(strings.Split(str, ",")[10])
		vsstresc := CPUtils.Str2StrP(strings.Split(str, ",")[11])
		vsorresu := CPUtils.Str2StrP(strings.Split(str, ",")[12])
		vsstresu := CPUtils.Str2StrP(strings.Split(str, ",")[13])
		vsblfl, _ := strconv.ParseBool(strings.Split(str, ",")[14])
		vsdtc, _ := time.Parse("2006-01-02", strings.Split(str, ",")[15])
		vsdy, _ := strconv.Atoi(strings.Split(str, ",")[16])

		vsx = append(vsx, &Vsrec{
			Studyid:  studyid,
			Domain:   domain,
			Subjid:   subjid,
			Siteid:   siteid,
			Usubjid:  usubjid,
			Vsseq:    vsseq,
			Visitnum: vnum,
			Vstestcd: vstestcd,
			Vstest:   vstest,
			Vsorres:  vsorres,
			Vsstresn: vsstresn,
			Vsstresc: vsstresc,
			Vsorresu: vsorresu,
			Vsstresu: vsstresu,
			Vsblfl:   vsblfl,
			Vsdtc:    vsdtc,
			Vsdy:     vsdy,
		})
	}
	return vsx
}

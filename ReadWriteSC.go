// Program to generate SDTM data for a fictitious study.
//
// This is not meant to represent an SC domain, but just be a framework upon which DM and VS can be built,
// so this data is used as a template for DM and VS, ensuring they are both internally consistent.
//
// Structure of the study:
// - Choose 100 subjects allocated to 5 sites.
// - Recruitment between 01Jan2010 and 31Dec2010 (any date in that window with no allowance for weekends, public holidays etc.)
// - A random 5% of the population will be screening failures. (RECTYPE=0)
// - Of the remainder, 35% will withdraw at any time after their start (RECTYPE=1).
// - 60% will last the 15 visits of the study (RECTYPE=2)
// - The full course of the study will be fortnightly visits for a maximum of 28 weeks
// - For simplicity, withdrawal is assumed at a scheduled visit, no unscheduled visits will be considered.
// - Screening (demog data) will be visit 0; subsequent visits (VS data) will be 1, 2, 3 etc to a maximum of 14
// Metadata:
// - STUDYID Char 6 (constant) Study Identifier
// - USUBJID Char 18 STUDYID-SITEID-SUBJID Unique Subject Identifier
// - SUBJID  Char 6 Subject Identifier
// - SITEID  Char 4 Site Identifier
// - RFSTDTC Char ISO8601 First date of study med exposure
// - RFENDTC Char ISO8601 Last date of study med exposure
// - DMDTC   Char ISO8601 Date/Time of Collection
// - RECTYPE Num  0=SF, 1=WD, 2=Completer
// - ENDV    Num  Last visit attended in study. RECTYPE=0 records will have 0 for this.
// - ARMCD   Num     Treatment Arm code
// - ARM     Char 7  Treatment Arm

package SC

import (
	"bufio"
	// 	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/phil0lucas/GoForCP/CPUtils"
)

// Some variables can have missing values, so they are modelled by a pointer.
// In the case of an MV the value of the pointer address is nil.
type Subject struct {
	Studyid string
	Subjid  string
	Siteid  string
	Usubjid string
	Rectype int
	Dmdtc   time.Time
	Endv    int
	Rfstdtc *time.Time
	Rfendtc *time.Time
	Armcd   *int
	Arm     *string
}

// Constants can only be numbers, strings or boolean
const (
	studyid   = "XYZ123" // Study Identifier
	nSubj     = 100      // Number of subjects in the study
	lastVisit = 14
)

// The study can start at any date within 2010
var baseDate = time.Date(2010, time.January, 1, 0, 0, 0, 0, time.UTC)

// The SITEID is chosen from one of these 5 values
var siteids = []string{"1", "2", "3", "4", "5"}

// To allow a random choice of arm
var arm = map[int]string{0: "Placebo", 1: "Active"}

// Use this to flag subjects as
// - screening failures (~5%)
// - withdrawers (~35%)
// - completers (~60%)
func ptype() int {
	rand.Seed(time.Now().UTC().UnixNano())
	x := rand.Float64()
	switch {
	case x <= 0.05:
		return 0
	case x > 0.05 && x < 0.4:
		return 1
	default:
		return 2
	}
}

// For each subject randomly select their last visit
// depending upon whether they are withdrawers, completers or screening failures.
func endv(r int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	switch r {
	case 0:
		return 0
	case 1:
		// Because the discontinuing patient has been dosed,
		// cannot finish at visit 0
		// So choose randomly from 0 to 13 and then add 1
		return (rand.Intn(lastVisit - 1)) + 1
	default:
		return lastVisit
	}
}

// Construct a reference start date based on the record type.
// This may be missing, so a pointer type is used.
func startDate(r int, d time.Time) *time.Time {
	switch r {
	case 0:
		return nil
	default:
		d2 := d.AddDate(0, 0, 14)
		return &d2
	}
}

// Construct an end date dependent upon the last visit
func endDate(r int, e int, d time.Time) *time.Time {
	switch r {
	case 0:
		return nil
	case 1:
		d2 := d.AddDate(0, 0, (e * 14))
		return &d2
	default:
		d2 := d.AddDate(0, 0, (lastVisit * 14))
		return &d2
	}
}

//	Randomly select the treatment arm and its code
func getArm(r int) (*int, *string) {
	rand.Seed(time.Now().UTC().UnixNano())
	if r != 0 {
		armcd := rand.Intn(len(arm))
		arm := arm[armcd]
		return &armcd, &arm
	} else {
		return nil, nil
	}
}

//	Create a CSV file of a row per subject
func WriteSC(f *string) {
	// Create slice of pointers to Subject types
	sSubj := make([]*Subject, nSubj)

	for ii := 0; ii < nSubj; ii++ {
		subjid := CPUtils.LeftPad2Len(strconv.Itoa(ii+1), "0", 6)
		siteid := CPUtils.LeftPad2Len(CPUtils.Choice(siteids), "0", 4)
		usubjsl := []string{studyid, siteid, subjid}
		usubjid := strings.Join(usubjsl, "-")
		rectype := ptype()
		dmdtc := baseDate.AddDate(0, 0, rand.Intn(364))
		endv := endv(rectype)
		rfstdtc := startDate(rectype, dmdtc)
		rfendtc := endDate(rectype, endv, dmdtc)
		armcd, arm := getArm(rectype)

		// Add the address of the struct into the slice
		sSubj[ii] = &Subject{
			studyid,
			subjid,
			siteid,
			usubjid,
			rectype,
			dmdtc,
			endv,
			rfstdtc,
			rfendtc,
			armcd,
			arm,
		}
	}

	// Output to external file via strings
	fo, err := os.Create(*f)
	if err != nil {
		log.Fatal(err)
	}
	defer fo.Close()

	//  Create a buffered writer to the file
	w := bufio.NewWriter(fo)

	// For each subject write a row
	for ii := 0; ii < nSubj; ii++ {
		bytesWritten, err := w.WriteString(
			sSubj[ii].Studyid + "," +
				sSubj[ii].Subjid + "," +
				sSubj[ii].Siteid + "," +
				sSubj[ii].Usubjid + "," +
				strconv.Itoa(sSubj[ii].Rectype) + "," +
				sSubj[ii].Dmdtc.Format("2006-01-02") + "," +
				strconv.Itoa(sSubj[ii].Endv) + "," +
				CPUtils.DateP2Str(sSubj[ii].Rfstdtc) + "," +
				CPUtils.DateP2Str(sSubj[ii].Rfendtc) + "," +
				CPUtils.IntP2Str(sSubj[ii].Armcd) + "," +
				CPUtils.StrP2Str(sSubj[ii].Arm) +
				"\n")

		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Bytes written: %d\n", bytesWritten)
	}
	// 	Write to disk
	w.Flush()
}

//	Read the CSV into the same struct as used to write it
func ReadSC(infile *string) []*Subject {
	// open the file and pass it to a Scanner object
	file, err := os.Open(*infile)
	if err != nil {
		panic(fmt.Sprintf("error opening %s: %v", *infile, err))
	}
	defer file.Close()

	// Pass the opened file to a scanner
	scanner := bufio.NewScanner(file)

	var subj []*Subject
	for i := 0; scanner.Scan(); i++ {
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "error reading from file:", err)
			os.Exit(3)
		}
		str := scanner.Text()

		// Split up each row from the CSV
		studyid := strings.Split(str, ",")[0]
		subjid := strings.Split(str, ",")[1]
		siteid := strings.Split(str, ",")[2]
		usubjid := strings.Split(str, ",")[3]
		rectype, _ := strconv.Atoi(strings.Split(str, ",")[4])

		// Screening date
		dmdtc, _ := time.Parse("2006-01-02", strings.Split(str, ",")[5])

		// Last visit number
		endv, _ := strconv.Atoi(strings.Split(str, ",")[6])

		// First date of dosing for randomized subjects.
		// For screening failures this will be a nil pointer.
		rfstdtc := CPUtils.Str2DateP(strings.Split(str, ",")[7])

		//	Last day of dosing.
		//	This will also be missing if the subject is a screening failure
		rfendtc := CPUtils.Str2DateP(strings.Split(str, ",")[8])

		// These may be missing, so pointer types have been used.
		armcd := CPUtils.Str2IntP(strings.Split(str, ",")[9])
		arm := CPUtils.Str2StrP(strings.Split(str, ",")[10])

		// The output object is a slice of pointers to the Subject struct.
		subj = append(subj, &Subject{
			Studyid: studyid,
			Subjid:  subjid,
			Siteid:  siteid,
			Usubjid: usubjid,
			Rectype: rectype,
			Dmdtc:   dmdtc,
			Endv:    endv,
			Rfstdtc: rfstdtc,
			Rfendtc: rfendtc,
			Armcd:   armcd,
			Arm:     arm,
		})
	}
	return subj
}

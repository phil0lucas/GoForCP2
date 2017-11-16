// Program to generate SDTM data for a fictitious study.
// Domain DM
// Metadata :
// - STUDYID Char 6  (constant) Study Identifier
// - DOMAIN  Char 2  (constant) Domain abbreviation
// - USUBJID Char 18 STUDYID-SITEID-SUBJID Unique Subject Identifier
// - SUBJID  Char 6  Subject Identifier
// - SITEID  Char 4  Site Identifier
// - RFSTDTC Date 10 ISO8601 First date of study med exposure
// - RFENDTC Date 10 ISO8601 Last date of study med exposure
// - DMDTC   Date 10 ISO8601 Date/Time of Collection
// - INVID   Char 3  Investigator code
// - INVNAME Char 8  Investigator Name
// - COUNTRY Char 3  ISO3166 Country code
// - BRTHDTC Date 10 ISO8601 Subjects date of birth
// - AGE	 Num     Subject's age (min 20, Max 80)
// - AGEU    Char 5  (constant) Age units
// - SEX     Char 1  Subject's gender ((M/F)
// - RACE    Char 5  Subject's race (White, Black, Asian)
// - ARMCD   Num     Treatment Arm code
// - ARM     Char 7  Treatment Arm
// - DMDY    Num     Study Day of collection

// 	Screening Failure subjects will be included and have missing values
//	for some of their data fields
package DM

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/phil0lucas/GoForCP/CPUtils"
)

// The data will be created as a slice of pointers to objects of the struct
// shown below.
// This will mirror the metadata above with more natural types.
// Those fields that may have a missing value are modeled with pointers.
type Dmrec struct {
	Studyid string
	Domain  string
	Usubjid string
	Subjid  string
	Siteid  string
	Rfstdtc *time.Time
	Rfendtc *time.Time
	Dmdtc   time.Time
	Invid   string
	Invname string
	Country string
	Ageu    string
	Age     *int
	Brthdtc *time.Time
	Sex     *string
	Race    *string
	Armcd   *int
	Arm     *string
	Dmdy    int
}

// Various lookups for random selection
var invid = map[int]string{0: "AAA", 1: "BBB", 2: "CCC", 3: "DDD", 4: "EEE"}
var invnm = map[int]string{0: "Smith", 1: "Jones", 2: "Robinson", 3: "Brown", 4: "Green"}
var ctrymap = map[int]string{0: "GBR", 1: "USA", 2: "FRA", 3: "GER", 4: "SWE"}
var sexmp = map[int]string{0: "M", 1: "F"}
var racemp = map[int]string{0: "White", 1: "Black", 2: "Asian"}

// Constants in use in the program
const (
	domain = "DM"
	ageu   = "Years"
	dmdy   = 0
)

// To generate a random age within a range
// min age = 20, max age = 80
func getAge() *int {
	if CPUtils.FlagMiss(0) == false {
		rand.Seed(time.Now().UTC().UnixNano())
		r := rand.Intn(59) + 20
		return &r
	} else {
		return nil
	}
}

// Generate a birth date based on the recorded age
func getBday(dmdtc time.Time, age *int) *time.Time {
	// Birth date is recorded at screening, which is DMDTC here.
	// Having randomly generated an age, calculate the last possible
	// birthday at that age and then subtract a random number
	// of days between 0 and 364
	if age != nil {
		v := *age
		_bdate := dmdtc.AddDate(-v, 0, 0)
		rand.Seed(time.Now().UTC().UnixNano())
		offset := rand.Intn(364)
		b := _bdate.AddDate(0, 0, -offset)
		return &b
	} else {
		return nil
	}
}

// Generate the DM data for each subject and write to a slice of pointers
// before writing to an output CSV
func WriteDM(infile, outfile *string) {

	// open the file and pass it to a Scanner object
	file, err := os.Open(*infile)
	if err != nil {
		panic(fmt.Sprintf("error opening %s: %v", infile, err))
	}
	defer file.Close()

	// Output slice of pointers to structs
	var dm []*Dmrec

	// Pass the opened file to a scanner
	scanner := bufio.NewScanner(file)

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
		// Screening date
		dmdtc, _ := time.Parse("2006-01-02", strings.Split(str, ",")[5])
		// 		// First date of dosing for randomized subjects.
		// 		// For screening failures this will be a nil pointer.
		rfstdtc := CPUtils.Str2DateP(strings.Split(str, ",")[7])
		// 		//	Last day of dosing.
		// 		//	This will also be missing if the subject is a screening failure
		rfendtc := CPUtils.Str2DateP(strings.Split(str, ",")[8])
		iKey, invid := CPUtils.RandItem(invid)
		invname := invnm[iKey]
		_, country := CPUtils.RandItem(ctrymap)
		age := getAge()
		brthdtc := getBday(dmdtc, age)
		sex := CPUtils.RandItemP(sexmp)
		race := CPUtils.RandItemP(racemp)
		armcd := CPUtils.Str2IntP(strings.Split(str, ",")[9])
		arm := CPUtils.Str2StrP(strings.Split(str, ",")[10])
		//
		dm = append(dm, &Dmrec{
			Studyid: studyid,
			Domain:  domain,
			Usubjid: usubjid,
			Subjid:  subjid,
			Siteid:  siteid,
			Rfstdtc: rfstdtc,
			Rfendtc: rfendtc,
			Dmdtc:   dmdtc,
			Invid:   invid,
			Invname: invname,
			Country: country,
			Ageu:    ageu,
			Age:     age,
			Brthdtc: brthdtc,
			Sex:     sex,
			Race:    race,
			Armcd:   armcd,
			Arm:     arm,
			Dmdy:    dmdy,
		})
	}

	// Output file writing section
	fo, err := os.Create(*outfile)
	if err != nil {
		log.Fatal(err)
	}
	defer fo.Close()

	// Create a buffered writer to the file
	w := bufio.NewWriter(fo)

	for ii, _ := range dm {
		bytesWritten, err := w.WriteString(
			dm[ii].Studyid + "," +
				dm[ii].Domain + "," +
				dm[ii].Subjid + "," +
				dm[ii].Siteid + "," +
				dm[ii].Usubjid + "," +
				CPUtils.DateP2Str(dm[ii].Rfstdtc) + "," +
				CPUtils.DateP2Str(dm[ii].Rfendtc) + "," +
				dm[ii].Dmdtc.Format("2006-01-02") + "," +
				dm[ii].Invid + "," +
				dm[ii].Invname + "," +
				dm[ii].Country + "," +
				CPUtils.IntP2Str(dm[ii].Age) + "," +
				dm[ii].Ageu + "," +
				CPUtils.DateP2Str(dm[ii].Brthdtc) + "," +
				CPUtils.StrP2Str(dm[ii].Sex) + "," +
				CPUtils.StrP2Str(dm[ii].Race) + "," +
				CPUtils.IntP2Str(dm[ii].Armcd) + "," +
				CPUtils.StrP2Str(dm[ii].Arm) + "," +
				strconv.Itoa(dm[ii].Dmdy) +
				"\n")

		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Bytes written: %d\n", bytesWritten)
	}

	// Write to disk
	w.Flush()
}

// Read the CSV and write to the same slice of structs.
func ReadDM(infile *string) []*Dmrec {
	// open the file and pass it to a Scanner object
	file, err := os.Open(*infile)
	if err != nil {
		panic(fmt.Sprintf("error opening %s: %v", *infile, err))
	}
	defer file.Close()

	// Pass the opened file to a scanner
	scanner := bufio.NewScanner(file)

	var dmx []*Dmrec
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
		rfstdtc := CPUtils.Str2DateP(strings.Split(str, ",")[5])
		rfendtc := CPUtils.Str2DateP(strings.Split(str, ",")[6])
		dmdtc, _ := time.Parse("2006-01-02", strings.Split(str, ",")[7])
		invid := strings.Split(str, ",")[8]
		invname := strings.Split(str, ",")[9]
		country := strings.Split(str, ",")[10]
		ageu := strings.Split(str, ",")[12]
		age := CPUtils.Str2IntP(strings.Split(str, ",")[11])
		bday := CPUtils.Str2DateP(strings.Split(str, ",")[13])
		sex := CPUtils.Str2StrP(strings.Split(str, ",")[14])
		race := CPUtils.Str2StrP(strings.Split(str, ",")[15])
		armcd := CPUtils.Str2IntP(strings.Split(str, ",")[16])
		arm := CPUtils.Str2StrP(strings.Split(str, ",")[17])
		dmdy, _ := strconv.Atoi(strings.Split(str, ",")[18])

		dmx = append(dmx, &Dmrec{
			Studyid: studyid,
			Domain:  domain,
			Subjid:  subjid,
			Siteid:  siteid,
			Usubjid: usubjid,
			Rfstdtc: rfstdtc,
			Rfendtc: rfendtc,
			Dmdtc:   dmdtc,
			Invid:   invid,
			Invname: invname,
			Country: country,
			Ageu:    ageu,
			Age:     age,
			Brthdtc: bday,
			Sex:     sex,
			Race:    race,
			Armcd:   armcd,
			Arm:     arm,
			Dmdy:    dmdy,
		})
	}
	return dmx
}

// Some utilities related to this 'Domain'
// This counts subjects per treatment group (Arm)
// including counts for the full screened population
// the screen failures and the overall count of randomized subjects (Active + Placebo)
func CountByTG(dm []*Dmrec) map[string]int {
	m := make(map[string]int)
	m["Screened"] = len(dm)
	for _, v := range dm {
		if v.Arm != nil {
			m[*v.Arm]++
		} else {
			m["SF"]++
		}
	}

	total := 0
	for k, v := range m {
		if k != "SF" && k != "Screened" {
			total += v
		}
	}
	m["Overall"] = total
	return m
}

// Returns a slice of the unique treatment group (Arm) values
func UniqueTG(dm []*Dmrec) []string {
	var s []string
	for _, v := range dm {
		if v.Arm != nil && !CPUtils.StringInSlice(*v.Arm, s) {
			s = append(s, *v.Arm)
		}
	}
	s = append(s, "Overall")
	return s
}

// Takes the slice of pointers to the Dmrec structs and returns the same
// excluding the screen failure subjects.
func RemoveSF(dm []*Dmrec) []*Dmrec {
	var dm2 []*Dmrec
	for _, v := range dm {
		// Exclude SFs
		if v.Arm != nil {
			dm2 = append(dm2, v)
		}
	}
	return dm2
}

// For a given Arm, subset the slice of pointers to Dmrec
func SubsetByArm(dm []*Dmrec, value string) []*Dmrec {
	var subdm []*Dmrec
	for _, v := range dm {
		if *v.Arm == value || value == "Overall" {
			subdm = append(subdm, v)
		}
	}
	return subdm
}

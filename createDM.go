//	This program is a driver program for the DM domain

package main

import (
	"flag"
	"github.com/phil0lucas/GoForCP2/DM"
)

// The program will be run with flags to specify the input & output files.
// 	When the program is run the input and output files can be changed using the
//	-i and -o flags
var infile = flag.String("i", "sc.csv", "Name of input file")
var outfile = flag.String("o", "dm.csv", "Name of output file")

func main() {
	flag.Parse()
	DM.WriteDM(infile, outfile)
}

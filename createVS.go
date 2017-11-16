// This is a driver program to create the VS domain data set

package main

import (
	"flag"
	"github.com/phil0lucas/GoForCP2/VS"
)

// 	The program will be run with flags to specify the input & output files
// 	When the program is run the input and output files can be changed using the
//	-i and -o flags
var infile = flag.String("i", "sc.csv", "Name of input file")
var outfile = flag.String("o", "vs.csv", "Name of output file")

func main() {
	flag.Parse()
	VS.WriteVS(infile, outfile)
}

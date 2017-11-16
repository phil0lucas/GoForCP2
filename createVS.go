// This is a driver program to create the VS domain data set

package main

import (
	"flag"
	"github.com/phil0lucas/GoForCP/VS"
)

// 	The program will be run with flags to specify the input & output files
// 	When the program is run the input and output files can be changed using the
//	-i and -o flags
var infile = flag.String("i", "sc3.csv", "Name of input file")
var outfile = flag.String("o", "vs3.csv", "Name of output file")

func main() {
	flag.Parse()
	VS.WriteVS(infile, outfile)
}

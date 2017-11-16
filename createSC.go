// 	This is a driver program to create the SC data set.
//	This data provides a common basis for the other two 'domains'

package main

import (
	"flag"
	"github.com/phil0lucas/GoForCP/SC"
)

//	The -o flag allows change of the output file name
var outfile = flag.String("o", "sc3.csv", "Name of output file")

func main() {
	flag.Parse()
	SC.WriteSC(outfile)
}

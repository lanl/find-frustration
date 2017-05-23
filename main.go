/*
find-frustration reads a graph and reports various statistics on how
much frustration exists in the graph when treated as an Ising or QUBO
problem.
*/
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

// notify is used to output error messages.
var notify *log.Logger

// Empty represents a zero-byte object.
type Empty struct{}

// checkError is a convenience function that aborts on error.
func checkError(e error) {
	if e != nil {
		notify.Fatal(e)
	}
}

// A Graph is a collection of named vertices and edges.  Both vertices and
// edges have an associated weight.
type Graph struct {
	Vs map[string]float64    // Map from a vertex to a weight
	Es map[[2]string]float64 // Map from an edge to a weight
}

func main() {
	// Parse the command line.
	var err error
	notify = log.New(os.Stderr, os.Args[0]+": ", 0)
	inFmt := ""
	flag.StringVar(&inFmt, "format", "qubist", `input file format: "qubist" (default), "qubo", "qmasm", or "bqpjson"`)
	flag.StringVar(&inFmt, "f", "qubist", "shorthand for --format")
	outFile := ""
	flag.StringVar(&outFile, "output", "", "output file name (default: standard output)")
	flag.StringVar(&outFile, "o", "", "shorthand for --output")
	allCycs := flag.Bool("all-cycles", false, "Combine base cycles into elementary cycles (extremely slow; default: false)")
	flag.Parse()

	// Open the output file.
	var w io.Writer = os.Stdout
	if outFile != "" {
		f, err := os.Create(outFile)
		checkError(err)
		defer f.Close()
		w = f
	}

	// Open the input file.
	var r io.Reader
	switch flag.NArg() {
	case 0:
		// Read from standard input.
		r = os.Stdin
	case 1:
		// Read from the named file.
		r, err = os.Open(flag.Arg(0))
		checkError(err)
	default:
		notify.Fatal("More than one input file was specified")
	}

	// Read the input file into a graph.
	var g Graph
	switch inFmt {
	case "qmasm":
		g = ReadQMASMFile(r)
	case "qubist":
		g = ReadQubistFile(r)
	case "qubo":
		g = ReadQUBOFile(r)
	case "bqpjson":
		g = ReadBqpjsonFile(r)
	default:
		notify.Fatalf("Unrecognized input format %q", inFmt)
	}

	// Acquire a list of basic cycles and from that, if requested, a list
	// of elementary cycles.
	bPath := g.baseCyclePaths()
	bcs := make([][][2]string, len(bPath))
	for i, p := range bPath {
		bcs[i] = g.pathToEdges(p)
	}
	if len(bcs) == 0 {
		notify.Print("Graph is acyclic; no frustration can exist")
		os.Exit(0)
	}
	fmt.Fprintf(w, "#BCS %d\n", len(bcs))
	var ecs [][][2]string
	if *allCycs {
		ecs = g.elementaryCycles(bcs)
		fmt.Fprintf(w, "#ECS %d\n", len(ecs))
	} else {
		ecs = bcs
	}

	// Tell the user what we discovered.
	OutputResults(w, g, ecs)
}

/* This file outputs various statistics about the frustration that appears in
a graph. */

package main

import (
	"fmt"
	"io"
)

// outputVertices outputs all vertices, categorized and tallied.
func outputVertices(w io.Writer, g Graph, ps [][]string, isFrust []bool) {
	// Tally the number of times each vertex appears in a frustrated cycle
	// and in a non-frustrated cycle.
	fVerts := make(map[string]int)
	nfVerts := make(map[string]int)
	for i, p := range ps {
		for _, v := range p {
			if isFrust[i] {
				fVerts[v]++
			} else {
				nfVerts[v]++
			}
		}
	}

	// Output each vertex, categorized and tallied.  Keep track of the
	// number of vertices that are more frustrated than not frustrated.
	nfvs := 0 // Number of frustrated vertices
	for v, t := range fVerts {
		if t > nfVerts[v] {
			fmt.Fprintf(w, "FV   %d %d | %s\n", t, t-nfVerts[v], v)
			nfvs++
		}
	}
	for v, t := range nfVerts {
		if t >= fVerts[v] {
			fmt.Fprintf(w, "NFV  %d %d | %s\n", t, t-fVerts[v], v)
		}
	}

	// Output some summary statistics.
	fmt.Fprintf(w, "#FV  %d / %d = %f\n", nfvs, len(g.Vs), float64(nfvs)/float64(len(g.Vs)))
}

// outputEdges outputs all edges, categorized and tallied.
func outputEdges(w io.Writer, g Graph, ps [][]string, isFrust []bool) {
	// Tally the number of times each edge appears in a frustrated cycle
	// and in a non-frustrated cycle.
	fEdges := make(map[[2]string]int)
	nfEdges := make(map[[2]string]int)
	for i, p := range ps {
		for j, v1 := range p {
			v2 := p[(j+1)%len(p)]
			if v1 > v2 {
				v1, v2 = v2, v1
			}
			e := [2]string{v1, v2}
			if isFrust[i] {
				fEdges[e]++
			} else {
				nfEdges[e]++
			}
		}
	}

	// Output each edge, categorized and tallied.
	nfes := 0 // Number of frustrated edges
	for e, t := range fEdges {
		if t > nfEdges[e] {
			fmt.Fprintf(w, "FE   %d %d | %s %s\n", t, t-nfEdges[e], e[0], e[1])
			nfes++
		}
	}
	for e, t := range nfEdges {
		if t >= fEdges[e] {
			fmt.Fprintf(w, "NFE  %d %d | %s %s\n", t, t-fEdges[e], e[0], e[1])
		}
	}

	// Output some summary statistics.
	fmt.Fprintf(w, "#FE  %d / %d = %f\n", nfes, len(g.Es), float64(nfes)/float64(len(g.Es)))
}

// outputCycles outputs all cycles, categorized and tallied.
func outputCycles(w io.Writer, g Graph, ps [][]string, isFrust []bool) {
	// Output each cycle preceded by whether it is frustrated or not.  As
	// we go along, tally the number of frustrated cycles encountered.
	fvs := make(map[string]Empty, len(g.Vs))
	nfcs := 0 // Number of frustrated cycles
	for i, p := range ps {
		f := isFrust[i]
		if f {
			fmt.Fprintf(w, "FC  ")
			nfcs++
		} else {
			fmt.Fprintf(w, "NFC ")
		}
		for _, v := range p {
			fmt.Fprintf(w, " %s", v)
			if f {
				fvs[v] = Empty{}
			}
		}
		fmt.Fprintln(w, "")
	}

	// Output some summary statistics.
	fmt.Fprintf(w, "#FC  %d / %d = %f\n", nfcs, len(ps), float64(nfcs)/float64(len(ps)))
}

// OutputResults is the program's top-level output routine.  It outputs a
// variety of information about frustration within a graph.
func OutputResults(w io.Writer, g Graph, ecs [][][2]string) {
	// Convert the edges back to paths for a more readable presentation.
	// Determine which paths are frustrated cycles.
	ps := make([][]string, len(ecs))
	isFrust := make([]bool, len(ecs))
	for i, ec := range ecs {
		ps[i] = g.edgesToPath(ec)
		isFrust[i] = g.isFrustrated(ps[i])
	}

	// Output information about the graph's vertices, edges, and cycles.
	outputVertices(w, g, ps, isFrust)
	outputEdges(w, g, ps, isFrust)
	outputCycles(w, g, ps, isFrust)
}

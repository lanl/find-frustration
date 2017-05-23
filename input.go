/* This file provides functions for reading and parsing input files in
different formats. */

package main

import (
	"bufio"
	"encoding/json"
	"io"
	"strconv"
	"strings"
)

// quboToIsing converts a QUBO problem to an Ising problem.
func quboToIsing(vs map[string]float64, es map[[2]string]float64) {
	for i, wt := range vs {
		vs[i] = wt / 2
	}
	for ij, wt := range es {
		i, j := ij[0], ij[1]
		wt4 := wt / 4
		es[ij] = wt4
		vs[i] += wt4
		vs[j] += wt4
	}
}

// ReadQMASMFile returns the Ising Hamiltonian represented by a QMASM source
// file.
func ReadQMASMFile(r io.Reader) Graph {
	vs := make(map[string]float64)    // Map from a vertex to a weight
	es := make(map[[2]string]float64) // Map from an edge to a weight
	rb := bufio.NewReader(r)
	for {
		// Read one line.
		ln, err := rb.ReadString('\n')
		if err == io.EOF {
			break
		}
		checkError(err)

		// Discard comments.
		hIdx := strings.Index(ln, "#")
		if hIdx >= 0 {
			ln = ln[:hIdx]
		}

		// Parse the line.
		fs := strings.Fields(ln)
		switch len(fs) {
		case 2:
			// Vertex
			v := fs[0]
			wt, err := strconv.ParseFloat(fs[1], 64)
			checkError(err)
			vs[v] += wt
		case 3:
			// Edge, chain, or alias
			var u, v string
			var wt float64
			if fs[1] == "=" || fs[1] == "<->" {
				// Chain or alias
				u, v = fs[0], fs[2]
				wt = -1.0
			} else {
				u, v = fs[0], fs[1]
				wt, err = strconv.ParseFloat(fs[2], 64)
				checkError(err)
			}
			if u > v {
				u, v = v, u
			}
			es[[2]string{u, v}] += wt
			vs[u] += 0.0
			vs[v] += 0.0
		}
	}
	return Graph{Vs: vs, Es: es}
}

// ReadQubistFile returns the Ising Hamiltonian represented by a Qubist source
// file.
func ReadQubistFile(r io.Reader) Graph {
	// Read and discard the first (header) line.
	vs := make(map[string]float64)    // Map from a vertex to a weight
	es := make(map[[2]string]float64) // Map from an edge to a weight
	rb := bufio.NewReader(r)
	ln, err := rb.ReadString('\n')
	checkError(err)

	// Process all remaining lines.
	for {
		// Read one line.
		ln, err = rb.ReadString('\n')
		if err == io.EOF {
			break
		}
		checkError(err)

		// Parse the line.
		fs := strings.Fields(ln)
		if len(fs) == 3 {
			u, v := fs[0], fs[1]
			wt, err := strconv.ParseFloat(fs[2], 64)
			checkError(err)
			if u == v {
				// Vertex
				vs[u] += wt
			} else {
				// Edge
				if u > v {
					u, v = v, u
				}
				es[[2]string{u, v}] += wt
				vs[u] += 0.0
				vs[v] += 0.0
			}
		} else {
			notify.Fatalf("Failed to parse Qubist line %q", strings.TrimSpace(ln))
		}
	}
	return Graph{Vs: vs, Es: es}
}

// ReadQUBOFile returns the Ising Hamiltonian represented by a QUBO source file.
func ReadQUBOFile(r io.Reader) Graph {
	// Read a list of edges and vertices in QUBO format.
	vs := make(map[string]float64)    // Map from a vertex to a weight
	es := make(map[[2]string]float64) // Map from an edge to a weight
	rb := bufio.NewReader(r)
	for {
		// Read one line.
		ln, err := rb.ReadString('\n')
		if err == io.EOF {
			break
		}
		checkError(err)

		// Parse the line.
		fs := strings.Fields(ln)
		if len(fs) == 0 {
			continue // Blank line
		}
		switch fs[0] {
		case "c":
			continue // Comment
		case "p":
			if len(fs) != 6 || fs[1] != "qubo" {
				notify.Fatalf("Failed to parse QUBO line %q", strings.TrimSpace(ln))
			}
			continue // Don't bother validating the problem size.
		}
		if len(fs) != 3 {
			notify.Fatalf("Failed to parse QUBO line %q", strings.TrimSpace(ln))
		}
		u, v := fs[0], fs[1]
		wt, err := strconv.ParseFloat(fs[2], 64)
		checkError(err)
		if u == v {
			// Vertex
			vs[u] += wt
		} else {
			// Edge
			if u > v {
				u, v = v, u
			}
			es[[2]string{u, v}] += wt
			vs[u] += 0.0
			vs[v] += 0.0
		}

	}

	// Convert from a QUBO problem to an Ising problem and return that.
	quboToIsing(vs, es)
	return Graph{Vs: vs, Es: es}
}

// ReadBqpjsonFile returns the Ising Hamiltonian represented by a bqpjson
// source file (cf. https://github.com/lanl-ansi/bqpjson).
func ReadBqpjsonFile(r io.Reader) Graph {
	// Define the contents of a linear term.
	type LinearTerm struct {
		V      int     `json:"id"`    // Variable ID
		Weight float64 `json:"coeff"` // Variable weight
	}

	// Define the contents of a quadratic term.
	type QuadraticTerm struct {
		U      int     `json:"id_tail"` // First variable ID
		V      int     `json:"id_head"` // Second variable ID
		Weight float64 `json:"coeff"`   // Edge weight
	}

	// Specify only the parts of the bqpjson format in which we're
	// interested.
	type Bqpjson struct {
		VarDomain string          `json:"variable_domain"` // "spin" or "boolean"
		Scale     float64         `json:"scale"`           // Scale factor for all coefficients
		Offset    float64         `json:"offset"`          // Offset value for all coefficients
		LinTerms  []LinearTerm    `json:"linear_terms"`    // List of linear terms
		QuadTerms []QuadraticTerm `json:"quadratic_terms"` // List of quadratic terms
	}

	// Read the graph description in bqpjson format.
	var desc Bqpjson
	dec := json.NewDecoder(r)
	err := dec.Decode(&desc)
	checkError(err)

	// Extract a list of edges and a list of vertices.
	vs := make(map[string]float64)    // Map from a vertex to a weight
	es := make(map[[2]string]float64) // Map from an edge to a weight
	for _, lt := range desc.LinTerms {
		vs[strconv.Itoa(lt.V)] += lt.Weight
	}
	for _, qt := range desc.QuadTerms {
		u := strconv.Itoa(qt.U)
		v := strconv.Itoa(qt.V)
		if u > v {
			u, v = v, u
		}
		es[[2]string{u, v}] += qt.Weight
		vs[u] += 0.0
		vs[v] += 0.0
	}

	// Multiply all weights by the scale parameter then add the offset
	// parameter.
	for v, wt := range vs {
		vs[v] = wt*desc.Scale + desc.Offset
	}
	for v, wt := range es {
		es[v] = wt*desc.Scale + desc.Offset
	}

	// Convert from QUBO to Ising if the problem was specified as QUBO.
	switch desc.VarDomain {
	case "boolean":
		quboToIsing(vs, es)
	case "spin":
	default:
		notify.Fatalf("Unrecognized variable_domain %q", desc.VarDomain)
	}

	// Return the resulting graph.
	return Graph{Vs: vs, Es: es}
}

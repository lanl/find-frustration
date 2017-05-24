find-frustration
================

Description
-----------

find-frustration reads an optimization problem expressed as either a [QUBO](https://en.wikipedia.org/wiki/Quadratic_unconstrained_binary_optimization) or an [Ising Hamiltonian](https://en.wikipedia.org/wiki/Ising_model) and outputs some metrics that indicate how difficult the problem may be to solve based on the amount of frustration it contains.

Explanation
-----------

Consider, for example, an optimization problem that expresses the following constraints on variables *A*, *B*, and *C*, each of which is ±1 (i.e., an Ising problem):

  * *A* = *B*
  * *B* = *C*
  * *C* = *A*

Because it is possible to satisfy all three constraints—the three variables can all be +1 or all be −1—we say there is no frustration in the problem.  However, if the constraints were instead

  * *A* = *B*
  * *B* = *C*
  * *C* ≠ *A*

then at most two of the three constraints can be satisfied.  In this case we say the problem is *frustrated*.  Our hypothesis is that the difficulty of satisfying the maximal number of constraints is correlated to the amount of frustration in the system.

It helps to think of QUBO/Ising problems as graphs.  Each vertex corresponds to a variable, and each edge correspond to a constraint between two variables.  Vertex and edge weights can be interpreted as follows:

| Weight type | Weight | Vertex 1 | Vertex 2 | Meaning                                 |
| :---------- | -----: | :------: | :------: | :-------------------------------------- |
| Vertex      | < 0    | *U*      | N/A      | Want *U* = +1                           |
| Vertex      | = 0    | *U*      | N/A      | Don't care what value *U* takes         |
| Vertex      | > 0    | *U*      | N/A      | Want *U* = −1                           |
| Edge        | < 0    | *U*      | *V*      | Want *U* = *V*                          |
| Edge        | = 0    | *U*      | *V*      | Don't care what relation *U* has to *V* |
| Edge        | > 0    | *U*      | *V*      | Want *U* ≠ *V*                          |

The greater in magnitude the weight, the stronger the desire expressed.

Installation
------------

find-frustration is written in [Go](https://golang.org/) so you'll need to install a Go compiler.  Then, find-frustration can be installed with
```bash
go get github.com/lanl/find-frustration
```

Alternatively, you can clone the find-frustration repository from GitHub, switch into the find-frustration directory, and build with
```bash
go build -o find-frustration *.go
```

Usage
-----

Run
```bash
find-frustration --help
```
for a list of command-line options.  The most important option is `--format`, which specifies the input format: `qubist` (the default), [`qubo`](https://github.com/dwavesystems/qbsolv), [`qmasm`](https://github.com/lanl/qmasm), or [`bqpjson`](https://github.com/lanl-ansi/bqpjson).

Qubist format comprises a header line that specifies the maximum vertex number + 1 and the number of rows that follow.  Each row specifies two vertices (non-negative integers) and the weight of the edge that connects them (a floating-point number).  The frustrated system presented under *Explanation* can be expressed like this:
```
1152 3
0 1 -1.0
1 2 -1.0
2 0  1.0
```

Running that through find-frustration produces output like the following:
```
#BCS 1
FV   1 1 | 0
FV   1 1 | 1
FV   1 1 | 2
#FV  3 / 3 = 1.000000
FE   1 1 | 0 1
FE   1 1 | 1 2
FE   1 1 | 0 2
#FE  3 / 3 = 1.000000
FC   0 1 2
#FC  1 / 1 = 1.000000
```
Note that output from find-frustration is non-deterministic and can vary slightly from run to run.

Interpretation
--------------

The output of find-frustration is designed to be easy to parse mechanically yet also simple for a human to follow.  Information is output as a sequence of lines.  Each line consists of a set of space-separated columns beginning with a tag.  The following information is output:

  * Number of basic cycles

    - Tag: `#BCS`
    - Argument: Number of basic (a.k.a. fundamental) cycles
    - Number of occurrences: 1

  * Number of elementary cycles

    - Tag: `#ECS`
    - Argument: Number of elementary cycles
    - Number of occurrences: 1 if `--all-cycles` is specified on the command line, 0 otherwise

  * Non-frustrated vertex

    - Tag: `NFV`
    - Arguments: 〈# of non-frustrated cycles containing the vertex〉〈# of non-frustrated cycles containing the vertex minus # of frustrated cycles containing the vertex> `|` 〈vertex name〉
    - Number of occurrences: 1 for each vertex that occurs more often in non-frustrated cycles than in frustrated cycles

  * Frustrated vertex

    - Tag: `FV`
    - Arguments: 〈# of frustrated cycles containing the vertex〉〈# of frustrated cycles containing the vertex minus # of non-frustrated cycles containing the vertex> `|` 〈vertex name〉
    - Number of occurrences: 1 for each vertex that occurs more often in frustrated cycles than in non-frustrated cycles

  * Number of frustrated vertices

    - Tag: `#FV`
    - Arguments: 〈# of `FV` tags〉`/` 〈total # of vertices> `=` 〈quotient〉
    - Number of occurrences: 1

  * Non-frustrated edge

    - Tag: `NFE`
    - Arguments: 〈# of non-frustrated cycles containing the edge〉〈# of non-frustrated cycles containing the edge minus # of frustrated cycles containing the edge> `|` 〈name of vertex 1〉 〈name of vertex 2〉
    - Number of occurrences: 1 for each edge that occurs more often in non-frustrated cycles than in frustrated cycles

  * Frustrated edge

    - Tag: `FE`
    - Arguments: 〈# of frustrated cycles containing the edge〉〈# of frustrated cycles containing the edge minus # of non-frustrated cycles containing the edge> `|` 〈name of vertex 1〉 〈name of vertex 2〉
    - Number of occurrences: 1 for each edge that occurs more often in frustrated cycles than in non-frustrated cycles

  * Number of frustrated edges

    - Tag: `#FE`
    - Arguments: 〈# of `FE` tags〉`/` 〈total # of edges> `=` 〈quotient〉
    - Number of occurrences: 1

  * Non-frustrated cycle

    - Tag: `NFC`
    - Arguments: 〈vertex〉…
    - Number of occurrences: 1 for each non-frustrated cycle

  * Frustrated cycle

    - Tag: `FC`
    - Arguments: 〈vertex〉…
    - Number of occurrences: 1 for each frustrated cycle

  * Number of frustrated cycles

    - Tag: `#FC`
    - Arguments: 〈# of `FC` tags〉`/` 〈total # of cycles> `=` 〈quotient〉
    - Number of occurrences: 1

License
-------

find-frustration is provided under a BSD-ish license with a "modifications must be indicated" clause.  See [the LICENSE file](http://github.com/lanl/find-frustration/blob/master/LICENSE.md) for the full text.

This package is part of the Hybrid Quantum-Classical Computing suite, known internally as LA-CC-16-032.

Author
------

Scott Pakin, <pakin@lanl.gov>

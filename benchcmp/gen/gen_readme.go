// Command gen_readme converts raw `go test -bench` output into a markdown
// section comparing github.com/fluxrpc/solana-go ("flux") against
// github.com/gagliardetto/solana-go ("gagl").
//
// Usage: go run ./gen results.txt > readme_section.md
//
// The input may contain output concatenated from multiple runs; the last
// occurrence of a benchmark name wins.
package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// result holds one implementation's numbers for a single type+op.
type result struct {
	ok       bool
	nsPerOp  float64
	bPerOp   int64 // -1 when not reported
	allocsOp int64 // -1 when not reported
}

type opRow struct {
	op   string
	flux result
	gagl result
}

type typeGroup struct {
	name string
	ops  []*opRow
}

// benchLine matches e.g.:
//
//	BenchmarkPublicKey_String/flux-8  3144274  70.13 ns/op  48 B/op  1 allocs/op
var benchLine = regexp.MustCompile(`^Benchmark([A-Za-z0-9]+)_([A-Za-z0-9]+)/(flux|gagl)(?:-\d+)?\s+\d+\s+(.*)$`)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <results-file>\n", os.Args[0])
		os.Exit(2)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	var groups []*typeGroup
	byName := map[string]*typeGroup{}

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<20)
	for sc.Scan() {
		m := benchLine.FindStringSubmatch(strings.TrimSpace(sc.Text()))
		if m == nil {
			continue
		}
		typeName, opName, impl, rest := m[1], m[2], m[3], m[4]

		res, ok := parseMetrics(rest)
		if !ok {
			continue
		}

		g := byName[typeName]
		if g == nil {
			g = &typeGroup{name: typeName}
			byName[typeName] = g
			groups = append(groups, g)
		}
		var row *opRow
		for _, r := range g.ops {
			if r.op == opName {
				row = r
				break
			}
		}
		if row == nil {
			row = &opRow{op: opName}
			g.ops = append(g.ops, row)
		}
		// Later lines override earlier ones (concatenated runs).
		if impl == "flux" {
			row.flux = res
		} else {
			row.gagl = res
		}
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(groups) == 0 {
		fmt.Fprintln(os.Stderr, "no benchmark results found in input")
		os.Exit(1)
	}

	var out strings.Builder
	writeHeader(&out)
	for _, g := range groups {
		writeTable(&out, g)
	}
	writeChart(&out, groups)
	os.Stdout.WriteString(out.String())
}

// parseMetrics parses the "70.13 ns/op  48 B/op  1 allocs/op" tail of a
// benchmark line.
func parseMetrics(rest string) (result, bool) {
	res := result{bPerOp: -1, allocsOp: -1}
	fields := strings.Fields(rest)
	for i := 0; i+1 < len(fields); i += 2 {
		val, unit := fields[i], fields[i+1]
		switch unit {
		case "ns/op":
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return res, false
			}
			res.nsPerOp = v
			res.ok = true
		case "B/op":
			if v, err := strconv.ParseInt(val, 10, 64); err == nil {
				res.bPerOp = v
			}
		case "allocs/op":
			if v, err := strconv.ParseInt(val, 10, 64); err == nil {
				res.allocsOp = v
			}
		}
	}
	return res, res.ok
}

func writeHeader(out *strings.Builder) {
	out.WriteString("## Benchmarks\n\n")
	fmt.Fprintf(out,
		"Comparing `github.com/fluxrpc/solana-go` (with `github.com/fluxrpc/base58`) against upstream `github.com/gagliardetto/solana-go` (with its bundled base58) for identical operations.\n\n")
	fmt.Fprintf(out, "Machine: %s. Go: %s (%s/%s).\n\n",
		cpuModel(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func writeTable(out *strings.Builder, g *typeGroup) {
	fmt.Fprintf(out, "### %s\n\n", g.name)
	out.WriteString("| Operation | fluxrpc ns/op | upstream ns/op | speedup | fluxrpc B/op | upstream B/op | fluxrpc allocs/op | upstream allocs/op |\n")
	out.WriteString("|---|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, r := range g.ops {
		fmt.Fprintf(out, "| %s | %s | %s | %s | %s | %s | %s | %s |\n",
			r.op,
			fmtNs(r.flux), fmtNs(r.gagl),
			speedup(r),
			fmtInt(r.flux, r.flux.bPerOp), fmtInt(r.gagl, r.gagl.bPerOp),
			fmtInt(r.flux, r.flux.allocsOp), fmtInt(r.gagl, r.gagl.allocsOp),
		)
	}
	out.WriteString("\n")
}

func writeChart(out *strings.Builder, groups []*typeGroup) {
	const maxWidth = 40

	out.WriteString("### ns/op comparison\n\n")
	out.WriteString("```text\n")
	first := true
	for _, g := range groups {
		for _, r := range g.ops {
			if !r.flux.ok || !r.gagl.ok {
				continue
			}
			if !first {
				out.WriteString("\n")
			}
			first = false

			slower := r.flux.nsPerOp
			if r.gagl.nsPerOp > slower {
				slower = r.gagl.nsPerOp
			}
			fmt.Fprintf(out, "%s_%s\n", g.name, r.op)
			writeBar(out, "flux", r.flux.nsPerOp, slower, maxWidth, r.flux.nsPerOp <= r.gagl.nsPerOp)
			writeBar(out, "gagl", r.gagl.nsPerOp, slower, maxWidth, r.gagl.nsPerOp < r.flux.nsPerOp)
		}
	}
	out.WriteString("```\n")
}

func writeBar(out *strings.Builder, label string, ns, slower float64, maxWidth int, faster bool) {
	width := maxWidth
	if slower > 0 {
		width = int(ns/slower*float64(maxWidth) + 0.5)
	}
	if width < 1 {
		width = 1
	}
	mark := ""
	if faster {
		mark = "  <-- faster"
	}
	fmt.Fprintf(out, "  %s  %-*s %s ns/op%s\n",
		label, maxWidth, strings.Repeat("█", width), fmtFloat(ns), mark)
}

func speedup(r *opRow) string {
	if !r.flux.ok || !r.gagl.ok || r.flux.nsPerOp == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%.1fx", r.gagl.nsPerOp/r.flux.nsPerOp)
}

func fmtNs(r result) string {
	if !r.ok {
		return "n/a"
	}
	return fmtFloat(r.nsPerOp)
}

func fmtInt(r result, v int64) string {
	if !r.ok || v < 0 {
		return "n/a"
	}
	return strconv.FormatInt(v, 10)
}

// fmtFloat formats ns/op with a precision that keeps small numbers readable
// and large numbers uncluttered.
func fmtFloat(v float64) string {
	switch {
	case v >= 1000:
		return strconv.FormatFloat(v, 'f', 0, 64)
	case v >= 100:
		return strconv.FormatFloat(v, 'f', 1, 64)
	default:
		return strconv.FormatFloat(v, 'f', 2, 64)
	}
}

func cpuModel() string {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return "unknown CPU"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "model name") {
			if _, after, ok := strings.Cut(line, ":"); ok {
				return strings.TrimSpace(after)
			}
		}
	}
	return "unknown CPU"
}

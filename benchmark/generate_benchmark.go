package main

import (
	"fmt"
	"os"
	"strings"
)

const numTypes = 100

func main() {
	var sb strings.Builder

	sb.WriteString("import ballerina/io;\n\n")

	// Generate 100 record type definitions
	for i := 0; i < numTypes; i++ {
		sb.WriteString(fmt.Sprintf("type RecordA%d record {\n", i))
		sb.WriteString("    int id;\n")
		sb.WriteString("    string name;\n")
		sb.WriteString(fmt.Sprintf("    int value%d;\n", i))
		sb.WriteString("};\n\n")

		sb.WriteString(fmt.Sprintf("type RecordB%d record {\n", i))
		sb.WriteString("    int id;\n")
		sb.WriteString("    float score;\n")
		sb.WriteString(fmt.Sprintf("    float metric%d;\n", i))
		sb.WriteString("};\n\n")
	}

	// Generate 100 functions with type narrowing on primitive unions
	for i := 0; i < numTypes; i++ {
		sb.WriteString(fmt.Sprintf("function narrowFunc%d(int|float|string input) returns int {\n", i))
		sb.WriteString("    if input is int {\n")
		sb.WriteString(fmt.Sprintf("        return input + %d;\n", i))
		sb.WriteString("    } else if input is float {\n")
		sb.WriteString(fmt.Sprintf("        return %d;\n", i*2))
		sb.WriteString("    } else {\n")
		sb.WriteString(fmt.Sprintf("        return %d;\n", -(i + 1)))
		sb.WriteString("    }\n")
		sb.WriteString("}\n\n")
	}

	// Generate 100 functions that use the record types
	for i := 0; i < numTypes; i++ {
		sb.WriteString(fmt.Sprintf("function useRecords%d(RecordA%d a, RecordB%d b) returns int {\n", i, i, i))
		sb.WriteString(fmt.Sprintf("    int v = a.value%d;\n", i))
		sb.WriteString(fmt.Sprintf("    int|float|string x = a.id + v;\n"))
		sb.WriteString(fmt.Sprintf("    int result = narrowFunc%d(x);\n", i))
		sb.WriteString(fmt.Sprintf("    return result + b.id;\n"))
		sb.WriteString("}\n\n")
	}

	// Generate main function that constructs records and calls all functions
	sb.WriteString("public function main() {\n")
	sb.WriteString("    int total = 0;\n")
	for i := 0; i < numTypes; i++ {
		sb.WriteString(fmt.Sprintf("    RecordA%d a%d = {id: %d, name: \"item%d\", value%d: %d};\n", i, i, i, i, i, i*10))
		sb.WriteString(fmt.Sprintf("    RecordB%d b%d = {id: %d, score: %d.5, metric%d: %d.0};\n", i, i, i+100, i, i, i*2))
		sb.WriteString(fmt.Sprintf("    total = total + useRecords%d(a%d, b%d);\n", i, i, i))
		sb.WriteString(fmt.Sprintf("    total = total + narrowFunc%d(a%d.id);\n", i, i))
		sb.WriteString(fmt.Sprintf("    total = total + narrowFunc%d(%d.5);\n", i, i))
		sb.WriteString(fmt.Sprintf("    total = total + narrowFunc%d(\"str%d\");\n", i, i))
	}
	sb.WriteString("    io:println(total);\n")
	sb.WriteString("}\n")

	outPath := "benchmark/large_benchmark.bal"
	if len(os.Args) > 1 {
		outPath = os.Args[1]
	}
	err := os.WriteFile(outPath, []byte(sb.String()), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Generated benchmark/large_benchmark.bal")
}

package dangerJs

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// parseDiffContent extracts the diff parsing logic for testing
func parseDiffContent(diffContent string) FileDiff {
	var fileDiff FileDiff
	// Only match lines that start with + or - but not +++ or --- (file headers)
	addedRe := regexp.MustCompile(`^\+([^+].*|$)`)
	removedRe := regexp.MustCompile(`^-([^-].*|$)`)

	lines := strings.Split(diffContent, "\n")
	for _, line := range lines {
		if matches := addedRe.FindStringSubmatch(line); len(matches) > 1 {
			fileDiff.AddedLines = append(fileDiff.AddedLines, DiffLine{Content: matches[1]})
		} else if matches := removedRe.FindStringSubmatch(line); len(matches) > 1 {
			fileDiff.RemovedLines = append(fileDiff.RemovedLines, DiffLine{Content: matches[1]})
		}
	}

	return fileDiff
}

func TestParseDiffContent(t *testing.T) {
	tests := []struct {
		name          string
		gitDiffOutput string
		wantFileDiff  FileDiff
	}{
		{
			name: "basic added and removed lines",
			gitDiffOutput: `diff --git a/test.go b/test.go
index 123..456 100644
--- a/test.go
+++ b/test.go
@@ -1 +1 @@
-func oldFunction() {
+func newFunction() {
@@ -5 +5,2 @@
+	return "added line"
-	fmt.Println("removed line")`,
			wantFileDiff: FileDiff{
				AddedLines: []DiffLine{
					{Content: "func newFunction() {", Line: 0},
					{Content: "\treturn \"added line\"", Line: 0},
				},
				RemovedLines: []DiffLine{
					{Content: "func oldFunction() {", Line: 0},
					{Content: "\tfmt.Println(\"removed line\")", Line: 0},
				},
			},
		},
		{
			name: "only added lines",
			gitDiffOutput: `diff --git a/new.go b/new.go
index 123..456 100644
--- a/new.go
+++ b/new.go
@@ -1,0 +1,3 @@
+package main
+
+func main() {}`,
			wantFileDiff: FileDiff{
				AddedLines: []DiffLine{
					{Content: "package main", Line: 0},
					{Content: "", Line: 0},
					{Content: "func main() {}", Line: 0},
				},
				RemovedLines: nil,
			},
		},
		{
			name: "only removed lines",
			gitDiffOutput: `diff --git a/old.go b/old.go
index 123..456 100644
--- a/old.go
+++ b/old.go
@@ -1,3 +0,0 @@
-package main
-
-func old() {}`,
			wantFileDiff: FileDiff{
				AddedLines: nil,
				RemovedLines: []DiffLine{
					{Content: "package main", Line: 0},
					{Content: "", Line: 0},
					{Content: "func old() {}", Line: 0},
				},
			},
		},
		{
			name:          "no changes",
			gitDiffOutput: ``,
			wantFileDiff: FileDiff{
				AddedLines:   nil,
				RemovedLines: nil,
			},
		},
		{
			name: "complex diff with context lines",
			gitDiffOutput: `diff --git a/complex.go b/complex.go
index 123..456 100644
--- a/complex.go
+++ b/complex.go
@@ -10,5 +10,6 @@
 	unchanged line 1
 	unchanged line 2
-	old implementation
+	new implementation
+	additional line
 	unchanged line 3`,
			wantFileDiff: FileDiff{
				AddedLines: []DiffLine{
					{Content: "\tnew implementation", Line: 0},
					{Content: "\tadditional line", Line: 0},
				},
				RemovedLines: []DiffLine{
					{Content: "\told implementation", Line: 0},
				},
			},
		},
		{
			name: "lines with special characters and whitespace",
			gitDiffOutput: `diff --git a/special.go b/special.go
index 123..456 100644
--- a/special.go
+++ b/special.go
@@ -1,2 +1,2 @@
-	fmt.Printf("Hello %s\n", name)
+	fmt.Printf("Hi %s!\n", name)`,
			wantFileDiff: FileDiff{
				AddedLines: []DiffLine{
					{Content: "\tfmt.Printf(\"Hi %s!\\n\", name)", Line: 0},
				},
				RemovedLines: []DiffLine{
					{Content: "\tfmt.Printf(\"Hello %s\\n\", name)", Line: 0},
				},
			},
		},
		{
			name: "lines with only symbols",
			gitDiffOutput: `diff --git a/symbols.go b/symbols.go
index 123..456 100644
--- a/symbols.go
+++ b/symbols.go
@@ -1,2 +1,2 @@
-}
+},`,
			wantFileDiff: FileDiff{
				AddedLines: []DiffLine{
					{Content: "},", Line: 0},
				},
				RemovedLines: []DiffLine{
					{Content: "}", Line: 0},
				},
			},
		},
		{
			name: "empty added and removed lines",
			gitDiffOutput: `diff --git a/empty.go b/empty.go
index 123..456 100644
--- a/empty.go
+++ b/empty.go
@@ -1,2 +1,2 @@
-
+
- 
+ `,
			wantFileDiff: FileDiff{
				AddedLines: []DiffLine{
					{Content: "", Line: 0},
					{Content: " ", Line: 0},
				},
				RemovedLines: []DiffLine{
					{Content: "", Line: 0},
					{Content: " ", Line: 0},
				},
			},
		},
		{
			name: "diff with file headers only",
			gitDiffOutput: `diff --git a/test.go b/test.go
index 123..456 100644
--- a/test.go
+++ b/test.go`,
			wantFileDiff: FileDiff{
				AddedLines:   nil,
				RemovedLines: nil,
			},
		},
		{
			name: "multiple hunks with mixed changes",
			gitDiffOutput: `diff --git a/multi.go b/multi.go
index 123..456 100644
--- a/multi.go
+++ b/multi.go
@@ -1,3 +1,4 @@
 package main
 
+import "fmt"
 func main() {
@@ -10,6 +11,7 @@
 	if true {
-		fmt.Println("old")
+		fmt.Println("new")
+		fmt.Println("extra")
 	}
 }`,
			wantFileDiff: FileDiff{
				AddedLines: []DiffLine{
					{Content: "import \"fmt\"", Line: 0},
					{Content: "\t\tfmt.Println(\"new\")", Line: 0},
					{Content: "\t\tfmt.Println(\"extra\")", Line: 0},
				},
				RemovedLines: []DiffLine{
					{Content: "\t\tfmt.Println(\"old\")", Line: 0},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFileDiff := parseDiffContent(tt.gitDiffOutput)
			require.Equal(t, tt.wantFileDiff, gotFileDiff)
		})
	}
}

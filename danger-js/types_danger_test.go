package dangerJs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Note: parseDiffContent is now a shared function in types_danger.go

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
					{Content: "func newFunction() {", Line: 1},
					{Content: "\treturn \"added line\"", Line: 5},
				},
				RemovedLines: []DiffLine{
					{Content: "func oldFunction() {", Line: 1},
					{Content: "\tfmt.Println(\"removed line\")", Line: 5},
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
					{Content: "package main", Line: 1},
					{Content: "", Line: 2},
					{Content: "func main() {}", Line: 3},
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
					{Content: "package main", Line: 1},
					{Content: "", Line: 2},
					{Content: "func old() {}", Line: 3},
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
					{Content: "\tnew implementation", Line: 10},
					{Content: "\tadditional line", Line: 11},
				},
				RemovedLines: []DiffLine{
					{Content: "\told implementation", Line: 10},
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
					{Content: "\tfmt.Printf(\"Hi %s!\\n\", name)", Line: 1},
				},
				RemovedLines: []DiffLine{
					{Content: "\tfmt.Printf(\"Hello %s\\n\", name)", Line: 1},
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
					{Content: "},", Line: 1},
				},
				RemovedLines: []DiffLine{
					{Content: "}", Line: 1},
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
					{Content: "", Line: 1},
					{Content: " ", Line: 2},
				},
				RemovedLines: []DiffLine{
					{Content: "", Line: 1},
					{Content: " ", Line: 2},
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
					{Content: "import \"fmt\"", Line: 1},
					{Content: "\t\tfmt.Println(\"new\")", Line: 11},
					{Content: "\t\tfmt.Println(\"extra\")", Line: 12},
				},
				RemovedLines: []DiffLine{
					{Content: "\t\tfmt.Println(\"old\")", Line: 10},
				},
			},
		},
		{
			name: "malformed diff without hunk headers",
			gitDiffOutput: `diff --git a/bad.go b/bad.go
index 123..456 100644
--- a/bad.go
+++ b/bad.go
+added line without hunk header
-removed line without hunk header`,
			wantFileDiff: FileDiff{
				AddedLines:   nil, // Should be empty since no valid hunk header
				RemovedLines: nil, // Should be empty since no valid hunk header
			},
		},
		{
			name: "malformed hunk header",
			gitDiffOutput: `diff --git a/bad.go b/bad.go
index 123..456 100644
--- a/bad.go
+++ b/bad.go
@@ invalid hunk header @@
+added line after invalid header
-removed line after invalid header`,
			wantFileDiff: FileDiff{
				AddedLines:   nil, // Should be empty since hunk header is invalid
				RemovedLines: nil, // Should be empty since hunk header is invalid
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

func TestValidateFilePath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantValid bool
	}{
		{
			name:      "valid relative path",
			path:      "src/main.go",
			wantValid: true,
		},
		{
			name:      "empty path",
			path:      "",
			wantValid: false,
		},
		{
			name:      "path traversal attempt",
			path:      "../../../etc/passwd",
			wantValid: false,
		},
		{
			name:      "absolute path",
			path:      "/etc/passwd",
			wantValid: false,
		},
		{
			name:      "path with shell metacharacters",
			path:      "file; rm -rf /",
			wantValid: false,
		},
		{
			name:      "path with backticks",
			path:      "file`whoami`.go",
			wantValid: false,
		},
		{
			name:      "path with pipes",
			path:      "file|cat /etc/passwd",
			wantValid: false,
		},
		{
			name:      "path with spaces in filename",
			path:      "my file.go", // Invalid due to spaces (potential shell injection)
			wantValid: false,
		},
		{
			name:      "valid deeply nested path",
			path:      "src/pkg/utils/helper.go",
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid := validateFilePath(tt.path)
			require.Equal(t, tt.wantValid, gotValid)
		})
	}
}

func TestValidateGitRef(t *testing.T) {
	tests := []struct {
		name      string
		ref       string
		wantValid bool
	}{
		{
			name:      "valid branch name",
			ref:       "main",
			wantValid: true,
		},
		{
			name:      "valid commit hash",
			ref:       "abc123def456",
			wantValid: true,
		},
		{
			name:      "empty ref",
			ref:       "",
			wantValid: false,
		},
		{
			name:      "ref with shell metacharacters",
			ref:       "branch; rm -rf /",
			wantValid: false,
		},
		{
			name:      "ref with whitespace",
			ref:       "my branch",
			wantValid: false,
		},
		{
			name:      "ref with path traversal",
			ref:       "../main",
			wantValid: false,
		},
		{
			name:      "ref starting with slash",
			ref:       "/main",
			wantValid: false,
		},
		{
			name:      "ref ending with slash",
			ref:       "main/",
			wantValid: false,
		},
		{
			name:      "valid feature branch",
			ref:       "feature/new-diff-parsing",
			wantValid: true,
		},
		{
			name:      "ref with control characters",
			ref:       "main\x00",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid := validateGitRef(tt.ref)
			require.Equal(t, tt.wantValid, gotValid)
		})
	}
}

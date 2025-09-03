package runner

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"strings"

	danger "github.com/danger/golang"
	dangerJs "github.com/danger/golang/danger-js"
)

const dangerURLPrefix = "danger://dsl/"

// Run reads the danger DSL URL from stdin, invokes the Go dangerfile as a
// plugin, and then writes the results JSON to stdout.
func Run() {
	in := readAll()
	in = strings.TrimSpace(in)

	if !strings.HasPrefix(in, dangerURLPrefix) {
		log.Fatalf("did not receive a DSL URL")
	}

	jsonPath := strings.Replace(in, dangerURLPrefix, "", 1)
	jsonBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		log.Fatalf("failed to read JSON file at %s", jsonPath)
	}

	var jsonData struct {
		Danger dangerJs.DSLData `json:"danger"`
	}
	err = json.Unmarshal(jsonBytes, &jsonData)
	if err != nil {
		fmt.Println("JSON\n", string(jsonBytes))
		log.Fatalf("failed to unmarshal DSL JSON: %s", err.Error())
	}

	dangerFile := "dangerfile.go"
	// TODO: Find a way to build dangerfile.go that is in project's root... will
	// have to copy along go.mod & go.sum or create new ones in temp directory.
	// TODO: Take -d/--dangerfile arg into account
	libPath, clearTempDir, err := buildPlugin(dangerFile)
	if err != nil {
		log.Fatalf("building plugin from dangerfile: %s", err.Error())
	}
	defer func() { _ = clearTempDir() }()

	fn, err := loadPlugin(libPath)
	if err != nil {
		log.Fatalf("loading dangerfile plugin: %s", err.Error())
	}

	d := danger.New()
	fn(d, jsonData.Danger.ToInterface())
	respJSON, err := d.Results()
	if err != nil {
		log.Fatalf("marshalling response: %s", err.Error())
	}
	fmt.Print(respJSON)
}

// buildPlugin builds the plugin and stores the artifacts in a temporary
// directory. If the function succeeds the caller can clear the temporary
// directory with the returned callback.
func buildPlugin(dangerFilePath string) (string, func() error, error) {
	_, err := os.Stat(dangerFilePath)
	if os.IsNotExist(err) {
		return "", nil, fmt.Errorf("`%s` does not exist", dangerFilePath)
	} else if err != nil {
		return "", nil, fmt.Errorf("getting file state: %w", err)
	}

	// Create a temporary directory to build the plugin in
	tempDir, err := os.MkdirTemp("", "danger-go-build-")
	if err != nil {
		return "", nil, fmt.Errorf("creating temp directory: %w", err)
	}
	clearTempDir := func() error {
		return os.RemoveAll(tempDir)
	}

	outputFile := filepath.Join(tempDir, "dangerfile.so")

	cmd := exec.Command("go", "build", "-o", outputFile, "-buildmode=plugin", dangerFilePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Building dangerfile plugin using `%s`\n", dangerFilePath)
	err = cmd.Run()
	if err != nil {
		_ = clearTempDir()
		return "", nil, err
	}
	return outputFile, clearTempDir, nil
}

type MainFunc = func(d *danger.T, pr danger.DSL)

func loadPlugin(libPath string) (MainFunc, error) {
	fmt.Println("Loading dangerfile plugin:", libPath)

	p, err := plugin.Open(libPath)
	if err != nil {
		return nil, err
	}

	dangerSymbol, err := p.Lookup("Run")
	if err != nil {
		return nil, err
	}

	dangerFn, ok := dangerSymbol.(MainFunc)
	if !ok {
		return nil, errors.New("failed to cast DSL function")
	}

	return dangerFn, nil
}

// readAll reads everything on stdin until io.EOF and returns the result
func readAll() string {
	reader := bufio.NewReader(os.Stdin)
	var bb bytes.Buffer
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("reading stdin: %s", err.Error())
		}
		bb.Write(line)
	}
	return bb.String()
}

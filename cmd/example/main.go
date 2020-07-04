// package main is used to showcase an example use of workflow
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/percybolmer/workflow"
	fileprocessors "github.com/percybolmer/workflow/processors/file-processors"
	"github.com/percybolmer/workflow/processors/processmanager"

	_ "github.com/percybolmer/workflow/processors/terminal-processors"
)

func main() {

	go ReadAndPrint()
	//go WithProcessMananger()

	//go WithoutProcessManager()

	time.Sleep(5 * time.Second)

}

// ReadAndPrint is a example flow that reads files
func ReadAndPrint() {
	w := workflow.NewWorkflow("file_printer_stdout")

	readproc, err := processmanager.GetProcessor("ReadFile")
	if err != nil {
		panic(err)
	}
	readproc.SetProperty("remove_after", false)
	readproc.SetProperty("path", "files/csv.txt")
	stdoutProc, err := processmanager.GetProcessor("Stdout")
	if err != nil {
		panic(err)
	}
	stdoutProc.SetProperty("forward", true)

	stdoutProc2, err := processmanager.GetProcessor("Stdout")
	if err != nil {
		panic(err)
	}

	w.AddProcessor(readproc, stdoutProc, stdoutProc2)
	err = w.Start()
	if err != nil {
		panic(err)
	}
	time.Sleep(2 * time.Second)
	mets := readproc.GetMetrics()
	for _, m := range mets {
		fmt.Printf("%s: %v", m.Name, m.Value)
	}
}

// WithProcessMananger shows examples of using the processmanager
func WithProcessMananger() {
	w := workflow.NewWorkflow("file_mover")

	listdirProc, err := processmanager.GetProcessor("ListDirectory")
	if err != nil {
		panic(err)
	}
	readproc, err := processmanager.GetProcessor("ReadFile")
	if err != nil {
		panic(err)
	}
	writeproc, err := processmanager.GetProcessor("WriteFile")
	if err != nil {
		panic(err)
	}
	csvproc, err := processmanager.GetProcessor("ParseCsv")
	if err != nil {
		panic(err)
	}

	readproc.SetProperty("remove_after", false)

	writeproc.SetProperty("path", "csvAsMap")
	writeproc.SetProperty("append", true)

	listdirProc.SetProperty("path", "files/")
	w.AddProcessor(listdirProc)
	w.AddProcessor(readproc)
	w.AddProcessor(csvproc)
	w.AddProcessor(writeproc)

	err = w.Start()
	if err != nil {
		panic(err)
	}

	time.Sleep(3 * time.Second)
	mets := readproc.GetMetrics()
	for _, m := range mets {
		fmt.Printf("%s: %v", m.Name, m.Value)
	}
}

// WithoutProcessManager shows how u can setup a flow wihtout processmanager related
func WithoutProcessManager() {
	w := workflow.NewWorkflow("file_mover")

	f, err := os.Create("thisexample.txt")
	if err != nil {
		panic(err)
	}
	f.Write([]byte(`Hello world`))
	reader := fileprocessors.NewReadFile()
	reader.SetProperty("remove_after", true)
	reader.SetProperty("filepath", "thisexample.txt")

	writer := fileprocessors.NewWriteFile()
	writer.SetProperty("path", "here")
	writer.SetProperty("append", true)

	w.AddProcessor(reader)
	w.AddProcessor(writer)

	w.Start()

	time.Sleep(2 * time.Second)

}

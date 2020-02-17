package readers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/percybolmer/filewatcher"
	"github.com/percybolmer/workflow/flow"
)

const (
	// FileReaderType is a const representation of Payloads read through Files
	FileReaderType = "file"
)

var (
	//ErrInvalidPath is thrown when the path for a file is not correct
	ErrInvalidPath error = errors.New("The path provided is not a proper path to a file or directory")
	//ErrBadWriteData is thrown when the size written to file is not the same as the payload
	ErrBadWriteData error = errors.New("The size written to file does not match the payload")
)

//FileReader is used to read a file into payload
// TODO break out FileWriter struct into its own?
type FileReader struct {
	Path            string `json:"path"`
	RemoveAfterRead bool   `json:"removefiles"`
	AppendTo        bool   `json:"append"`
}

// ReadFile will read the bytes from a file and set them as the current payload
func ReadFile(inflow *flow.Flow) {
	confByte := inflow.GetConfiguration()

	fr := FileReader{}

	err := json.Unmarshal(confByte, &fr)
	if err != nil {
		inflow.Log(err)
		return
	}

	payload, err := fr.Read(fr.Path)
	if err != nil {
		inflow.Log(err)
		return
	}
	// Add stats about the payload
	inflow.Statistics.AddStat("egress_bytes", len(payload))
	inflow.Statistics.AddStat("egress_flows", 1)

	outChan := make(chan flow.Payload, 1)
	inflow.SetEgressChannel(outChan)

	output := &flow.BasePayload{}
	output.SetPayload(payload)
	output.SetSource(fr.Path)
	outChan <- output

}

// WriteFile will take a Flows Payload and write it to file
func WriteFile(inflow *flow.Flow) {
	confByte := inflow.GetConfiguration()

	fr := FileReader{}

	err := json.Unmarshal(confByte, &fr)
	if err != nil {
		inflow.Log(err)
		return
	}

	if fr.Path == "" {
		inflow.Log(ErrInvalidPath)
		return
	}
	outChan := make(chan flow.Payload)
	inflow.SetEgressChannel(outChan)
	wg := inflow.GetWaitGroup()
	go func() {
		defer wg.Done()
		wg.Add(1)
		for {
			select {
			case newflow := <-inflow.GetIngressChannel():
				// TODO add Epoch timestamp for unique names
				inflow.Statistics.AddStat("ingress_flows", 1)
				inflow.Statistics.AddStat("ingress_bytes", len(newflow.GetPayload()))
				file := filepath.Base(newflow.GetSource())
				err := fr.WriteFile(fmt.Sprintf("%s/%s", fr.Path, file), newflow.GetPayload())
				if err != nil {
					inflow.Log(err)
					continue
				}
				// Add stats about the payload
				inflow.Statistics.AddStat("egress_bytes", len(newflow.GetPayload()))
				inflow.Statistics.AddStat("egress_flows", 1)
				outChan <- newflow
			case <-inflow.StopChannel:
				return
			}
		}
	}()
}

// WriteFile is used to write payloads to files
// If the file exists it will use  the
// append setting to check wether to overwrite or append to file
func (fr *FileReader) WriteFile(path string, payload []byte) error {
	if fr.AppendTo {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		n, err := f.Write(payload)
		if err != nil {
			return err
		}
		if n != len(payload) {
			return ErrBadWriteData
		}
		return nil
	}
	return ioutil.WriteFile(path, payload, 0644)

}

// MonitorDirectory is used to read from a directory for a given time
func MonitorDirectory(inflow *flow.Flow) {
	confByte := inflow.GetConfiguration()
	fr := FileReader{}

	err := json.Unmarshal(confByte, &fr)
	if err != nil {
		inflow.Log(err)
		return
	}
	// Make sure directory exists
	if _, err := os.Stat(fr.Path); os.IsNotExist(err) {
		inflow.Log(err)
		return
	}
	filechannel := make(chan string)
	watcher := filewatcher.NewFileWatcher()
	watcher.ChangeExecutionTime(1)

	wg := inflow.GetWaitGroup()

	go watcher.WatchDirectory(filechannel, fr.Path)
	folderPath := fr.Path
	egressChannel := make(chan flow.Payload)
	inflow.SetEgressChannel(egressChannel)
	// Start a goroutine to watch over the filechannel and Ingest the new Files
	go func(filechannel chan string, inflow *flow.Flow, egressChannel chan flow.Payload) {
		defer wg.Done()
		wg.Add(1)
		for {
			select {
			case newFile := <-filechannel:
				fmt.Printf("flow: %+v \n stat:%+v", inflow, inflow.Statistics)
				inflow.Statistics.AddStat("ingress_flows", 1)
				file := filepath.Base(newFile)
				var filePath string
				if strings.HasSuffix(folderPath, "/") {
					filePath = fmt.Sprintf("%s%s", folderPath, file)
				} else {
					filePath = fmt.Sprintf("%s/%s", folderPath, file)
				}
				bytes, err := fr.Read(filePath)
				if err != nil {
					inflow.Log(err)
					continue
				}
				if len(bytes) != 0 {
					// Add stats about the payload
					inflow.Statistics.AddStat("egress_bytes", len(bytes))
					inflow.Statistics.AddStat("egress_flows", 1)
					payload := &flow.BasePayload{}
					payload.SetSource(filePath)
					payload.SetPayload(bytes)
					egressChannel <- payload
				}
				if fr.RemoveAfterRead {
					os.Remove(filePath)
				}
			case <-inflow.StopChannel:
				return
			}

		}
	}(filechannel, inflow, egressChannel)
	return
}

// Read is used to read a file and return the Byte array of the value
func (fr *FileReader) Read(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		file.Close()
		if fr.RemoveAfterRead {
			os.Remove(path)
		}
	}()
	return ioutil.ReadAll(file)
}

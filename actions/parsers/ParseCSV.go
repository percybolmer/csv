// Package parsers is generated by actiongenerator tooling
// Make sure to insert real Description here
package parsers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"strings"

	"github.com/perbol/workflow/payload"
	"github.com/perbol/workflow/property"
	"github.com/perbol/workflow/register"
)

var (
	// DefaultDelimiter is the delimiter used if nothing is set
	DefaultDelimiter = ","
	// DefaultHeaderLength is the headerlength to use if nothing else is set
	DefaultHeaderLength = 1
	// DefaultSkipRows is the default rows to skip if nothing is set
	DefaultSkipRows = 0

	//ErrNotCsv is triggered when the input file is not proper csv
	ErrNotCsv error = errors.New("this is not a proper csv file")
	//ErrHeaderMismatch is triggered when header is longer than CSV records
	ErrHeaderMismatch error = errors.New("the header is not the same size as the records")
)

// ParseCSV is used to parse CSV files, expects whole payloads
type ParseCSV struct {
	// Cfg is values needed to properly run the Handle func
	Cfg  *property.Configuration `json:"configs" yaml:"configs"`
	Name string                  `json:"name" yaml:"name"`
	// delimiter is the character to use for delimiting
	delimiter string
	// headerlength is a int that is used for the base of the header, some files has duplicate headers etc
	headerlength int
	// skiprows is used to skip some rows if there is excess rows in the file
	skiprows int

	subscriptionless bool
}

func init() {
	register.Register("ParseCSV", NewParseCSVAction())
}

// NewParseCSVAction generates a new ParseCSV action
func NewParseCSVAction() *ParseCSV {
	act := &ParseCSV{
		Cfg: &property.Configuration{
			Properties: make([]*property.Property, 0),
		},
		Name:         "ParseCSV",
		delimiter:    DefaultDelimiter,
		headerlength: DefaultHeaderLength,
		skiprows:     DefaultSkipRows,
	}
	act.Cfg.AddProperty("delimiter", "The character or string to use as a Delimiter", false)
	act.Cfg.AddProperty("headerlength", "How many rows the header is", false)
	act.Cfg.AddProperty("skiprows", "How many rows will be skipped in each file before starting to process", false)
	return act
}

// GetActionName is used to retrun a unqiue string name
func (a *ParseCSV) GetActionName() string {
	return a.Name
}

// Handle will go through a CSV payload and output all the CSV rows
func (a ParseCSV) Handle(input payload.Payload) ([]payload.Payload, error) {
	buf := bytes.NewBuffer(input.GetPayload())

	scanner := bufio.NewScanner(buf)
	// Index keeps track of Line index in payload
	var index int

	header := make([]string, 0)
	result := make([]payload.Payload, 0)

	for scanner.Scan() {
		line := scanner.Text()
		// Handle skiprows
		if index < a.skiprows {
			index++
			continue
		}

		// Handle Header row
		values := strings.Split(line, a.delimiter)
		if len(values) <= 1 {
			return nil, ErrNotCsv
		}

		// Handle Unique Cases of header rows longer than 1 line

		if index < (a.skiprows + a.headerlength) {
			header = append(header, values...)
			index++
			continue
		}

		// Make sure header is no longer than current values
		if len(header) != len(values) {
			return nil, ErrHeaderMismatch
		}
		// Handle the CSV ROW as a Map of string, should this be interface?
		newRow := &CsvRow{
			Payload: make(map[string]string, len(values)),
		}
		for i, value := range values {
			newRow.Payload[header[i]] = value
		}
		result = append(result, newRow)
	}
	return result, nil
}

// ValidateConfiguration is used to see that all needed configurations are assigned before starting
func (a *ParseCSV) ValidateConfiguration() (bool, []string) {
	// Check if Cfgs are there as needed
	delimiterProp := a.Cfg.GetProperty("delimiter")
	headerProp := a.Cfg.GetProperty("headerlength")
	skiprowProp := a.Cfg.GetProperty("skiprows")

	missing := make([]string, 0)

	if delimiterProp == nil || headerProp == nil || skiprowProp == nil {
		// Abit lazy here and just return all 3 props
		missing = append(missing, "delimiter", "headerlength", "skiprows")
		return false, missing
	}

	if delimiterProp.Value != nil {
		a.delimiter = delimiterProp.String()
	}
	if headerProp.Value != nil {
		headerlength, err := headerProp.Int()
		if err != nil {
			return false, nil
		}
		a.headerlength = headerlength
	}
	if skiprowProp.Value != nil {
		skiprow, err := skiprowProp.Int()
		if err != nil {
			return false, nil
		}

		a.skiprows = skiprow
	}
	return true, nil
}

// GetConfiguration will return the CFG for the action
func (a *ParseCSV) GetConfiguration() *property.Configuration {
	return a.Cfg
}

// Subscriptionless will return true/false if the action is genereating payloads itself
func (a *ParseCSV) Subscriptionless() bool {
	return a.subscriptionless
}

//CsvRow is a struct representing Csv data as a map
//Its also a part of the Payload interface
type CsvRow struct {
	Payload map[string]string `json:"payload"`
	Source  string            `json:"source"`
	Error   error             `json:"error"`
}

// GetPayloadLength will return the payload X Bytes
func (nf *CsvRow) GetPayloadLength() float64 {
	data, err := json.Marshal(nf.Payload)
	if err != nil {
		nf.Error = err
	}
	return float64(len(data))
}

// GetPayload is used to return an actual value for the Flow
func (nf *CsvRow) GetPayload() []byte {
	data, err := json.Marshal(nf.Payload)
	if err != nil {
		nf.Error = err
	}
	return data
}

//SetPayload will change the value of the Flow
func (nf *CsvRow) SetPayload(newpayload []byte) {
	nf.Error = json.Unmarshal(newpayload, &nf.Payload)
}

//GetSource will return the source of the flow
func (nf *CsvRow) GetSource() string {
	return nf.Source
}

//SetSource will change the value of the configured source
//The source value should represent something that makes it possible to traceback
//Errors, so for files etc its the filename.
func (nf *CsvRow) SetSource(s string) {
	nf.Source = s
}

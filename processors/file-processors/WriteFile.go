// package fileprocessors is generated by generate-processor tooling
// Make sure to insert real Description here
package fileprocessors

import (
    "context"
    "errors"
    "github.com/percybolmer/workflow/failure"
    "github.com/percybolmer/workflow/metric"
    "github.com/percybolmer/workflow/payload"
    "github.com/percybolmer/workflow/processors/processmanager"
    "github.com/percybolmer/workflow/properties"
    "github.com/percybolmer/workflow/relationships"
    "io/ioutil"
    "os"
)
// WriteFile is used to $INSERT DESCRIPTION
type WriteFile struct{
    Name     string
    running  bool
    cancel   context.CancelFunc
    ingress  relationships.PayloadChannel
    egress   relationships.PayloadChannel
    failures relationships.FailurePipe
    *properties.PropertyMap
    *metric.Metrics
}

var (
    //ErrBadWriteData is thrown when the size written to file is not the same as the payload
    ErrBadWriteData error = errors.New("the size written to file does not match the payload")
    //ErrFileExists is when trying to write to Files that already exist, but Append is set to false
    ErrFileExists = errors.New("trying to write to file that already exists, but append is false")
)

func init() {
    err := processmanager.RegisterProcessor("WriteFile", NewWriteFileInterface)
    if err != nil {
        panic(err)
    }
}
// NewWriteFileInterface is used to register ReadFile
func NewWriteFileInterface() interface{}{
    return NewWriteFile()
}


// NewWriteFile is used to initialize and generate a new processor
func NewWriteFile() *WriteFile {
    proc := &WriteFile{
        egress: make(relationships.PayloadChannel, 1000),
        PropertyMap: properties.NewPropertyMap(),
        Metrics: metric.NewMetrics(),
    }

    // Add Required Props
    proc.AddRequirement("append", "path")
    return proc
}

// Initialize will make sure all needed Properties and Metrics are generated
func (proc *WriteFile) Initialize() error {

    // Make sure Properties are there
    ok, _ := proc.ValidateProperties()
    if !ok {
        return properties.ErrRequiredPropertiesNotFulfilled
    }
    // If you need to read data from Properties and add to your Processor struct, this is the place to do it
    return nil
}

// Start will spawn a goroutine that reads file and Exits either on Context.Done or When processing is finished
func (proc *WriteFile) Start(ctx context.Context) error {
    if proc.running {
        return failure.ErrAlreadyRunning
    }
    // Uncomment if u need to Processor to require an Ingress relationship
    if proc.ingress == nil {
        return failure.ErrIngressRelationshipNeeded
    }

    proc.running = true
    append, err := proc.GetProperty("append").Bool()
    if err != nil {
        return err
    }
    path := proc.GetProperty("path").String()
    // context will be used to spawn a Cancel func
    c, cancel := context.WithCancel(ctx)
    proc.cancel = cancel
    go func() {
        for {
            select {
                case payload := <-proc.ingress:
                    // Do your processing here
                    err := proc.Write(path, append, payload)
                    if err != nil {
                        proc.AddMetric("failures", "the number of failures that occured in the processor", 1)
                        proc.failures <- failure.Failure{
                            Err:       err,
                            Payload:   payload,
                            Processor: "WriteFile",
                        }
                    }
                    proc.AddMetric("writes", "the number of writes the processor has performed", 1)

                case <- c.Done():
                    return
            }
        }
    }()
    return nil
}
// Write is used to write down payloads to file
func (proc *WriteFile) Write(path string, append bool, payload payload.Payload) error {
    finfo, err := os.Stat(path)
    if err != nil && !os.IsNotExist(err){
        return err
    }
    // DONT WRITE TO FILES THAT EXIST IF APPEND IS FALSE
    if finfo != nil && !append{
        return ErrFileExists
    }
    if finfo != nil && finfo.IsDir(){
        // So, Write is to a Folder, No need for error, but lets create a Name for the file
        return createTempFileAndWrite(path, payload)
    }


    return appendToFileOrCreate(path, payload)
}
// createTempFileAndWrite is used to create a random filename and write to it
func createTempFileAndWrite(path string, payload payload.Payload) error {
    file,err := ioutil.TempFile(path, "WriteFile")
    if err != nil {
        return err
    }
    return write(file, payload)
}
// appendToFileOrCreate will create or append to file
func appendToFileOrCreate(path string, payload payload.Payload) error {
    file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    return write(file, payload)
}
// write takes care of writing to file and closing it
func write(file *os.File, payload payload.Payload) error {
    defer file.Close()
    n, err := file.Write(payload.GetPayload())
    if err != nil {
        return err
    }
    if n != len(payload.GetPayload()) {
        return ErrBadWriteData
    }
    return nil
}

// IsRunning will return true or false based on if the processor is currently running
func (proc *WriteFile) IsRunning() bool {
    return proc.running
}
// GetMetrics will return a bunch of generated metrics, or nil if there isn't any
func (proc *WriteFile) GetMetrics() []*metric.Metric {
    return proc.GetAllMetrics()
}
// SetFailureChannel will configure the failure channel of the Processor
func (proc *WriteFile) SetFailureChannel(fp relationships.FailurePipe) {
    proc.failures = fp
}

// Stop will stop the processing
func (proc *WriteFile) Stop() {
    if !proc.running {
        return
    }
    proc.running = false
    proc.cancel()
}
// SetIngress will change the ingress of the processor, Restart is needed before applied changes
func (proc *WriteFile) SetIngress(i relationships.PayloadChannel) {
    proc.ingress = i
    return
}
// GetEgress will return an Egress that is used to output the processed items
func (proc *WriteFile) GetEgress() relationships.PayloadChannel {
    return proc.egress
}
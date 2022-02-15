// Package payload contains all the payload related stuff
// Payload is used to hold data between Processors,
// Payload is just an interface so each Processor could create its own Struct to handle data
// as long as it fulfills our interface
package payload

import (
	"encoding"

	"github.com/percybolmer/go4data/property"
)

// Payload is a interface that will allows different Processors to send data between them in a unified fashion
type Payload interface {
	// GetPayloadLength returns the payload length in flota64
	GetPayloadLength() float64
	// GetPayload will return a byte array with the Payload from the ingress
	// Payload should be limited to 512 MB since thats the MAX cap for a redis payload
	// Also note that JSON payloads will be base64 encoded
	GetPayload() []byte
	// SetPayload will change the values of the payload
	SetPayload([]byte)
	// GetMetaData should return a configuration object that contains metadata about the payload
	GetMetaData() *property.Configuration
	// Force Payloads to also be part of the Encoding package interfaces
	// This is needed for Redis purpose
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

package domain

// SensorReader interface for generic sensor operations
type SensorReader interface {
    GetLastReading(serial string) (interface{}, error)
}

// KY026Reader specific interface for KY026 sensor
type KY026Reader interface {
    GetLastReading(serial string) (*KY026Reading, error)
    ProcessKY026Alert(alert *Alert) error
}
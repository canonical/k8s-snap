package mock

import (
	"context"
	"path/filepath"
)

// Snap is a generic mock for the snap.Snap interface.
type Snap struct {
	StartServiceCalledWith []string
	StopServiceCalledWith  []string

	WriteServiceArgumentsCalled bool

	ServiceArguments map[string]string
}

func (s *Snap) StartService(_ context.Context, service string) error {
	s.StartServiceCalledWith = append(s.StartServiceCalledWith, service)
	return nil
}

func (s *Snap) StopService(_ context.Context, service string) error {
	s.StopServiceCalledWith = append(s.StopServiceCalledWith, service)
	return nil
}

func (s *Snap) ReadServiceArguments(service string) (string, error) {
	if s.ServiceArguments == nil {
		s.ServiceArguments = make(map[string]string)
	}
	return s.ServiceArguments[service], nil
}

func (s *Snap) WriteServiceArguments(service string, b []byte) error {
	if s.ServiceArguments == nil {
		s.ServiceArguments = make(map[string]string)
	}
	s.ServiceArguments[service] = string(b)
	s.WriteServiceArgumentsCalled = true
	return nil
}

func (s *Snap) Path(parts ...string) string {
	return filepath.Join(parts...)
}

func (s *Snap) DataPath(parts ...string) string {
	return filepath.Join(parts...)
}
func (s *Snap) CommonPath(parts ...string) string {
	return filepath.Join(parts...)
}

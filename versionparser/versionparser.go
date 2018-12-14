package versionparser

import (
	"errors"
)

var (
	ErrWrongDistinction    = errors.New("Other Version Parser has a different distinction")
	ErrWrongVPType         = errors.New("Cannot compare with this Version Parser")
	ErrVersionNotSupported = errors.New("Version Parser does not support this version string")
)

type VersionParser interface {
	Distinction() string
	IsGreaterThan(other VersionParser) (bool, error)
	String() string
	Weight() int
}

type DefaultRegistry struct {
	factories []func(string) (VersionParser, error)
}

func (d *DefaultRegistry) FindForVersion(version string) VersionParser {
	for _, fac := range d.factories {
		vp, err := fac(version)
		if err == nil {
			return vp
		}
	}

	return &Unknown{raw: version}
}

func NewDefaultRegistry() *DefaultRegistry {
	return &DefaultRegistry{
		factories: []func(string) (VersionParser, error){
			majorFactory,
			majorMinorFactory,
			majorMinorPatchFactory,
			nameDateFactory,
			staticFactory,
		},
	}
}

var Registry *DefaultRegistry = NewDefaultRegistry()

func FindForVersion(version string) VersionParser {
	return Registry.FindForVersion(version)
}

package versionparser

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	majorRegexp           = regexp.MustCompile("^v?(\\d+)(-.*)?$")
	majorMinorRegexp      = regexp.MustCompile("^v?(\\d+)\\.(\\d+)(-.*)?$")
	majorMinorPatchRegexp = regexp.MustCompile("^^v?(\\d+)\\.(\\d+)\\.(\\d+)(-.*)?$")
)

type Major struct {
	build   string
	version int
	raw     string
}

func (p *Major) Distinction() string {
	return fmt.Sprintf("major%s", p.build)
}

func (p *Major) IsGreaterThan(other VersionParser) (bool, error) {
	o, ok := other.(*Major)
	if !ok {
		return false, ErrWrongVPType
	}

	return p.version > o.version, nil
}

func (p *Major) String() string {
	return p.raw
}

func (p *Major) Weight() int {
	return 80
}

func majorFactory(version string) (VersionParser, error) {
	matches := majorRegexp.FindStringSubmatch(version)
	matchCount := len(matches)
	if matchCount == 0 {
		return nil, ErrVersionNotSupported
	}

	versionInt, _ := strconv.Atoi(matches[1])
	return &Major{
		build:   matches[2],
		version: versionInt,
		raw:     version,
	}, nil
}

type MajorMinor struct {
	build string
	major int
	minor int
	raw   string
}

func (p *MajorMinor) Distinction() string {
	return fmt.Sprintf("majorMinor%s", p.build)
}

func (p *MajorMinor) IsGreaterThan(other VersionParser) (bool, error) {
	o, ok := other.(*MajorMinor)
	if !ok {
		return false, ErrWrongVPType
	}

	if p.major > o.major {
		return true, nil
	}

	if p.major == o.major {
		if p.minor > o.minor {
			return true, nil
		}
	}

	return false, nil
}

func (p *MajorMinor) String() string {
	return p.raw
}

func (p *MajorMinor) Weight() int {
	return 90
}

func majorMinorFactory(version string) (VersionParser, error) {
	matches := majorMinorRegexp.FindStringSubmatch(version)
	matchCount := len(matches)
	if matchCount == 0 {
		return nil, ErrVersionNotSupported
	}

	majorVersionInt, _ := strconv.Atoi(matches[1])
	minorVersionInt, _ := strconv.Atoi(matches[2])
	if matchCount == 1 {
		return &MajorMinor{
			major: majorVersionInt,
			minor: minorVersionInt,
			raw:   version,
		}, nil
	}

	return &MajorMinor{
		build: matches[3],
		major: majorVersionInt,
		minor: minorVersionInt,
		raw:   version,
	}, nil
}

type MajorMinorPatch struct {
	build string
	major int
	minor int
	patch int
	raw   string
}

func (p *MajorMinorPatch) Distinction() string {
	return fmt.Sprintf("majorMinorPatch%s", p.build)
}

func (p *MajorMinorPatch) IsGreaterThan(other VersionParser) (bool, error) {
	o, ok := other.(*MajorMinorPatch)
	if !ok {
		return false, ErrWrongVPType
	}

	if p.major > o.major {
		return true, nil
	}

	if p.major == o.major {
		if p.minor > o.minor {
			return true, nil
		}

		if p.minor == o.minor {
			if p.patch > o.patch {
				return true, nil
			}
		}
	}

	return false, nil
}

func (p *MajorMinorPatch) String() string {
	return p.raw
}

func (p *MajorMinorPatch) Weight() int {
	return 100
}

func majorMinorPatchFactory(version string) (VersionParser, error) {
	matches := majorMinorPatchRegexp.FindStringSubmatch(version)
	matchCount := len(matches)
	if matchCount == 0 {
		return nil, ErrVersionNotSupported
	}

	majorVersionInt, _ := strconv.Atoi(matches[1])
	minorVersionInt, _ := strconv.Atoi(matches[2])
	patchVersionInt, _ := strconv.Atoi(matches[3])
	return &MajorMinorPatch{
		build: matches[4],
		major: majorVersionInt,
		minor: minorVersionInt,
		patch: patchVersionInt,
		raw:   version,
	}, nil
}

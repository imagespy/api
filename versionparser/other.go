package versionparser

import (
	"fmt"
	"strconv"
)

type NameDate struct {
	date int
	name string
	raw  string
}

func (p *NameDate) Distinction() string {
	return fmt.Sprintf("nameDate-%s", p.name)
}

func (p *NameDate) IsGreaterThan(other VersionParser) (bool, error) {
	o, ok := other.(*NameDate)
	if !ok {
		return false, ErrWrongVPType
	}

	return p.date > o.date, nil
}

func (p *NameDate) String() string {
	return p.raw
}

func nameDateFactory(version string) (VersionParser, error) {
	matches := nameDateRegexp.FindStringSubmatch(version)
	matchCount := len(matches)
	if matchCount == 0 {
		return nil, ErrVersionNotSupported
	}

	date, _ := strconv.Atoi(matches[2])
	return &NameDate{
		date: date,
		name: matches[1],
		raw:  matches[0],
	}, nil
}

type Static struct {
	raw string
}

func (p *Static) Distinction() string {
	return fmt.Sprintf("static-%s", p.raw)
}

func (p *Static) IsGreaterThan(other VersionParser) (bool, error) {
	_, ok := other.(*Static)
	if !ok {
		return false, ErrWrongVPType
	}

	return false, nil
}

func (p *Static) String() string {
	return p.raw
}

func staticFactory(version string) (VersionParser, error) {
	_, ok := staticKnownTags[version]
	if !ok {
		return nil, ErrVersionNotSupported
	}

	return &Static{
		raw: version,
	}, nil
}

type Unknown struct {
	raw string
}

func (p *Unknown) Distinction() string {
	return fmt.Sprintf("unknown-%s", p.raw)
}

func (p *Unknown) IsGreaterThan(other VersionParser) (bool, error) {
	_, ok := other.(*Unknown)
	if !ok {
		return false, ErrWrongVPType
	}

	return false, nil
}

func (p *Unknown) String() string {
	return p.raw
}

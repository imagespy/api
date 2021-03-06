// Code generated by MockGen. DO NOT EDIT.
// Source: ./scrape/scraper.go

// Package scrape is a generated GoMock package.
package scrape

import (
	gomock "github.com/golang/mock/gomock"
	registry "github.com/imagespy/api/registry"
	reflect "reflect"
)

// MockScraper is a mock of Scraper interface
type MockScraper struct {
	ctrl     *gomock.Controller
	recorder *MockScraperMockRecorder
}

// MockScraperMockRecorder is the mock recorder for MockScraper
type MockScraperMockRecorder struct {
	mock *MockScraper
}

// NewMockScraper creates a new mock instance
func NewMockScraper(ctrl *gomock.Controller) *MockScraper {
	mock := &MockScraper{ctrl: ctrl}
	mock.recorder = &MockScraperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockScraper) EXPECT() *MockScraperMockRecorder {
	return m.recorder
}

// ScrapeImage mocks base method
func (m *MockScraper) ScrapeImage(i registry.Image) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ScrapeImage", i)
	ret0, _ := ret[0].(error)
	return ret0
}

// ScrapeImage indicates an expected call of ScrapeImage
func (mr *MockScraperMockRecorder) ScrapeImage(i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ScrapeImage", reflect.TypeOf((*MockScraper)(nil).ScrapeImage), i)
}

// ScrapeLatestImage mocks base method
func (m *MockScraper) ScrapeLatestImage(i registry.Image) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ScrapeLatestImage", i)
	ret0, _ := ret[0].(error)
	return ret0
}

// ScrapeLatestImage indicates an expected call of ScrapeLatestImage
func (mr *MockScraperMockRecorder) ScrapeLatestImage(i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ScrapeLatestImage", reflect.TypeOf((*MockScraper)(nil).ScrapeLatestImage), i)
}

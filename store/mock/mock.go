// Code generated by MockGen. DO NOT EDIT.
// Source: ./store/store.go

// Package mock is a generated GoMock package.
package mock

import (
	gomock "github.com/golang/mock/gomock"
	store "github.com/imagespy/api/store"
	reflect "reflect"
)

// MockStore is a mock of Store interface
type MockStore struct {
	ctrl     *gomock.Controller
	recorder *MockStoreMockRecorder
}

// MockStoreMockRecorder is the mock recorder for MockStore
type MockStoreMockRecorder struct {
	mock *MockStore
}

// NewMockStore creates a new mock instance
func NewMockStore(ctrl *gomock.Controller) *MockStore {
	mock := &MockStore{ctrl: ctrl}
	mock.recorder = &MockStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockStore) EXPECT() *MockStoreMockRecorder {
	return m.recorder
}

// Close mocks base method
func (m *MockStore) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (mr *MockStoreMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockStore)(nil).Close))
}

// Images mocks base method
func (m *MockStore) Images() store.ImageStore {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Images")
	ret0, _ := ret[0].(store.ImageStore)
	return ret0
}

// Images indicates an expected call of Images
func (mr *MockStoreMockRecorder) Images() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Images", reflect.TypeOf((*MockStore)(nil).Images))
}

// Layers mocks base method
func (m *MockStore) Layers() store.LayerStore {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Layers")
	ret0, _ := ret[0].(store.LayerStore)
	return ret0
}

// Layers indicates an expected call of Layers
func (mr *MockStoreMockRecorder) Layers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Layers", reflect.TypeOf((*MockStore)(nil).Layers))
}

// LayerPositions mocks base method
func (m *MockStore) LayerPositions() store.LayerPositionStore {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LayerPositions")
	ret0, _ := ret[0].(store.LayerPositionStore)
	return ret0
}

// LayerPositions indicates an expected call of LayerPositions
func (mr *MockStoreMockRecorder) LayerPositions() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LayerPositions", reflect.TypeOf((*MockStore)(nil).LayerPositions))
}

// Platforms mocks base method
func (m *MockStore) Platforms() store.PlatformStore {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Platforms")
	ret0, _ := ret[0].(store.PlatformStore)
	return ret0
}

// Platforms indicates an expected call of Platforms
func (mr *MockStoreMockRecorder) Platforms() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Platforms", reflect.TypeOf((*MockStore)(nil).Platforms))
}

// Tags mocks base method
func (m *MockStore) Tags() store.TagStore {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Tags")
	ret0, _ := ret[0].(store.TagStore)
	return ret0
}

// Tags indicates an expected call of Tags
func (mr *MockStoreMockRecorder) Tags() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Tags", reflect.TypeOf((*MockStore)(nil).Tags))
}

// Transaction mocks base method
func (m *MockStore) Transaction() (store.StoreTransaction, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Transaction")
	ret0, _ := ret[0].(store.StoreTransaction)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Transaction indicates an expected call of Transaction
func (mr *MockStoreMockRecorder) Transaction() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Transaction", reflect.TypeOf((*MockStore)(nil).Transaction))
}

// MockStoreTransaction is a mock of StoreTransaction interface
type MockStoreTransaction struct {
	ctrl     *gomock.Controller
	recorder *MockStoreTransactionMockRecorder
}

// MockStoreTransactionMockRecorder is the mock recorder for MockStoreTransaction
type MockStoreTransactionMockRecorder struct {
	mock *MockStoreTransaction
}

// NewMockStoreTransaction creates a new mock instance
func NewMockStoreTransaction(ctrl *gomock.Controller) *MockStoreTransaction {
	mock := &MockStoreTransaction{ctrl: ctrl}
	mock.recorder = &MockStoreTransactionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockStoreTransaction) EXPECT() *MockStoreTransactionMockRecorder {
	return m.recorder
}

// Close mocks base method
func (m *MockStoreTransaction) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close
func (mr *MockStoreTransactionMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockStoreTransaction)(nil).Close))
}

// Images mocks base method
func (m *MockStoreTransaction) Images() store.ImageStore {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Images")
	ret0, _ := ret[0].(store.ImageStore)
	return ret0
}

// Images indicates an expected call of Images
func (mr *MockStoreTransactionMockRecorder) Images() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Images", reflect.TypeOf((*MockStoreTransaction)(nil).Images))
}

// Layers mocks base method
func (m *MockStoreTransaction) Layers() store.LayerStore {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Layers")
	ret0, _ := ret[0].(store.LayerStore)
	return ret0
}

// Layers indicates an expected call of Layers
func (mr *MockStoreTransactionMockRecorder) Layers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Layers", reflect.TypeOf((*MockStoreTransaction)(nil).Layers))
}

// LayerPositions mocks base method
func (m *MockStoreTransaction) LayerPositions() store.LayerPositionStore {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LayerPositions")
	ret0, _ := ret[0].(store.LayerPositionStore)
	return ret0
}

// LayerPositions indicates an expected call of LayerPositions
func (mr *MockStoreTransactionMockRecorder) LayerPositions() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LayerPositions", reflect.TypeOf((*MockStoreTransaction)(nil).LayerPositions))
}

// Platforms mocks base method
func (m *MockStoreTransaction) Platforms() store.PlatformStore {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Platforms")
	ret0, _ := ret[0].(store.PlatformStore)
	return ret0
}

// Platforms indicates an expected call of Platforms
func (mr *MockStoreTransactionMockRecorder) Platforms() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Platforms", reflect.TypeOf((*MockStoreTransaction)(nil).Platforms))
}

// Tags mocks base method
func (m *MockStoreTransaction) Tags() store.TagStore {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Tags")
	ret0, _ := ret[0].(store.TagStore)
	return ret0
}

// Tags indicates an expected call of Tags
func (mr *MockStoreTransactionMockRecorder) Tags() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Tags", reflect.TypeOf((*MockStoreTransaction)(nil).Tags))
}

// Transaction mocks base method
func (m *MockStoreTransaction) Transaction() (store.StoreTransaction, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Transaction")
	ret0, _ := ret[0].(store.StoreTransaction)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Transaction indicates an expected call of Transaction
func (mr *MockStoreTransactionMockRecorder) Transaction() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Transaction", reflect.TypeOf((*MockStoreTransaction)(nil).Transaction))
}

// Commit mocks base method
func (m *MockStoreTransaction) Commit() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Commit")
	ret0, _ := ret[0].(error)
	return ret0
}

// Commit indicates an expected call of Commit
func (mr *MockStoreTransactionMockRecorder) Commit() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Commit", reflect.TypeOf((*MockStoreTransaction)(nil).Commit))
}

// Rollback mocks base method
func (m *MockStoreTransaction) Rollback() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Rollback")
	ret0, _ := ret[0].(error)
	return ret0
}

// Rollback indicates an expected call of Rollback
func (mr *MockStoreTransactionMockRecorder) Rollback() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Rollback", reflect.TypeOf((*MockStoreTransaction)(nil).Rollback))
}

// MockImageStore is a mock of ImageStore interface
type MockImageStore struct {
	ctrl     *gomock.Controller
	recorder *MockImageStoreMockRecorder
}

// MockImageStoreMockRecorder is the mock recorder for MockImageStore
type MockImageStoreMockRecorder struct {
	mock *MockImageStore
}

// NewMockImageStore creates a new mock instance
func NewMockImageStore(ctrl *gomock.Controller) *MockImageStore {
	mock := &MockImageStore{ctrl: ctrl}
	mock.recorder = &MockImageStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockImageStore) EXPECT() *MockImageStoreMockRecorder {
	return m.recorder
}

// Create mocks base method
func (m *MockImageStore) Create(i *store.Image) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", i)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create
func (mr *MockImageStoreMockRecorder) Create(i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockImageStore)(nil).Create), i)
}

// FindByLayerIDHavingLayerCountGreaterThan mocks base method
func (m *MockImageStore) FindByLayerIDHavingLayerCountGreaterThan(layerID, count int) ([]*store.Image, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByLayerIDHavingLayerCountGreaterThan", layerID, count)
	ret0, _ := ret[0].([]*store.Image)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByLayerIDHavingLayerCountGreaterThan indicates an expected call of FindByLayerIDHavingLayerCountGreaterThan
func (mr *MockImageStoreMockRecorder) FindByLayerIDHavingLayerCountGreaterThan(layerID, count interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByLayerIDHavingLayerCountGreaterThan", reflect.TypeOf((*MockImageStore)(nil).FindByLayerIDHavingLayerCountGreaterThan), layerID, count)
}

// Get mocks base method
func (m *MockImageStore) Get(o store.ImageGetOptions) (*store.Image, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", o)
	ret0, _ := ret[0].(*store.Image)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockImageStoreMockRecorder) Get(o interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockImageStore)(nil).Get), o)
}

// List mocks base method
func (m *MockImageStore) List(o store.ImageListOptions) ([]*store.Image, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", o)
	ret0, _ := ret[0].([]*store.Image)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockImageStoreMockRecorder) List(o interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockImageStore)(nil).List), o)
}

// Update mocks base method
func (m *MockImageStore) Update(i *store.Image) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", i)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockImageStoreMockRecorder) Update(i interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockImageStore)(nil).Update), i)
}

// MockLayerStore is a mock of LayerStore interface
type MockLayerStore struct {
	ctrl     *gomock.Controller
	recorder *MockLayerStoreMockRecorder
}

// MockLayerStoreMockRecorder is the mock recorder for MockLayerStore
type MockLayerStoreMockRecorder struct {
	mock *MockLayerStore
}

// NewMockLayerStore creates a new mock instance
func NewMockLayerStore(ctrl *gomock.Controller) *MockLayerStore {
	mock := &MockLayerStore{ctrl: ctrl}
	mock.recorder = &MockLayerStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockLayerStore) EXPECT() *MockLayerStoreMockRecorder {
	return m.recorder
}

// Create mocks base method
func (m *MockLayerStore) Create(l *store.Layer) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", l)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create
func (mr *MockLayerStoreMockRecorder) Create(l interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockLayerStore)(nil).Create), l)
}

// Get mocks base method
func (m *MockLayerStore) Get(o store.LayerGetOptions) (*store.Layer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", o)
	ret0, _ := ret[0].(*store.Layer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockLayerStoreMockRecorder) Get(o interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockLayerStore)(nil).Get), o)
}

// List mocks base method
func (m *MockLayerStore) List(o store.LayerListOptions) ([]*store.Layer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", o)
	ret0, _ := ret[0].([]*store.Layer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockLayerStoreMockRecorder) List(o interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockLayerStore)(nil).List), o)
}

// Update mocks base method
func (m *MockLayerStore) Update(l *store.Layer) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", l)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockLayerStoreMockRecorder) Update(l interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockLayerStore)(nil).Update), l)
}

// MockLayerPositionStore is a mock of LayerPositionStore interface
type MockLayerPositionStore struct {
	ctrl     *gomock.Controller
	recorder *MockLayerPositionStoreMockRecorder
}

// MockLayerPositionStoreMockRecorder is the mock recorder for MockLayerPositionStore
type MockLayerPositionStoreMockRecorder struct {
	mock *MockLayerPositionStore
}

// NewMockLayerPositionStore creates a new mock instance
func NewMockLayerPositionStore(ctrl *gomock.Controller) *MockLayerPositionStore {
	mock := &MockLayerPositionStore{ctrl: ctrl}
	mock.recorder = &MockLayerPositionStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockLayerPositionStore) EXPECT() *MockLayerPositionStoreMockRecorder {
	return m.recorder
}

// Create mocks base method
func (m *MockLayerPositionStore) Create(arg0 *store.LayerPosition) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create
func (mr *MockLayerPositionStoreMockRecorder) Create(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockLayerPositionStore)(nil).Create), arg0)
}

// List mocks base method
func (m *MockLayerPositionStore) List(o store.LayerPositionListOptions) ([]*store.LayerPosition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", o)
	ret0, _ := ret[0].([]*store.LayerPosition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockLayerPositionStoreMockRecorder) List(o interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockLayerPositionStore)(nil).List), o)
}

// MockPlatformStore is a mock of PlatformStore interface
type MockPlatformStore struct {
	ctrl     *gomock.Controller
	recorder *MockPlatformStoreMockRecorder
}

// MockPlatformStoreMockRecorder is the mock recorder for MockPlatformStore
type MockPlatformStoreMockRecorder struct {
	mock *MockPlatformStore
}

// NewMockPlatformStore creates a new mock instance
func NewMockPlatformStore(ctrl *gomock.Controller) *MockPlatformStore {
	mock := &MockPlatformStore{ctrl: ctrl}
	mock.recorder = &MockPlatformStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPlatformStore) EXPECT() *MockPlatformStoreMockRecorder {
	return m.recorder
}

// Create mocks base method
func (m *MockPlatformStore) Create(arg0 *store.Platform) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create
func (mr *MockPlatformStoreMockRecorder) Create(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockPlatformStore)(nil).Create), arg0)
}

// Get mocks base method
func (m *MockPlatformStore) Get(o store.PlatformGetOptions) (*store.Platform, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", o)
	ret0, _ := ret[0].(*store.Platform)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockPlatformStoreMockRecorder) Get(o interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockPlatformStore)(nil).Get), o)
}

// List mocks base method
func (m *MockPlatformStore) List(o store.PlatformListOptions) ([]*store.Platform, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", o)
	ret0, _ := ret[0].([]*store.Platform)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockPlatformStoreMockRecorder) List(o interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockPlatformStore)(nil).List), o)
}

// MockTagStore is a mock of TagStore interface
type MockTagStore struct {
	ctrl     *gomock.Controller
	recorder *MockTagStoreMockRecorder
}

// MockTagStoreMockRecorder is the mock recorder for MockTagStore
type MockTagStoreMockRecorder struct {
	mock *MockTagStore
}

// NewMockTagStore creates a new mock instance
func NewMockTagStore(ctrl *gomock.Controller) *MockTagStore {
	mock := &MockTagStore{ctrl: ctrl}
	mock.recorder = &MockTagStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTagStore) EXPECT() *MockTagStoreMockRecorder {
	return m.recorder
}

// Create mocks base method
func (m *MockTagStore) Create(arg0 *store.Tag) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create
func (mr *MockTagStoreMockRecorder) Create(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockTagStore)(nil).Create), arg0)
}

// Get mocks base method
func (m *MockTagStore) Get(o store.TagGetOptions) (*store.Tag, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", o)
	ret0, _ := ret[0].(*store.Tag)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockTagStoreMockRecorder) Get(o interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockTagStore)(nil).Get), o)
}

// List mocks base method
func (m *MockTagStore) List(o store.TagListOptions) ([]*store.Tag, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", o)
	ret0, _ := ret[0].([]*store.Tag)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List
func (mr *MockTagStoreMockRecorder) List(o interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockTagStore)(nil).List), o)
}

// Update mocks base method
func (m *MockTagStore) Update(arg0 *store.Tag) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update
func (mr *MockTagStoreMockRecorder) Update(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockTagStore)(nil).Update), arg0)
}

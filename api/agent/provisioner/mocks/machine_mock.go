// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/api/provisioner (interfaces: MachineProvisioner)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	instance "github.com/juju/juju/core/instance"
	life "github.com/juju/juju/core/life"
	status "github.com/juju/juju/core/status"
	watcher "github.com/juju/juju/core/watcher"
	params "github.com/juju/juju/rpc/params"
	names "github.com/juju/names/v4"
	version "github.com/juju/version/v2"
)

// MockMachineProvisioner is a mock of MachineProvisioner interface.
type MockMachineProvisioner struct {
	ctrl     *gomock.Controller
	recorder *MockMachineProvisionerMockRecorder
}

// MockMachineProvisionerMockRecorder is the mock recorder for MockMachineProvisioner.
type MockMachineProvisionerMockRecorder struct {
	mock *MockMachineProvisioner
}

// NewMockMachineProvisioner creates a new mock instance.
func NewMockMachineProvisioner(ctrl *gomock.Controller) *MockMachineProvisioner {
	mock := &MockMachineProvisioner{ctrl: ctrl}
	mock.recorder = &MockMachineProvisionerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMachineProvisioner) EXPECT() *MockMachineProvisionerMockRecorder {
	return m.recorder
}

// AvailabilityZone mocks base method.
func (m *MockMachineProvisioner) AvailabilityZone() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AvailabilityZone")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AvailabilityZone indicates an expected call of AvailabilityZone.
func (mr *MockMachineProvisionerMockRecorder) AvailabilityZone() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AvailabilityZone", reflect.TypeOf((*MockMachineProvisioner)(nil).AvailabilityZone))
}

// DistributionGroup mocks base method.
func (m *MockMachineProvisioner) DistributionGroup() ([]instance.Id, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DistributionGroup")
	ret0, _ := ret[0].([]instance.Id)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DistributionGroup indicates an expected call of DistributionGroup.
func (mr *MockMachineProvisionerMockRecorder) DistributionGroup() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DistributionGroup", reflect.TypeOf((*MockMachineProvisioner)(nil).DistributionGroup))
}

// EnsureDead mocks base method.
func (m *MockMachineProvisioner) EnsureDead() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureDead")
	ret0, _ := ret[0].(error)
	return ret0
}

// EnsureDead indicates an expected call of EnsureDead.
func (mr *MockMachineProvisionerMockRecorder) EnsureDead() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureDead", reflect.TypeOf((*MockMachineProvisioner)(nil).EnsureDead))
}

// Id mocks base method.
func (m *MockMachineProvisioner) Id() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Id")
	ret0, _ := ret[0].(string)
	return ret0
}

// Id indicates an expected call of Id.
func (mr *MockMachineProvisionerMockRecorder) Id() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Id", reflect.TypeOf((*MockMachineProvisioner)(nil).Id))
}

// InstanceId mocks base method.
func (m *MockMachineProvisioner) InstanceId() (instance.Id, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InstanceId")
	ret0, _ := ret[0].(instance.Id)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InstanceId indicates an expected call of InstanceId.
func (mr *MockMachineProvisionerMockRecorder) InstanceId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InstanceId", reflect.TypeOf((*MockMachineProvisioner)(nil).InstanceId))
}

// InstanceStatus mocks base method.
func (m *MockMachineProvisioner) InstanceStatus() (status.Status, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InstanceStatus")
	ret0, _ := ret[0].(status.Status)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// InstanceStatus indicates an expected call of InstanceStatus.
func (mr *MockMachineProvisionerMockRecorder) InstanceStatus() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InstanceStatus", reflect.TypeOf((*MockMachineProvisioner)(nil).InstanceStatus))
}

// KeepInstance mocks base method.
func (m *MockMachineProvisioner) KeepInstance() (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "KeepInstance")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// KeepInstance indicates an expected call of KeepInstance.
func (mr *MockMachineProvisionerMockRecorder) KeepInstance() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "KeepInstance", reflect.TypeOf((*MockMachineProvisioner)(nil).KeepInstance))
}

// Life mocks base method.
func (m *MockMachineProvisioner) Life() life.Value {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Life")
	ret0, _ := ret[0].(life.Value)
	return ret0
}

// Life indicates an expected call of Life.
func (mr *MockMachineProvisionerMockRecorder) Life() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Life", reflect.TypeOf((*MockMachineProvisioner)(nil).Life))
}

// MachineTag mocks base method.
func (m *MockMachineProvisioner) MachineTag() names.MachineTag {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MachineTag")
	ret0, _ := ret[0].(names.MachineTag)
	return ret0
}

// MachineTag indicates an expected call of MachineTag.
func (mr *MockMachineProvisionerMockRecorder) MachineTag() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MachineTag", reflect.TypeOf((*MockMachineProvisioner)(nil).MachineTag))
}

// MarkForRemoval mocks base method.
func (m *MockMachineProvisioner) MarkForRemoval() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MarkForRemoval")
	ret0, _ := ret[0].(error)
	return ret0
}

// MarkForRemoval indicates an expected call of MarkForRemoval.
func (mr *MockMachineProvisionerMockRecorder) MarkForRemoval() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarkForRemoval", reflect.TypeOf((*MockMachineProvisioner)(nil).MarkForRemoval))
}

// ModelAgentVersion mocks base method.
func (m *MockMachineProvisioner) ModelAgentVersion() (*version.Number, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ModelAgentVersion")
	ret0, _ := ret[0].(*version.Number)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ModelAgentVersion indicates an expected call of ModelAgentVersion.
func (mr *MockMachineProvisionerMockRecorder) ModelAgentVersion() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ModelAgentVersion", reflect.TypeOf((*MockMachineProvisioner)(nil).ModelAgentVersion))
}

// ProvisioningInfo mocks base method.
func (m *MockMachineProvisioner) ProvisioningInfo() (*params.ProvisioningInfoV10, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProvisioningInfo")
	ret0, _ := ret[0].(*params.ProvisioningInfoV10)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProvisioningInfo indicates an expected call of ProvisioningInfo.
func (mr *MockMachineProvisionerMockRecorder) ProvisioningInfo() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProvisioningInfo", reflect.TypeOf((*MockMachineProvisioner)(nil).ProvisioningInfo))
}

// Refresh mocks base method.
func (m *MockMachineProvisioner) Refresh() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Refresh")
	ret0, _ := ret[0].(error)
	return ret0
}

// Refresh indicates an expected call of Refresh.
func (mr *MockMachineProvisionerMockRecorder) Refresh() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Refresh", reflect.TypeOf((*MockMachineProvisioner)(nil).Refresh))
}

// Remove mocks base method.
func (m *MockMachineProvisioner) Remove() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove")
	ret0, _ := ret[0].(error)
	return ret0
}

// Remove indicates an expected call of Remove.
func (mr *MockMachineProvisionerMockRecorder) Remove() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockMachineProvisioner)(nil).Remove))
}

// SetCharmProfiles mocks base method.
func (m *MockMachineProvisioner) SetCharmProfiles(arg0 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetCharmProfiles", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetCharmProfiles indicates an expected call of SetCharmProfiles.
func (mr *MockMachineProvisionerMockRecorder) SetCharmProfiles(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetCharmProfiles", reflect.TypeOf((*MockMachineProvisioner)(nil).SetCharmProfiles), arg0)
}

// SetInstanceInfo mocks base method.
func (m *MockMachineProvisioner) SetInstanceInfo(arg0 instance.Id, arg1, arg2 string, arg3 *instance.HardwareCharacteristics, arg4 []params.NetworkConfig, arg5 []params.Volume, arg6 map[string]params.VolumeAttachmentInfo, arg7 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetInstanceInfo", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetInstanceInfo indicates an expected call of SetInstanceInfo.
func (mr *MockMachineProvisionerMockRecorder) SetInstanceInfo(arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetInstanceInfo", reflect.TypeOf((*MockMachineProvisioner)(nil).SetInstanceInfo), arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7)
}

// SetInstanceStatus mocks base method.
func (m *MockMachineProvisioner) SetInstanceStatus(arg0 status.Status, arg1 string, arg2 map[string]interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetInstanceStatus", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetInstanceStatus indicates an expected call of SetInstanceStatus.
func (mr *MockMachineProvisionerMockRecorder) SetInstanceStatus(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetInstanceStatus", reflect.TypeOf((*MockMachineProvisioner)(nil).SetInstanceStatus), arg0, arg1, arg2)
}

// SetModificationStatus mocks base method.
func (m *MockMachineProvisioner) SetModificationStatus(arg0 status.Status, arg1 string, arg2 map[string]interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetModificationStatus", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetModificationStatus indicates an expected call of SetModificationStatus.
func (mr *MockMachineProvisionerMockRecorder) SetModificationStatus(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetModificationStatus", reflect.TypeOf((*MockMachineProvisioner)(nil).SetModificationStatus), arg0, arg1, arg2)
}

// SetPassword mocks base method.
func (m *MockMachineProvisioner) SetPassword(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetPassword", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetPassword indicates an expected call of SetPassword.
func (mr *MockMachineProvisionerMockRecorder) SetPassword(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetPassword", reflect.TypeOf((*MockMachineProvisioner)(nil).SetPassword), arg0)
}

// SetStatus mocks base method.
func (m *MockMachineProvisioner) SetStatus(arg0 status.Status, arg1 string, arg2 map[string]interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetStatus", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetStatus indicates an expected call of SetStatus.
func (mr *MockMachineProvisionerMockRecorder) SetStatus(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetStatus", reflect.TypeOf((*MockMachineProvisioner)(nil).SetStatus), arg0, arg1, arg2)
}

// SetSupportedContainers mocks base method.
func (m *MockMachineProvisioner) SetSupportedContainers(arg0 ...instance.ContainerType) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "SetSupportedContainers", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetSupportedContainers indicates an expected call of SetSupportedContainers.
func (mr *MockMachineProvisionerMockRecorder) SetSupportedContainers(arg0 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetSupportedContainers", reflect.TypeOf((*MockMachineProvisioner)(nil).SetSupportedContainers), arg0...)
}

// Status mocks base method.
func (m *MockMachineProvisioner) Status() (status.Status, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Status")
	ret0, _ := ret[0].(status.Status)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Status indicates an expected call of Status.
func (mr *MockMachineProvisionerMockRecorder) Status() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Status", reflect.TypeOf((*MockMachineProvisioner)(nil).Status))
}

// String mocks base method.
func (m *MockMachineProvisioner) String() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "String")
	ret0, _ := ret[0].(string)
	return ret0
}

// String indicates an expected call of String.
func (mr *MockMachineProvisionerMockRecorder) String() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "String", reflect.TypeOf((*MockMachineProvisioner)(nil).String))
}

// SupportedContainers mocks base method.
func (m *MockMachineProvisioner) SupportedContainers() ([]instance.ContainerType, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SupportedContainers")
	ret0, _ := ret[0].([]instance.ContainerType)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// SupportedContainers indicates an expected call of SupportedContainers.
func (mr *MockMachineProvisionerMockRecorder) SupportedContainers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SupportedContainers", reflect.TypeOf((*MockMachineProvisioner)(nil).SupportedContainers))
}

// SupportsNoContainers mocks base method.
func (m *MockMachineProvisioner) SupportsNoContainers() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SupportsNoContainers")
	ret0, _ := ret[0].(error)
	return ret0
}

// SupportsNoContainers indicates an expected call of SupportsNoContainers.
func (mr *MockMachineProvisionerMockRecorder) SupportsNoContainers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SupportsNoContainers", reflect.TypeOf((*MockMachineProvisioner)(nil).SupportsNoContainers))
}

// Tag mocks base method.
func (m *MockMachineProvisioner) Tag() names.Tag {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Tag")
	ret0, _ := ret[0].(names.Tag)
	return ret0
}

// Tag indicates an expected call of Tag.
func (mr *MockMachineProvisionerMockRecorder) Tag() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Tag", reflect.TypeOf((*MockMachineProvisioner)(nil).Tag))
}

// WatchAllContainers mocks base method.
func (m *MockMachineProvisioner) WatchAllContainers() (watcher.StringsWatcher, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WatchAllContainers")
	ret0, _ := ret[0].(watcher.StringsWatcher)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WatchAllContainers indicates an expected call of WatchAllContainers.
func (mr *MockMachineProvisionerMockRecorder) WatchAllContainers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WatchAllContainers", reflect.TypeOf((*MockMachineProvisioner)(nil).WatchAllContainers))
}

// WatchContainers mocks base method.
func (m *MockMachineProvisioner) WatchContainers(arg0 instance.ContainerType) (watcher.StringsWatcher, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WatchContainers", arg0)
	ret0, _ := ret[0].(watcher.StringsWatcher)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WatchContainers indicates an expected call of WatchContainers.
func (mr *MockMachineProvisionerMockRecorder) WatchContainers(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WatchContainers", reflect.TypeOf((*MockMachineProvisioner)(nil).WatchContainers), arg0)
}

// Code generated by MockGen. DO NOT EDIT.
// Source: service.go

// Package mock_ports is a generated GoMock package.
package mock_ports

import (
	context "context"
	os "os"
	reflect "reflect"
	sync "sync"

	models "github.com/ScienceSoft-Inc/integrity-sum/internal/core/models"
	api "github.com/ScienceSoft-Inc/integrity-sum/pkg/api"
	gomock "github.com/golang/mock/gomock"
)

// MockIAppService is a mock of IAppService interface.
type MockIAppService struct {
	ctrl     *gomock.Controller
	recorder *MockIAppServiceMockRecorder
}

// MockIAppServiceMockRecorder is the mock recorder for MockIAppService.
type MockIAppServiceMockRecorder struct {
	mock *MockIAppService
}

// NewMockIAppService creates a new mock instance.
func NewMockIAppService(ctrl *gomock.Controller) *MockIAppService {
	mock := &MockIAppService{ctrl: ctrl}
	mock.recorder = &MockIAppServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIAppService) EXPECT() *MockIAppServiceMockRecorder {
	return m.recorder
}

// Check mocks base method.
func (m *MockIAppService) Check(ctx context.Context, dirPath string, sig chan os.Signal, deploymentData *models.DeploymentData, kuberData *models.KubeData) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Check", ctx, dirPath, sig, deploymentData, kuberData)
	ret0, _ := ret[0].(error)
	return ret0
}

// Check indicates an expected call of Check.
func (mr *MockIAppServiceMockRecorder) Check(ctx, dirPath, sig, deploymentData, kuberData interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Check", reflect.TypeOf((*MockIAppService)(nil).Check), ctx, dirPath, sig, deploymentData, kuberData)
}

// GetPID mocks base method.
func (m *MockIAppService) GetPID(procName string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPID", procName)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPID indicates an expected call of GetPID.
func (mr *MockIAppServiceMockRecorder) GetPID(procName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPID", reflect.TypeOf((*MockIAppService)(nil).GetPID), procName)
}

// IsExistDeploymentNameInDB mocks base method.
func (m *MockIAppService) IsExistDeploymentNameInDB(deploymentName string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsExistDeploymentNameInDB", deploymentName)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsExistDeploymentNameInDB indicates an expected call of IsExistDeploymentNameInDB.
func (mr *MockIAppServiceMockRecorder) IsExistDeploymentNameInDB(deploymentName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsExistDeploymentNameInDB", reflect.TypeOf((*MockIAppService)(nil).IsExistDeploymentNameInDB), deploymentName)
}

// LaunchHasher mocks base method.
func (m *MockIAppService) LaunchHasher(ctx context.Context, dirPath string, sig chan os.Signal) []*api.HashData {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LaunchHasher", ctx, dirPath, sig)
	ret0, _ := ret[0].([]*api.HashData)
	return ret0
}

// LaunchHasher indicates an expected call of LaunchHasher.
func (mr *MockIAppServiceMockRecorder) LaunchHasher(ctx, dirPath, sig interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LaunchHasher", reflect.TypeOf((*MockIAppService)(nil).LaunchHasher), ctx, dirPath, sig)
}

// Start mocks base method.
func (m *MockIAppService) Start(ctx context.Context, dirPath string, sig chan os.Signal, deploymentData *models.DeploymentData) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start", ctx, dirPath, sig, deploymentData)
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockIAppServiceMockRecorder) Start(ctx, dirPath, sig, deploymentData interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockIAppService)(nil).Start), ctx, dirPath, sig, deploymentData)
}

// MockIHashService is a mock of IHashService interface.
type MockIHashService struct {
	ctrl     *gomock.Controller
	recorder *MockIHashServiceMockRecorder
}

// MockIHashServiceMockRecorder is the mock recorder for MockIHashService.
type MockIHashServiceMockRecorder struct {
	mock *MockIHashService
}

// NewMockIHashService creates a new mock instance.
func NewMockIHashService(ctrl *gomock.Controller) *MockIHashService {
	mock := &MockIHashService{ctrl: ctrl}
	mock.recorder = &MockIHashServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIHashService) EXPECT() *MockIHashServiceMockRecorder {
	return m.recorder
}

// CreateHash mocks base method.
func (m *MockIHashService) CreateHash(path string) (*api.HashData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateHash", path)
	ret0, _ := ret[0].(*api.HashData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateHash indicates an expected call of CreateHash.
func (mr *MockIHashServiceMockRecorder) CreateHash(path interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateHash", reflect.TypeOf((*MockIHashService)(nil).CreateHash), path)
}

// DeleteFromTable mocks base method.
func (m *MockIHashService) DeleteFromTable(nameDeployment string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteFromTable", nameDeployment)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteFromTable indicates an expected call of DeleteFromTable.
func (mr *MockIHashServiceMockRecorder) DeleteFromTable(nameDeployment interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteFromTable", reflect.TypeOf((*MockIHashService)(nil).DeleteFromTable), nameDeployment)
}

// GetHashData mocks base method.
func (m *MockIHashService) GetHashData(dirPath string, deploymentData *models.DeploymentData) ([]*models.HashDataFromDB, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHashData", dirPath, deploymentData)
	ret0, _ := ret[0].([]*models.HashDataFromDB)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHashData indicates an expected call of GetHashData.
func (mr *MockIHashServiceMockRecorder) GetHashData(dirPath, deploymentData interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHashData", reflect.TypeOf((*MockIHashService)(nil).GetHashData), dirPath, deploymentData)
}

// IsDataChanged mocks base method.
func (m *MockIHashService) IsDataChanged(currentHashData []*api.HashData, hashSumFromDB []*models.HashDataFromDB, deploymentData *models.DeploymentData) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsDataChanged", currentHashData, hashSumFromDB, deploymentData)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsDataChanged indicates an expected call of IsDataChanged.
func (mr *MockIHashServiceMockRecorder) IsDataChanged(currentHashData, hashSumFromDB, deploymentData interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsDataChanged", reflect.TypeOf((*MockIHashService)(nil).IsDataChanged), currentHashData, hashSumFromDB, deploymentData)
}

// SaveHashData mocks base method.
func (m *MockIHashService) SaveHashData(allHashData []*api.HashData, deploymentData *models.DeploymentData) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveHashData", allHashData, deploymentData)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveHashData indicates an expected call of SaveHashData.
func (mr *MockIHashServiceMockRecorder) SaveHashData(allHashData, deploymentData interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveHashData", reflect.TypeOf((*MockIHashService)(nil).SaveHashData), allHashData, deploymentData)
}

// Worker mocks base method.
func (m *MockIHashService) Worker(wg *sync.WaitGroup, jobs <-chan string, results chan<- *api.HashData) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Worker", wg, jobs, results)
}

// Worker indicates an expected call of Worker.
func (mr *MockIHashServiceMockRecorder) Worker(wg, jobs, results interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Worker", reflect.TypeOf((*MockIHashService)(nil).Worker), wg, jobs, results)
}

// WorkerPool mocks base method.
func (m *MockIHashService) WorkerPool(jobs chan string, results chan *api.HashData) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "WorkerPool", jobs, results)
}

// WorkerPool indicates an expected call of WorkerPool.
func (mr *MockIHashServiceMockRecorder) WorkerPool(jobs, results interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WorkerPool", reflect.TypeOf((*MockIHashService)(nil).WorkerPool), jobs, results)
}

// MockIKuberService is a mock of IKuberService interface.
type MockIKuberService struct {
	ctrl     *gomock.Controller
	recorder *MockIKuberServiceMockRecorder
}

// MockIKuberServiceMockRecorder is the mock recorder for MockIKuberService.
type MockIKuberServiceMockRecorder struct {
	mock *MockIKuberService
}

// NewMockIKuberService creates a new mock instance.
func NewMockIKuberService(ctrl *gomock.Controller) *MockIKuberService {
	mock := &MockIKuberService{ctrl: ctrl}
	mock.recorder = &MockIKuberServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIKuberService) EXPECT() *MockIKuberServiceMockRecorder {
	return m.recorder
}

// ConnectionToK8sAPI mocks base method.
func (m *MockIKuberService) GetKubeData() (*models.KubeData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetKubeData")
	ret0, _ := ret[0].(*models.KubeData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ConnectionToK8sAPI indicates an expected call of ConnectionToK8sAPI.
func (mr *MockIKuberServiceMockRecorder) ConnectionToK8sAPI() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetKubeData", reflect.TypeOf((*MockIKuberService)(nil).GetKubeData))
}

// GetDataFromDeployment mocks base method.
func (m *MockIKuberService) GetDataFromDeployment(kuberData *models.KubeData) (*models.DeploymentData, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDataFromDeployment", kuberData)
	ret0, _ := ret[0].(*models.DeploymentData)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDataFromDeployment indicates an expected call of GetDataFromDeployment.
func (mr *MockIKuberServiceMockRecorder) GetDataFromDeployment(kuberData interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDataFromDeployment", reflect.TypeOf((*MockIKuberService)(nil).GetDataFromDeployment), kuberData)
}

// GetDataFromK8sAPI mocks base method.
func (m *MockIKuberService) GetDataFromK8sAPI() (*models.DataFromK8sAPI, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDataFromK8sAPI")
	ret0, _ := ret[0].(*models.DataFromK8sAPI)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDataFromK8sAPI indicates an expected call of GetDataFromK8sAPI.
func (mr *MockIKuberServiceMockRecorder) GetDataFromK8sAPI() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDataFromK8sAPI", reflect.TypeOf((*MockIKuberService)(nil).GetDataFromK8sAPI))
}

// RolloutDeployment mocks base method.
func (m *MockIKuberService) RolloutDeployment(kuberData *models.KubeData) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RolloutDeployment", kuberData)
	ret0, _ := ret[0].(error)
	return ret0
}

// RolloutDeployment indicates an expected call of RolloutDeployment.
func (mr *MockIKuberServiceMockRecorder) RolloutDeployment(kuberData interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RolloutDeployment", reflect.TypeOf((*MockIKuberService)(nil).RolloutDeployment), kuberData)
}

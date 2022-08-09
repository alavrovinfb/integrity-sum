package services

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/integrity-sum/internal/core/models"
	mock_ports "github.com/integrity-sum/internal/core/ports/mocks"
	"github.com/integrity-sum/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestCreateHash(t *testing.T) {
	testTable := []struct {
		name          string
		alg           string
		path          string
		mockBehavior  func(s *mock_ports.MockIHashService, path string)
		expected      *api.HashData
		expectedError bool
	}{
		{
			name: "exist file path",
			alg:  "SHA256",
			path: "test/test.txt",
			mockBehavior: func(s *mock_ports.MockIHashService, path string) {
				s.EXPECT().CreateHash(path).Return(&api.HashData{
					Hash:         "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
					FileName:     "test.txt",
					FullFilePath: "test/test.txt",
					Algorithm:    "SHA256",
				}, nil)

			},
			expected: &api.HashData{
				Hash:         "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				FileName:     "test.txt",
				FullFilePath: "test/test.txt",
				Algorithm:    "SHA256",
			},
		},
		{
			name: "exist dir path",
			alg:  "SHA256",
			path: "test/h",
			mockBehavior: func(s *mock_ports.MockIHashService, path string) {
				s.EXPECT().CreateHash(path).Return(&api.HashData{
					Hash:         "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
					FileName:     "test2.txt",
					FullFilePath: "test/h",
					Algorithm:    "SHA256",
				}, nil)

			},
			expected: &api.HashData{
				Hash:         "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				FileName:     "test2.txt",
				FullFilePath: "test/h",
				Algorithm:    "SHA256",
			},
		},
		{
			name: "another algorithm",
			alg:  "SHA1",
			path: "test/test.txt",
			mockBehavior: func(s *mock_ports.MockIHashService, path string) {
				s.EXPECT().CreateHash(path).Return(&api.HashData{
					Hash:         "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
					FileName:     "test.txt",
					FullFilePath: "test/test.txt",
					Algorithm:    "SHA1",
				}, nil)

			},
			expected: &api.HashData{
				Hash:         "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				FileName:     "test.txt",
				FullFilePath: "test/test.txt",
				Algorithm:    "SHA1",
			},
		},
		{
			name: "don`t exist file path",
			alg:  "SHA256",
			path: "/h/new.txt",
			mockBehavior: func(s *mock_ports.MockIHashService, path string) {
				s.EXPECT().CreateHash(path).Return(nil, errors.New("do not exist file path"))

			},
			expectedError: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			service := mock_ports.NewMockIHashService(c)
			testCase.mockBehavior(service, testCase.path)

			file, err := os.Open(testCase.path)
			if err != nil {
				require.Error(t, err)
			}
			defer file.Close()

			result, err := service.CreateHash(testCase.path)

			if testCase.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expected, result)
			}
		})
	}
}

func TestSaveHashData(t *testing.T) {
	testTable := []struct {
		name           string
		allHashData    []*api.HashData
		deploymentData *models.DeploymentData
		mockBehavior   func(s *mock_ports.MockIHashService, allHashData []*api.HashData, deploymentData *models.DeploymentData)
		expectedError  bool
		expected       error
	}{
		{
			name: "error is nil",
			allHashData: []*api.HashData{{
				Hash:         "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				FileName:     "test.txt",
				FullFilePath: "test/test.txt",
				Algorithm:    "SHA256",
			}},
			deploymentData: &models.DeploymentData{
				Image:                "nginx:latest",
				NamePod:              "app-nginx-hasher-integrity-6b64487565-l8ltd",
				Timestamp:            "",
				NameDeployment:       "app-nginx-hasher-integrity",
				LabelMainProcessName: "nginx",
				ReleaseName:          "app",
			},
			mockBehavior: func(s *mock_ports.MockIHashService, allHashData []*api.HashData, deploymentData *models.DeploymentData) {
				s.EXPECT().SaveHashData(allHashData, deploymentData).Return(nil)

			},
		},
		{
			name: "error isn`t nil",
			allHashData: []*api.HashData{{
				Hash:         "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				FileName:     "test.txt",
				FullFilePath: "test/test.txt",
				Algorithm:    "SHA256",
			}},
			deploymentData: &models.DeploymentData{
				Image:                "nginx:latest",
				NamePod:              "app-nginx-hasher-integrity-6b64487565-l8ltd",
				Timestamp:            "",
				NameDeployment:       "app-nginx-hasher-integrity",
				LabelMainProcessName: "nginx",
				ReleaseName:          "app",
			},
			mockBehavior: func(s *mock_ports.MockIHashService, allHashData []*api.HashData, deploymentData *models.DeploymentData) {
				s.EXPECT().SaveHashData(allHashData, deploymentData).Return(errors.New("error while saving data to database"))

			},
			expectedError: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			service := mock_ports.NewMockIHashService(c)
			testCase.mockBehavior(service, testCase.allHashData, testCase.deploymentData)

			err := service.SaveHashData(testCase.allHashData, testCase.deploymentData)

			if testCase.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expected, err)
			}
		})
	}
}

func TestGetHashData(t *testing.T) {
	testTable := []struct {
		name           string
		dirFiles       string
		deploymentData *models.DeploymentData
		mockBehavior   func(s *mock_ports.MockIHashService, dirFiles string, deploymentData *models.DeploymentData)
		hashData       []*models.HashDataFromDB
		expectedError  bool
		expected       []*models.HashDataFromDB
	}{
		{
			name:     "got hash data from database",
			dirFiles: "..test/",
			deploymentData: &models.DeploymentData{
				Image:                "nginx:latest",
				NamePod:              "app-nginx-hasher-integrity-6b64487565-l8ltd",
				Timestamp:            "",
				NameDeployment:       "app-nginx-hasher-integrity",
				LabelMainProcessName: "nginx",
				ReleaseName:          "app",
			},
			mockBehavior: func(s *mock_ports.MockIHashService, dirFiles string, deploymentData *models.DeploymentData) {
				s.EXPECT().GetHashData(dirFiles, deploymentData).Return([]*models.HashDataFromDB{}, nil)
			},
			expected: []*models.HashDataFromDB{},
		},
		{
			name:     "error isn`t nil",
			dirFiles: "..test/",
			deploymentData: &models.DeploymentData{
				Image:                "nginx:latest",
				NamePod:              "app-nginx-hasher-integrity-6b64487565-l8ltd",
				Timestamp:            "",
				NameDeployment:       "app-nginx-hasher-integrity",
				LabelMainProcessName: "nginx",
				ReleaseName:          "app",
			},
			mockBehavior: func(s *mock_ports.MockIHashService, dirFiles string, deploymentData *models.DeploymentData) {
				s.EXPECT().GetHashData(dirFiles, deploymentData).Return([]*models.HashDataFromDB{}, errors.New("hashData service didn't get hashData sum"))
			},
			expectedError: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			service := mock_ports.NewMockIHashService(c)
			testCase.mockBehavior(service, testCase.dirFiles, testCase.deploymentData)

			res, err := service.GetHashData(testCase.dirFiles, testCase.deploymentData)

			if testCase.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expected, res)
			}
		})
	}
}

func TestDeleteFromTable(t *testing.T) {
	testTable := []struct {
		name           string
		nameDeployment string
		mockBehavior   func(s *mock_ports.MockIHashService, nameDeployment string)
		expectedError  bool
		expected       error
	}{
		{
			name:           "error is nil",
			nameDeployment: "app-nginx-hasher-integrity",
			mockBehavior: func(s *mock_ports.MockIHashService, nameDeployment string) {
				s.EXPECT().DeleteFromTable(nameDeployment).Return(nil)
			},
		},
		{
			name:           "error isn`t nil",
			nameDeployment: "app-nginx-hasher-integrity",
			mockBehavior: func(s *mock_ports.MockIHashService, nameDeployment string) {
				s.EXPECT().DeleteFromTable(nameDeployment).Return(errors.New("err while deleting rows in database"))
			},
			expectedError: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			service := mock_ports.NewMockIHashService(c)
			testCase.mockBehavior(service, testCase.nameDeployment)

			err := service.DeleteFromTable(testCase.nameDeployment)

			if testCase.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expected, err)
			}
		})
	}
}

func TestIsDataChanged(t *testing.T) {
	testTable := []struct {
		name            string
		currentHashData []*api.HashData
		hashDataFromDB  []*models.HashDataFromDB
		deploymentData  *models.DeploymentData
		mockBehavior    func(s *mock_ports.MockIHashService, currentHashData []*api.HashData, hashDataFromDB []*models.HashDataFromDB, deploymentData *models.DeploymentData)
		expectedError   bool
		expected        bool
	}{
		{
			name: "the current data and the data in the database are the same",
			currentHashData: []*api.HashData{{
				Hash:         "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				FileName:     "test.txt",
				FullFilePath: "test/test.txt",
				Algorithm:    "SHA256",
			}},
			hashDataFromDB: []*models.HashDataFromDB{{
				ID:             1,
				Hash:           "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				FileName:       "test.txt",
				FullFilePath:   "test/test.txt",
				Algorithm:      "SHA256",
				ImageContainer: "nginx:latest",
				NamePod:        "app-nginx-hasher-integrity-6b64487565-l8ltd",
				NameDeployment: "app-nginx-hasher-integrity",
			}},
			deploymentData: &models.DeploymentData{
				Image:                "nginx:latest",
				NamePod:              "app-nginx-hasher-integrity-6b64487565-l8ltd",
				Timestamp:            "",
				NameDeployment:       "app-nginx-hasher-integrity",
				LabelMainProcessName: "nginx",
				ReleaseName:          "app",
			},
			mockBehavior: func(s *mock_ports.MockIHashService, currentHashData []*api.HashData, hashDataFromDB []*models.HashDataFromDB, deploymentData *models.DeploymentData) {
				s.EXPECT().IsDataChanged(currentHashData, hashDataFromDB, deploymentData).Return(false)
			},
		},
		{
			name:            "current data and data in database are different",
			currentHashData: []*api.HashData{{}},
			hashDataFromDB:  []*models.HashDataFromDB{{}},
			deploymentData:  &models.DeploymentData{},
			mockBehavior: func(s *mock_ports.MockIHashService, currentHashData []*api.HashData, hashDataFromDB []*models.HashDataFromDB, deploymentData *models.DeploymentData) {
				s.EXPECT().IsDataChanged(currentHashData, hashDataFromDB, deploymentData).Return(true)
			},
			expected: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			service := mock_ports.NewMockIHashService(c)
			testCase.mockBehavior(service, testCase.currentHashData, testCase.hashDataFromDB, testCase.deploymentData)

			res := service.IsDataChanged(testCase.currentHashData, testCase.hashDataFromDB, testCase.deploymentData)

			assert.Equal(t, testCase.expected, res)

		})
	}
}

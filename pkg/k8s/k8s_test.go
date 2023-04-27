package k8s_test

import (
	"github.com/ScienceSoft-Inc/integrity-sum/pkg/k8s"
	mockk8s "github.com/ScienceSoft-Inc/integrity-sum/pkg/k8s/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKubeClient(t *testing.T) {
	// Create a new mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create a mock implementation of the IKuberService interface
	mockService := mockk8s.NewMockIKuberService(ctrl)

	// Define test data
	deploymentData := &k8s.DeploymentData{
		Image:          "test-image",
		NamePod:        "test-pod",
		Timestamp:      "test-timestamp",
		NameDeployment: "test-deployment",
		ReleaseName:    "test-release",
		NameSpace:      "test-namespace",
	}

	// Test the Connect method
	mockService.EXPECT().Connect().Return(nil)
	err := mockService.Connect()
	assert.NoError(t, err)

	// Test the GetDataFromDeployment method
	mockService.EXPECT().GetDataFromDeployment().Return(deploymentData, nil)
	data, err := mockService.GetDataFromDeployment()
	assert.NoError(t, err)
	assert.Equal(t, deploymentData, data)

	// Test the RestartPod method
	mockService.EXPECT().RestartPod().Return(nil)
	err = mockService.RestartPod()
	assert.NoError(t, err)
}

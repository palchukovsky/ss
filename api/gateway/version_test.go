// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package apigateway_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/palchukovsky/ss"
	apigateway "github.com/palchukovsky/ss/api/gateway"
	mock_ss "github.com/palchukovsky/ss/mock"
	"github.com/stretchr/testify/assert"
)

func Test_APIGateway_CheckClientVersionActuality(test *testing.T) {
	mock := gomock.NewController(test)
	defer mock.Finish()
	assert := assert.New(test)

	config := ss.ServiceConfig{
		App: struct {
			MinVersion [4]uint `json:"minVer"`
			Domain     string  `json:"domain"`
			Android    struct {
				Package string `json:"package"`
			} `json:"android"`
			IOS struct {
				Bundle string `json:"bundle"`
			} `json:"ios"`
		}{
			MinVersion: [4]uint{50, 100, 150, 200},
		},
	}

	service := mock_ss.NewMockService(mock)
	service.EXPECT().Config().MinTimes(1).Return(config)
	ss.Set(service)

	{
		isActual, err := apigateway.CheckClientVersionActuality("123.123.123.qwe")
		assert.False(isActual)
		assert.EqualError(err, `failed to parse client version "123.123.123.qwe"`)
	}
	{
		isActual, err := apigateway.CheckClientVersionActuality("50.100.150.200")
		assert.NoError(err)
		assert.True(isActual)
	}
	{
		isActual, err := apigateway.CheckClientVersionActuality("51.100.150.200")
		assert.NoError(err)
		assert.True(isActual)
	}
	{
		isActual, err := apigateway.CheckClientVersionActuality("50.101.150.200")
		assert.NoError(err)
		assert.True(isActual)
	}
	{
		isActual, err := apigateway.CheckClientVersionActuality("50.100.151.200")
		assert.NoError(err)
		assert.True(isActual)
	}
	{
		isActual, err := apigateway.CheckClientVersionActuality("50.100.150.201")
		assert.NoError(err)
		assert.True(isActual)
	}
	{
		isActual, err := apigateway.CheckClientVersionActuality("49.100.150.200")
		assert.NoError(err)
		assert.False(isActual)
	}
	{
		isActual, err := apigateway.CheckClientVersionActuality("50.99.150.200")
		assert.NoError(err)
		assert.False(isActual)
	}
	{
		isActual, err := apigateway.CheckClientVersionActuality("50.100.149.200")
		assert.NoError(err)
		assert.False(isActual)
	}
	{
		isActual, err := apigateway.CheckClientVersionActuality("50.100.150.199")
		assert.NoError(err)
		assert.False(isActual)
	}
	{
		isActual, err := apigateway.CheckClientVersionActuality("51.100.149.200")
		assert.NoError(err)
		assert.True(isActual)
	}
	{
		isActual, err := apigateway.CheckClientVersionActuality("50.101.149.200")
		assert.NoError(err)
		assert.True(isActual)
	}
	{
		isActual, err := apigateway.CheckClientVersionActuality("51.99.149.200")
		assert.NoError(err)
		assert.True(isActual)
	}
}

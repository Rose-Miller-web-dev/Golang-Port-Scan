package main

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func Test_writeIpOnAChannel(t *testing.T) {

	var testCase struct {
		name string
		ip   addreObj
	}

	testCase.name = "write the 82.99.202.35 on the channel"
	testCase.ip.IP = "82.99.202.35"

	t.Run(testCase.name, func(t *testing.T) {
		writeIpOnAChannel(testCase.ip)
	})

}

func Test_getSingleChannelInstanceReturnsTheInstance(t *testing.T) {

	var testCase struct {
		name string
	}

	testCase.name = "CheckIfThe Instance Is created"

	t.Run(testCase.name, func(t *testing.T) {
		GetSingleChannelInstance()
	})

}

func Test_postIpPostsSampleIp(t *testing.T) {
	var testCase struct {
		name string
	}

	t.Run(testCase.name, func(t *testing.T) {
		router := gin.Default()
		router.POST("/addresses", postIp)
		router.Run("localhost:8080")
	})
}

func Test_runNmapForSampleIP(t *testing.T) {
	var testCase struct {
		name string
		ip   string
	}

	testCase.ip = "82.99.202.35"
	testCase.name = "Check If The IP is run on the nmap service"

	t.Run(testCase.name, func(t *testing.T) {
		runNmapForIp(testCase.ip)
	})
}

package mipv4_test

import (
	"testing"

	"github.com/graingo/maltose/net/mipv4"
	"github.com/stretchr/testify/assert"
)

func TestIsIntranet(t *testing.T) {
	testCases := []struct {
		name     string
		ip       string
		expected bool
	}{
		// A Class Public IPs
		{name: "a class public ip", ip: "1.1.1.1", expected: false},
		{name: "a class public ip edge", ip: "9.255.255.255", expected: false},
		{name: "a class public ip edge 2", ip: "11.0.0.0", expected: false},

		// A Class Private IPs
		{name: "a class private ip start", ip: "10.0.0.0", expected: true},
		{name: "a class private ip middle", ip: "10.1.2.3", expected: true},
		{name: "a class private ip end", ip: "10.255.255.255", expected: true},

		// B Class Public IPs
		{name: "b class public ip", ip: "172.15.255.255", expected: false},
		{name: "b class public ip 2", ip: "172.32.0.0", expected: false},
		{name: "b class public ip edge", ip: "171.255.255.255", expected: false},

		// B Class Private IPs
		{name: "b class private ip start", ip: "172.16.0.0", expected: true},
		{name: "b class private ip middle", ip: "172.20.10.1", expected: true},
		{name: "b class private ip end", ip: "172.31.255.255", expected: true},

		// C Class Public IPs
		{name: "c class public ip", ip: "192.167.255.255", expected: false},
		{name: "c class public ip 2", ip: "192.169.0.0", expected: false},

		// C Class Private IPs
		{name: "c class private ip start", ip: "192.168.0.0", expected: true},
		{name: "c class private ip middle", ip: "192.168.1.100", expected: true},
		{name: "c class private ip end", ip: "192.168.255.255", expected: true},

		// Other IPs
		{name: "google dns", ip: "8.8.8.8", expected: false},
		{name: "broadcast address", ip: "255.255.255.255", expected: false},
		{name: "loopback", ip: "127.0.0.1", expected: false},

		// Invalid Inputs
		{name: "invalid ip format", ip: "192.168.1", expected: false},
		{name: "invalid ip with letters", ip: "192.168.1.abc", expected: false},
		{name: "empty string", ip: "", expected: false},
		{name: "just a word", ip: "localhost", expected: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, mipv4.IsIntranet(tc.ip))
		})
	}
}

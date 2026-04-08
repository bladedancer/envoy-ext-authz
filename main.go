package main

import (
	cmd "github.com/bladedancer/envoy-ext-authz/cmd/ext-authz"
)

func main() {
	cmd.RootCmd.Execute()
}

package main

import (
	"github.com/OleG2e/collector/internal/container"
	"github.com/OleG2e/collector/internal/storage"
)

func main() {
	container.InitAgentContainer()
	storage.RunMonitor()
}

package main

import (
	"fmt"
	"log"

	// "net"
	"os"
	"os/exec"
	"os/signal"
	vb "sideDesert/shiba/internal/vbrowser"
	"syscall"
	// "github.com/pion/rtp"
	// "github.com/pion/webrtc/v4"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--pipeline" {
		runPipeline()
		return
	}

	fmt.Println("Starting Browser Session...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Spawn pipeline subprocess
	cmd := exec.Command(os.Args[0], "--pipeline")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatal("❌ Failed to start pipeline subprocess:", err)
	}

	// Only kill subprocess on Ctrl+C
	<-sigChan
	fmt.Println("Shutting Down...")
	_ = cmd.Process.Kill()

}

func runPipeline() {
	manager := vb.NewManager(99)

	err := manager.SetupPipeline()
	if err != nil {
		log.Fatal("❌ Pipeline setup failed:", err)
	}

	log.Println("✅ Pipeline Running...")

	// ⏸️ Keep this process alive while the pipeline runs
	select {}
}

// func main() {
// 	fmt.Println("Starting Browser Session...")

// 	sigChan := make(chan os.Signal, 1)
// 	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

// 	manager := vb.NewManager(99)

// 	// go manager.StartVirtualBrowser(ctx)
// 	manager.SetupPipeline()

// 	<-sigChan
// 	fmt.Println("Shutting Down...")
// }

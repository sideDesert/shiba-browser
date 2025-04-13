package vbrowser

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/go-gst/go-gst/gst"
)

func (d *VbrowserManager) StartVirtualBrowser(ctx context.Context) {
	portStr := fmt.Sprintf(":%d", d.Display.Port)
	displayStr := fmt.Sprintf("%dx%dx%d", d.Display.Width, d.Display.Height, 24)

	time.Sleep(1 * time.Second)

	checkXvfb := exec.Command("pgrep", "Xvfb")
	if err := checkXvfb.Start(); err == nil {
		log.Println("Warning: Xvfb is still running. Trying again...")
		exec.Command("pkill", "-9", "Xvfb").Start()
		time.Sleep(1 * time.Second)
	}

	xvfbCmd := exec.Command("Xvfb", portStr, "-screen", "0", displayStr)
	xvfbLog, _ := os.Create("xvfb.log")
	xvfbCmd.Stdout = xvfbLog
	xvfbCmd.Stderr = xvfbLog
	xvfbCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err := xvfbCmd.Start()
	if err != nil {
		log.Println("Failed to start Xvfb:", err)
	}
	log.Println("Xvfb started on DISPLAY=" + portStr)

	// Give Xvfb some time to start
	time.Sleep(2 * time.Second)
	d.Ready <- StepXvfbReady

	// Set DISPLAY environment variable
	os.Setenv("DISPLAY", portStr)

	// Start Chrome inside Xvfb
	chromeCmd := exec.Command("google-chrome",
		fmt.Sprintf("--window-size=%d,%d", d.Display.Width, d.Display.Height),
		"--no-sandbox",
		"--disable-gpu",
		"--new-window",
		"--user-data-dir=./tmp/chrome-xvfb", // Separate profile
		"--remote-debugging-port=9222",
		d.defaultUrl,
	)

	chromeLog, _ := os.Create("chrome.log")
	chromeCmd.Stdout = chromeLog
	chromeCmd.Stderr = chromeLog
	chromeCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // 🛡️ Same for Chrome
	err = chromeCmd.Start()

	if err != nil {
		log.Println("❌ Failed to start Chrome:", err)
	}
	log.Println("✅ Chrome started inside Xvfb" + portStr)
	d.Ready <- StepBrowserReady
	d.pid = chromeCmd.Process.Pid

	// Ensure Chrome is cleaned up on exit
	defer func() {
		log.Println("⛔ Stopping Chrome...")
		_ = syscall.Kill(-d.pid, syscall.SIGTERM) // graceful
		time.Sleep(1 * time.Second)
		_ = syscall.Kill(-d.pid, syscall.SIGKILL) // force kill
		_ = chromeCmd.Wait()                      // reap process

		log.Println("🧨 Killing Xvfb...")
		_ = syscall.Kill(-xvfbCmd.Process.Pid, syscall.SIGINT)
		time.Sleep(1 * time.Second)
		_ = xvfbCmd.Wait() // reap process

		exec.Command("pkill", "-9", "Xvfb").Start()
		time.Sleep(1 * time.Second)

		err := os.RemoveAll("/tmp/.X99-lock")
		if err != nil {
			log.Println("Could not remove lock file:", err)
		}
		log.Println("✅ Xvfb killed")
		log.Println("✅ Cleanup done")
	}()

	// Handle OS signals for cleanup
	<-ctx.Done()
	log.Println("🧹 Context cancelled, cleaning up Chrome & Xvfb")
}

func (d *VbrowserManager) StartVideoStream(ctx context.Context) {
	timeout := time.NewTimer(20 * time.Second)
	defer timeout.Stop()
	for {
		select {
		case step := <-d.Ready:
			switch step {
			case StepXvfbReady:
				log.Println("🔥🖥️ XVFB is running!")
				d.ConnReady <- StepXvfbReady
			case StepBrowserReady:
				log.Println("🔥🌍 Chrome is running!")
				err := d.SetupPipeline()
				if err != nil {
					log.Println("Error in StartVideoStream[PipelineSetup]:", err)
					return
				}
				d.ConnReady <- StepBrowserReady

			case StepPipelineReady:
				err := d.Pipeline.SetState(gst.StatePlaying)
				if err != nil {
					log.Println("Error in SetupPipeline[Setting Pipeline State]: ", err)
					return
				}

				log.Println("🔥🎥 Pipeline is ready!")
				defer func() {
					log.Println("Stopping Browser Stream...")
				}()
				d.ConnReady <- StepPipelineReady

				<-ctx.Done()
				if d.Pipeline != nil {
					err := d.Pipeline.SetState(gst.StateNull)
					if err != nil {
						log.Println("Error in StartVideoStream[Pipeline.SetState(gst.StateNull)]: ", err)
						return
					}
					d.Pipeline = nil
				}
			}

		case <-timeout.C:
			log.Println("⏰ Timeout waiting for steps to complete")
			return
		}
	}
}

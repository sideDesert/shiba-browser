package vbrowser

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/go-gst/go-gst/gst"
	"github.com/gorilla/websocket"
)

type Step int

const (
	StepXvfbReady Step = iota
	StepBrowserReady
	StepPipelineReady
	StepEstablishRTCConn
	StepCleanup
)

type VbrowserManager struct {
	Display      *Display
	Pipeline     *gst.Pipeline
	Ws           *websocket.Conn
	UdpVideoPort int
	UdpAudioPort int
	Ready        chan Step
	ConnReady    chan Step
	Cursor       Cursor
	Keyboard     Keyboard

	pid        int
	defaultUrl string
}

func NewManager(port int) *VbrowserManager {
	return &VbrowserManager{
		Display:      NewDisplay(port, 1080, 1920, 60),
		Ready:        make(chan Step, 5),
		ConnReady:    make(chan Step, 5),
		defaultUrl:   "https://www.youtube.com/watch?v=OPK14FrnjO0&ab_channel=JackHarlow",
		Cursor:       NewCursor(0, 0),
		Keyboard:     Keyboard{},
		UdpVideoPort: 5005,
		UdpAudioPort: 5006,
	}
}

func (m *VbrowserManager) SetWs(ws *websocket.Conn) {
	m.Ws = ws
}

func (m *VbrowserManager) SetupPipeline() error {
	// Make this work
	gst.Init(nil)
	height := m.Display.Height
	width := m.Display.Width
	os.Setenv("DISPLAY", ":"+strconv.Itoa(m.Display.Port))

	pipelineStr := fmt.Sprintf(`ximagesrc use-damage=0 display-name=":%d"
    ! queue
    ! videoconvert
    ! video/x-raw,format=I420
    ! queue
    ! videoscale
    ! video/x-raw,width=%d,height=%d,framerate=60/1
    ! queue
    ! x264enc bitrate=4000 tune=zerolatency speed-preset=veryfast key-int-max=30
    ! queue
    ! video/x-h264,stream-format=byte-stream
    ! queue
    ! appsink name=videoSink emit-signals=true sync=false

    pulsesrc
    ! queue
    ! audioconvert
    ! audioresample
    ! queue
    ! opusenc
    ! queue
    ! appsink name=audioSink emit-signals=true sync=false`, m.Display.Port, width, height)

	pipeline, err := gst.NewPipelineFromString(pipelineStr)

	if err != nil {
		fmt.Println("Error in SetupPipeline[Pipeline Creation]: ", err)
		return err
	}

	m.Pipeline = pipeline
	bus := pipeline.GetBus()

	go func() {
		for {
			msg := bus.Pop()
			if msg == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			switch msg.Type() {
			case gst.MessageError:
				err := msg.ParseError()
				fmt.Println("Pipeline Error:", err)
				_ = pipeline.SetState(gst.StateNull)

			case gst.MessageWarning:
				warn := msg.ParseWarning()
				fmt.Println("Pipeline Warning:", warn)
			}
		}
	}()

	m.Ready <- StepPipelineReady
	log.Println("✅Pipeline Setup Succesfully!")
	return nil
}

func (m *Cursor) Move(x float32, y float32) error {
	m.X = x
	m.Y = y
	cmdstr := fmt.Sprintf("DISPLAY :99 xdotool mousemove %f %f", x, y)
	cmd := exec.Command(cmdstr)
	return cmd.Run()
}

func (m *Cursor) Click() error {
	// Execute a click
	cmdstr := "DISPLAY :99 xdotool click 1"
	cmd := exec.Command(cmdstr)
	return cmd.Run()
}

func (k *Keyboard) SendKeys(keys string) error {
	// Do some of that xdotools stuff
	parsed := parseKey(keys)
	cmdstr := fmt.Sprintf("DISPLAY :99 xdotool key %s", parsed)
	cmd := exec.Command(cmdstr)

	return cmd.Run()
}

func parseKey(key string) string {
	return key
}

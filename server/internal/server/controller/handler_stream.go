package controller

import (
	"context"
	"fmt"
	"log"
	"sideDesert/shiba/internal/server/dto"
	"sideDesert/shiba/internal/server/lib"
	"sideDesert/shiba/internal/vbrowser"
	"time"

	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
)

func (c *Controller) handleStream(userId string, chatroomId string) error {

	// if isRemote := c.s.CheckUserIsRemoteForChatroom(userId, chatroomId); !isRemote {
	// 	return fmt.Errorf("User is not remote for chatroom")
	// }

	chatroomUsersIds, err := c.s.Store.GetUsersByChatroomId(c.s.Ctx, chatroomId)
	if err != nil {
		log.Println("Error in handleStream[GetUserByChatroomId]:", err)
		return err
	}

	c.mu.Lock()
	chatroomCtx, ok := c.chatroomCtx[chatroomId]
	c.mu.Unlock()

	if ok {
		if chatroomCtx.Streaming {
			log.Println("Error: Streaming already taking place for chatroom - ", chatroomId)
			return fmt.Errorf("Error: Streaming already taking place for this chatroom")
		}
	}

	if !ok {
		port, err := c.NewVideoPort()
		if err != nil {
			return err
		}
		nctx := NewChatroomCtx(context.Background(), port)
		c.mu.Lock()
		c.chatroomCtx[chatroomId] = nctx
		c.mu.Unlock()
	}
	c.mu.Lock()
	ctx := c.chatroomCtx[chatroomId].ctx
	c.mu.Unlock()

	bm := chatroomCtx.BrowserManager

	go bm.StartVirtualBrowser(ctx)
	go bm.StartVideoStream(ctx)

	for {
		step := <-bm.ConnReady
		if step == vbrowser.StepPipelineReady {

			activePeers, err := c.getActivePeers(chatroomUsersIds, chatroomId)
			if err != nil {
				log.Println("Error in getting active peers:", err)
				return err
			}

			// VIDEO Stream Handler
			elem, err := bm.Pipeline.GetElementByName("videoSink")
			if err != nil {
				log.Println("Error in converting videoSink element from pipeline:")
				return fmt.Errorf("error in getting videoSink element from pipeline")
			}

			videoSink := app.SinkFromElement(elem)
			if videoSink == nil {
				log.Println("")
			}
			videoSink.SetCallbacks(&app.SinkCallbacks{
				NewSampleFunc: func(sink *app.Sink) gst.FlowReturn {
					sample := sink.PullSample()

					if sample == nil {
						return gst.FlowEOS
					}
					buffer := sample.GetBuffer()
					if buffer == nil {
						return gst.FlowEOS
					}

					data := buffer.Bytes()
					if len(data) == 0 {
						return gst.FlowEOS
					}

					for _, pc := range activePeers {
						pc.videoStream.WriteSample(media.Sample{
							Data:     data,
							Duration: time.Second / 60,
						})
					}
					return gst.FlowOK
				},
			})

			// AUDIO
			elem, err = bm.Pipeline.GetElementByName("audioSink")
			if err != nil {
				log.Println("Error in getting audioSink element from pipeline:", err)
				return err
			}
			audioSink := app.SinkFromElement(elem)
			if audioSink == nil {
				log.Println("Error in converting audioSink element from pipeline:")
				return fmt.Errorf("error in getting audioSink element from pipeline")
			}
			audioSink.SetCallbacks(&app.SinkCallbacks{
				NewSampleFunc: func(sink *app.Sink) gst.FlowReturn {
					sample := sink.PullSample()

					if sample == nil {
						return gst.FlowEOS
					}
					buffer := sample.GetBuffer()
					if buffer == nil {
						return gst.FlowEOS
					}

					data := buffer.Bytes()
					if len(data) == 0 {
						return gst.FlowEOS
					}

					for _, pc := range activePeers {
						pc.audioStream.WriteSample(media.Sample{
							Data:     data,
							Duration: time.Second / 60,
						})
					}
					return gst.FlowOK
				},
			})
		}
	}
}

type ActivePeer struct {
	userId      string
	chatroomId  string
	videoStream *webrtc.TrackLocalStaticSample
	audioStream *webrtc.TrackLocalStaticSample
}

type HandleVideoStreamConfig struct {
	userId          string
	chatroomId      string
	videoPort       int
	audioPort       int
	videoStream     *webrtc.TrackLocalStaticRTP
	audioStream     *webrtc.TrackLocalStaticRTP
	videoSink       *gst.Element
	audioSink       *gst.Element
	videoSSRC       uint32
	audioSSRC       uint32
	videoSeqCounter uint16
	audioSeqCounter uint16
}

func (c *Controller) getActivePeers(chatroomUsersIds []string, chatroomId string) ([]ActivePeer, error) {
	activePeers := make([]ActivePeer, 0)
	for ws, config := range c.conns {
		if !lib.Contains(chatroomUsersIds, config.UserId) {
			continue
		}
		log.Println("Creating stream for user - ", config.UserId)
		pc := config.StreamConfig.PeerConnection
		log.Println("Connection state =", pc.ConnectionState())
		// Now Add track
		if pc.ConnectionState().String() == "closed" {
			_pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
			if err != nil {
				log.Println("Error Creating new Peer Connection:", err)
				return activePeers, err
			}
			config.StreamConfig.PeerConnection = _pc
			pc = _pc
		}
		pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
			// Send the ICE Candidate to the client
			if candidate == nil {
				return
			}
			msg := dto.Message[webrtc.ICECandidateInit]{
				Sender:  "server",
				Subject: "stream.ice." + chatroomId + "." + config.UserId,
				Payload: candidate.ToJSON(),
			}
			err := ws.WriteJSON(msg)
			if err != nil {
				log.Println("Could not send ICE Candidate")
				config.StreamConfig.IceCandidates = append(config.StreamConfig.IceCandidates, candidate)
			}
		})
		pc.OnICEGatheringStateChange(func(state webrtc.ICEGatheringState) {
			if state == webrtc.ICEGatheringStateComplete {
				for _, candidate := range config.StreamConfig.IceCandidates {
					msg := dto.Message[webrtc.ICECandidateInit]{
						Sender:  "server",
						Subject: "stream.ice." + chatroomId + "." + config.UserId,
						Payload: candidate.ToJSON(),
					}
					ws.WriteJSON(msg)
				}
			}
		})

		sdp, err := pc.CreateOffer(&webrtc.OfferOptions{})
		if err != nil {
			log.Println("Error creating SDP offer:", err)
			return activePeers, err
		}
		// log.Println("✅ SDP offer created")

		err = pc.SetLocalDescription(sdp)
		if err != nil {
			log.Println("Error setting local description:", err)
			return activePeers, err
		}

		ws.WriteJSON(dto.Message[string]{
			Sender:  "server",
			Subject: "stream.offer." + chatroomId + "." + config.UserId,
			Payload: sdp.SDP,
		})

		activePeers = append(activePeers, ActivePeer{
			userId:      config.UserId,
			chatroomId:  chatroomId,
			audioStream: config.StreamConfig.AudioTrack,
			videoStream: config.StreamConfig.VideoTrack,
		})
		log.Println("✅ SDP offer sent to user ", config.UserId)
	}

	return activePeers, nil
}

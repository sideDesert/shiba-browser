#!/bin/bash
gst-launch-1.0 ximagesrc display-name=:99 use-damage=0 \
	! video/x-raw,framerate=60/1 \
	! videoconvert \
	! x264enc tune=zerolatency \
	! rtph264pay \
	! webrtcbin bundle-policy=max-bundle \
	! tcpserversink host=127.0.0.1 port=5000

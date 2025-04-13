#!/bin/bash

# Kill any existing Xvfb on display :99
sudo pkill -f 'Xvfb :99'
# Wait a moment for the process to fully terminate
sleep 1

# Start Xvfb with larger screen and more features
Xvfb :99 -screen 0 1920x1080x24 -ac +extension RANDR +render -noreset &
export DISPLAY=:99

# Wait for Xvfb to be ready
for i in $(seq 1 10); do
    if xdpyinfo -display :99 >/dev/null 2>&1; then
        break
    fi
    sleep 1
done


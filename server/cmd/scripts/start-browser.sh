#!/bin/bash
export DISPLAY=:99
google-chrome \
    --no-sandbox \
    --disable-gpu \
    --disable-dev-shm-usage \
    --disable-software-rasterizer \
    --window-size=1920,1080 \
    --window-position=0,0 \
    --start-maximized \
    https://www.youtube.com/watch?v=Q86_nlRoIGw
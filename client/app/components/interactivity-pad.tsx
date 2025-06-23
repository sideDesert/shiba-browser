import {
  useState,
  useEffect,
  useLayoutEffect,
  useRef,
  type RefObject,
  useMemo,
  useId,
} from "react";
import { Button } from "./ui/button";
import { LoaderCircle } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { forwardRef } from "react";
import { Switch } from "./ui/switch";
import { CursorValue, NewCursorPayload } from "@/lib/types";
import { NewRemoteMessage } from "@/lib/chat";

export function useMouse() {
  const mouse = useRef({ x: 0, y: 0 });

  useEffect(() => {
    const handleMouseMove = (event: MouseEvent) => {
      mouse.current = { x: event.clientX, y: event.clientY };
    };

    window.addEventListener("mousemove", handleMouseMove);
    return () => window.removeEventListener("mousemove", handleMouseMove);
  }, []);

  return mouse;
}

export function useKeys() {
  const key = useRef("");

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      let _key = event.key;
      if (event.ctrlKey) {
        _key = "ctrl + " + event.key;
      }
      key.current = _key;
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  return key;
}

export function useMouseClick(ref: RefObject<any>) {
  const [click, setClick] = useState({ x: 0, y: 0 });

  useEffect(() => {
    if (!ref.current) return;
    const handleClick = (event: MouseEvent) => {
      setClick({ x: event.clientX, y: event.clientY });
    };

    window.addEventListener("click", handleClick);
    return () => window.removeEventListener("click", handleClick);
  }, [ref]);

  return click;
}

type InteractivityPadProps = {
  socket: WebSocket | null;
  handleStartStream: () => void;
  handleStopStream: () => void;
  responseIsLoading: boolean;
  streamConnectionStatus: string;
  response: object;
  userId: string;
  chatroomId: string;
  hasRemote: boolean;
};

function getActualVideoSize(elementWidth: number, elementHeight: number) {
  const naturalWidth = 1920;
  const naturalHeight = 1080;

  const naturalRatio = naturalWidth / naturalHeight;
  const elementRatio = elementWidth / elementHeight;

  let actualWidth, actualHeight;

  if (naturalRatio > elementRatio) {
    // Video is wider - limited by width
    actualWidth = elementWidth;
    actualHeight = elementWidth / naturalRatio;
  } else {
    // Video is taller - limited by height
    actualHeight = elementHeight;
    actualWidth = elementHeight * naturalRatio;
  }

  return {
    width: actualWidth,
    height: actualHeight,
    naturalWidth,
    naturalHeight,
    elementWidth,
    elementHeight,
  };
}

export const InteractivityPad = forwardRef<
  HTMLDivElement,
  InteractivityPadProps
>(
  (
    {
      socket,
      handleStartStream,
      handleStopStream,
      userId,
      responseIsLoading,
      streamConnectionStatus,
      chatroomId,
      hasRemote,
    },
    ref
  ) => {
    const [showButton, setShowButton] = useState(true);
    const isStreaming = !showButton;
    const [startStream, setStreamLoading] = useState<boolean | null>(null);
    const [sendKeys, setSendKeys] = useState(true);
    const qc = useQueryClient();

    let buttonText;
    if (startStream === null) {
      buttonText = "Start Shiba Instance";
    } else {
      if (
        startStream &&
        (streamConnectionStatus === "connecting" || responseIsLoading)
      ) {
        buttonText = (
          <>
            <LoaderCircle className="animate-spin" />
            Loading Stream...
          </>
        );
      }
      if (startStream && streamConnectionStatus === "connected") {
        buttonText = "Live Streaming ON!";
      }
    }

    return (
      <div
        style={{
          cursor: !showButton ? "none" : "auto",
        }}
        className=" h-full w-full relative flex justify-center items-center"
        ref={ref}
      >
        <>
          <Button
            style={{
              display:
                streamConnectionStatus === "connected"
                  ? "none"
                  : showButton
                  ? "flex"
                  : "none",
            }}
            disabled={!hasRemote}
            onClick={async () => {
              if (socket) {
                setStreamLoading(true);
                handleStartStream();
                await qc.fetchQuery({
                  queryKey: ["stream", chatroomId],
                });
                setShowButton(false);
                setStreamLoading(false);
              }
            }}
            className="rgb-button"
          >
            {buttonText}
          </Button>

          <video
            style={{
              display:
                streamConnectionStatus === "connected"
                  ? "block"
                  : showButton
                  ? "none"
                  : "block",
            }}
            className="border h-full w-full"
            id="video"
            autoPlay
            playsInline
          />

          <Button
            style={{
              display: !showButton ? "block" : "none",
            }}
            onClick={async () => {
              setStreamLoading(null);
              handleStopStream();
              setShowButton(true);
            }}
            variant="destructive"
            className="absolute bottom-4 right-4"
          >
            Stop Virtual Browser
          </Button>

          <div
            className="absolute items-center justify2center bottom-4 left-4 gap-2 bg-blue-100 px-3 py-2 rounded-md"
            style={{
              display: !showButton ? "flex" : "none",
            }}
          >
            <Switch
              className="inline-block bg-pink-100"
              checked={sendKeys}
              onClick={() => {
                setSendKeys(!sendKeys);
              }}
            />
            <span className="text-blue-500 text-sm">Keyboard mode</span>
          </div>
        </>
        <CursorPositionTag
          userId={userId}
          chatroomId={chatroomId}
          socket={socket}
          isStreaming={isStreaming}
        />
      </div>
    );
  }
);

function getBoundedCursorPos(
  x: number,
  y: number,
  left: number,
  right: number,
  top: number,
  bottom: number
) {
  let adjx = x;
  let adjy = y;

  if (x < left) {
    adjx = left;
  }

  if (x > right) {
    adjx = right;
  }

  if (y < top) {
    adjy = top;
  }

  if (y > bottom) {
    adjy = bottom;
  }

  return [adjx, adjy];
}

function CursorPositionTag({
  socket,
  userId,
  chatroomId,
  isStreaming,
}: {
  socket: WebSocket | null;
  userId: string;
  chatroomId: string;
  isStreaming: boolean;
}) {
  const tagRef = useRef<HTMLDivElement | null>(null);
  const [videoStuff, setVideoStuff] = useState<ReturnType<
    typeof getActualVideoSize
  > | null>(null);

  useLayoutEffect(() => {
    const tag = tagRef.current;
    if (!tag) return;

    const parent = tag.parentElement;
    if (!parent) return;

    const updateSize = () => {
      const rect = parent.getBoundingClientRect();
      const size = getActualVideoSize(rect.width, rect.height);
      setVideoStuff(size);
    };

    updateSize(); // initial size

    const resizeObserver = new ResizeObserver(updateSize);
    resizeObserver.observe(parent);

    return () => resizeObserver.disconnect();
  }, []);

  useLayoutEffect(() => {
    if (!tagRef.current) return;
    const rect = tagRef.current.getBoundingClientRect();
  }, []);

  useEffect(() => {
    const handleMouseMove = (event: MouseEvent) => {
      if (!tagRef.current) return;

      const parent = tagRef.current.parentElement;
      if (!parent) return;

      const rect = parent.getBoundingClientRect();

      let top = rect.top;
      let bottom = rect.bottom;
      let left = rect.left;
      let right = rect.right;

      if (videoStuff) {
        const edge = (rect.height - videoStuff.height) / 2;
        top += edge;
        bottom -= edge;
      }

      const [x, y] = getBoundedCursorPos(
        event.clientX,
        event.clientY,
        left,
        right,
        top,
        bottom
      );

      tagRef.current.style.left = `${x}px`;
      tagRef.current.style.top = `${y}px`;
      tagRef.current.children[0].textContent = `${scale(
        Math.max(0, x - left),
        1920,
        videoStuff ? videoStuff.width : 0
      ).toFixed(0)},`;
      tagRef.current.children[1].textContent = `${scale(
        Math.max(0, y - top),
        1080,
        videoStuff ? videoStuff.height : 0
      ).toFixed(0)}`;
    };

    function handleMouseClick(event: MouseEvent) {
      if (!tagRef.current) return;

      const parent = tagRef.current.parentElement;
      if (!parent) return;

      const rect = parent.getBoundingClientRect();

      let top = rect.top;
      let bottom = rect.bottom;
      let left = rect.left;
      let right = rect.right;

      if (videoStuff) {
        const edge = (rect.height - videoStuff.height) / 2;
        top += edge;
        bottom -= edge;
      }

      const [_x, _y] = getBoundedCursorPos(
        event.clientX,
        event.clientY,
        left,
        right,
        top,
        bottom
      );
      const x = Math.floor(
        scale(Math.max(0, _x - left), 1920, videoStuff ? videoStuff.width : 0)
      );
      const y = Math.floor(
        scale(Math.max(0, _y - top), 1080, videoStuff ? videoStuff.height : 0)
      );

      console.log({ x, y });
      const payload = NewCursorPayload(x, y, CursorValue.LeftClick);
      const message = NewRemoteMessage(userId, chatroomId, payload);
      console.log(message);
      // socket?.send(JSON.stringify(message))
    }

    const parent = tagRef.current?.parentElement;
    if (parent) {
      parent.addEventListener("click", handleMouseClick);
      parent.addEventListener("mousemove", handleMouseMove);
    }

    return () => {
      const parent = tagRef.current?.parentElement;
      if (parent) {
        parent.removeEventListener("mousemove", handleMouseMove);
        parent.removeEventListener("click", handleMouseClick);
      }
    };
  }, [videoStuff]);

  return (
    <div
      ref={tagRef}
      id="pos-tag"
      style={{
        position: "fixed",
        left: 100,
        top: 100,
        pointerEvents: "none",
      }}
      className="absolute flex"
    >
      <div className="text-xs bg-pink-600 text-white px-1.5 py-0.5 rounded-l-sm relative rounded-tl-none">
        0.0,
      </div>
      <div className="text-xs bg-pink-600 text-white px-1.5 pl-0 py-0.5 rounded-r-sm relative">
        0.0
      </div>
    </div>
  );
}

function scale(x: number, R: number, X: number) {
  return (R / X) * x;
}

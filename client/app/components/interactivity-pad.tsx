import { useState, useEffect, useRef, type RefObject } from "react";
import { Button } from "./ui/button";
import { LoaderCircle } from "lucide-react";
import { useQueryClient } from "@tanstack/react-query";
import { forwardRef } from "react";

export function useMouse() {
  const [mouse, setMouse] = useState({ x: 0, y: 0 });

  useEffect(() => {
    const handleMouseMove = (event: MouseEvent) => {
      setMouse({ x: event.clientX, y: event.clientY });
    };

    window.addEventListener("mousemove", handleMouseMove);
    return () => window.removeEventListener("mousemove", handleMouseMove);
  }, []);

  return mouse;
}

export function useKeys() {
  const [key, setKey] = useState("");

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      setKey(event.key);
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
  streamConnectionStatus: string,
  response: object;
  userId: string;
  chatroomId: string;
  hasRemote: boolean;
};

export const InteractivityPad = forwardRef<
  HTMLDivElement,
  InteractivityPadProps
>(
  (
    {
      socket,
      handleStartStream,
      handleStopStream,
      responseIsLoading,
      streamConnectionStatus,
      chatroomId,
      hasRemote,
    },
    ref
  ) => {

    const [showButton, setShowButton] = useState(true);
    const [startStream, setStreamLoading] = useState<boolean | null>(null);
    const qc = useQueryClient();


    let buttonText;
    if (startStream === null) {
      buttonText = "Start Shiba Instance"
    } else {
      if (startStream && (streamConnectionStatus === "connecting" || responseIsLoading)) {
        buttonText = <>
          <LoaderCircle className="animate-spin" />
          Loading Stream...
        </>
      }
      if (startStream && streamConnectionStatus === "connected") {
        buttonText = "Live Streaming ON!"
      }
    }



    const mouse = useMouse();
    const lastPressedKey = useKeys();
    const mouseClick = useMouseClick(ref as RefObject<HTMLDivElement>);

    function getRelativeMouse() {
      const current = (ref as RefObject<HTMLDivElement>).current;
      if (!current) {
        return { x: mouse.x, y: mouse.y };
      }

      const rect = current.getBoundingClientRect();
      return {
        x: Math.min(Math.max(0, mouse.x - rect.left), rect.width),
        y: Math.min(Math.max(0, mouse.y - rect.top), rect.height),
      };
    }

    console.log("Rerendered!!", streamConnectionStatus)

    return (
      <div
        className="h-full w-full relative flex justify-center items-center"
        ref={ref}
      >
        <>
          <Button
            style={{
              display: streamConnectionStatus === "connected" ? "none" : showButton ? "flex" : "none",
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
              display: streamConnectionStatus === "connected" ? "block" : showButton ? "none" : "block",
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
        </>
      </div>
    );
  }
);

import * as React from "react";
import { PhoneOffIcon, MicOffIcon, MinusIcon, PhoneCall } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

interface CallStatusWidgetProps {
  isVisible: boolean;
  callerName?: string;
  callerNumber?: string;
  callerImage?: string;
  status?: "calling" | "connected" | "stale";
  onEndCall?: () => void;
  onToggleMute?: () => void;
  onPlaceCall?: () => void;
  isMuted?: boolean;
  className?: string;
}

export function CallStatusWidget({
  isVisible,
  callerName = "John Doe",
  callerNumber = "+1 (555) 123-4567",
  callerImage,
  status = "calling",
  onEndCall,
  onToggleMute,
  onPlaceCall,
  isMuted = false,
  className = "",
}: CallStatusWidgetProps) {
  const [callDuration, setCallDuration] = React.useState(0);

  React.useEffect(() => {
    let interval: NodeJS.Timeout;
    if (status === "connected") {
      interval = setInterval(() => {
        setCallDuration((prev) => prev + 1);
      }, 1000);
    }
    return () => clearInterval(interval);
  }, [status]);

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, "0")}:${secs
      .toString()
      .padStart(2, "0")}`;
  };

  const getStatusText = () => {
    switch (status) {
      case "calling":
        return "Calling...";
      case "connected":
        return formatDuration(callDuration);
      case "stale":
        return "Place Call";
      default:
        return "Unknown";
    }
  };

  const getStatusColor = () => {
    switch (status) {
      case "calling":
        return "text-blue-600";
      case "connected":
        return "text-green-600";
      case "stale":
        return "text-black";
      default:
        return "text-gray-600";
    }
  };

  if (!isVisible) return null;

  return (
    <Card className={`relative mt-4 mx-2 z-50 shadow-sm border-2 ${className}`}>
      <div className="flex items-center space-x-3 p-3 bg-white rounded-lg min-w-[280px]">
        {/* Avatar */}
        <Avatar className="w-10 h-10">
          <AvatarImage
            src={callerImage || "/placeholder.svg"}
            alt={callerName}
          />
          <AvatarFallback className="text-sm">
            {callerName
              .split(" ")
              .map((n) => n[0])
              .join("")}
          </AvatarFallback>
        </Avatar>

        {/* Call info */}
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium truncate">{callerName}</p>
          <div className="flex items-center space-x-2">
            <div className="flex items-center space-x-1">
              {status === "calling" && (
                <div className="w-2 h-2 bg-blue-500 rounded-full animate-pulse"></div>
              )}
              {status === "connected" && (
                <div className="w-2 h-2 bg-green-500 rounded-full"></div>
              )}
              {status === "stale" && (
                <div className="w-2 h-2 bg-yellow-500 rounded-full animate-bounce"></div>
              )}
              <span className={`text-xs font-mono ${getStatusColor()}`}>
                {getStatusText()}
              </span>
            </div>
          </div>
        </div>

        {/* Controls */}
        <div className="flex items-center space-x-1">
          {status === "connected" && (
            <Button
              size="sm"
              variant={isMuted ? "default" : "ghost"}
              className="w-8 h-8 rounded-full p-0"
              onClick={onToggleMute}
            >
              <MicOffIcon className="w-3 h-3" />
            </Button>
          )}

          <Button
            size="sm"
            disabled={status === "calling" ? true : false}
            variant="ghost"
            className="w-8 h-8 bg-blue-500 hover:bg-blue-600 text-white hover:text-white rounded-full p-0"
            onClick={onPlaceCall}
          >
            <PhoneCall className="w-3 h-3" />
          </Button>

          <Button
            size="sm"
            disabled={status === "stale" ? true : false}
            variant="destructive"
            className="w-8 h-8 rounded-full p-0 bg-red-500 hover:bg-red-600 disabled:bg-red-300"
            onClick={onEndCall}
          >
            <PhoneOffIcon className="w-3 h-3" />
          </Button>
        </div>
      </div>
    </Card>
  );
}

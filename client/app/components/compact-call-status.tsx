import { PhoneIcon, PhoneOffIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

interface CompactCallStatusProps {
  isVisible: boolean;
  callerName?: string;
  status?: "calling" | "connected" | "on-hold";
  duration?: number;
  onEndCall?: () => void;
  className?: string;
}

export function CompactCallStatus({
  isVisible,
  callerName = "John Doe",
  status = "calling",
  duration = 0,
  onEndCall,
  className = "",
}: CompactCallStatusProps) {
  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, "0")}:${secs
      .toString()
      .padStart(2, "0")}`;
  };

  const getStatusInfo = () => {
    switch (status) {
      case "calling":
        return { text: "Calling...", color: "bg-blue-500", icon: PhoneIcon };
      case "connected":
        return {
          text: formatDuration(duration),
          color: "bg-green-500",
          icon: PhoneIcon,
        };
      case "on-hold":
        return { text: "On Hold", color: "bg-yellow-500", icon: PhoneIcon };
      default:
        return { text: "Unknown", color: "bg-gray-500", icon: PhoneIcon };
    }
  };

  if (!isVisible) return null;

  const statusInfo = getStatusInfo();
  const StatusIcon = statusInfo.icon;

  return (
    <div
      className={`flex justify-center items-center space-x-2 mt-4 ${className}`}
    >
      <Badge
        variant="secondary"
        className="flex items-center space-x-2 px-3 py-1"
      >
        <div className="flex items-center space-x-2">
          <div
            className={`w-2 h-2 rounded-full ${statusInfo.color} ${
              status === "calling" ? "animate-pulse" : ""
            }`}
          ></div>
          <StatusIcon className="w-3 h-3" />
          <span className="text-xs font-medium">{callerName}</span>
          <span className="text-xs text-muted-foreground">•</span>
          <span className="text-xs font-mono">{statusInfo.text}</span>
        </div>
      </Badge>

      <Button
        size="sm"
        variant="destructive"
        className="w-6 h-6 rounded-full p-0"
        onClick={onEndCall}
      >
        <PhoneOffIcon className="w-3 h-3" />
      </Button>
    </div>
  );
}

import * as React from "react";
import {
  PhoneIcon,
  PhoneOffIcon,
  MicOffIcon,
  VolumeXIcon,
  MessageSquareIcon,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Dialog, DialogContent } from "@/components/ui/dialog";

interface IncomingCallDialogProps {
  isOpen: boolean;
  onAnswer: () => void;
  onDecline: () => void;
  callerName?: string;
  callerNumber?: string;
  callerImage?: string;
}

export function IncomingCallDialog({
  isOpen,
  onAnswer,
  onDecline,
  callerName = "John Doe",
  callerNumber = "+1 (555) 123-4567",
  callerImage,
}: IncomingCallDialogProps) {
  const [callDuration, setCallDuration] = React.useState(0);
  const [isAnswered, setIsAnswered] = React.useState(false);

  React.useEffect(() => {
    let interval: NodeJS.Timeout;
    if (isAnswered) {
      interval = setInterval(() => {
        setCallDuration((prev) => prev + 1);
      }, 1000);
    }
    return () => clearInterval(interval);
  }, [isAnswered]);

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, "0")}:${secs
      .toString()
      .padStart(2, "0")}`;
  };

  const handleAnswer = () => {
    setIsAnswered(true);
    onAnswer();
  };

  const handleDecline = () => {
    setIsAnswered(false);
    setCallDuration(0);
    onDecline();
  };

  if (!isAnswered) {
    return (
      <Dialog open={isOpen} onOpenChange={() => {}}>
        <DialogContent className="sm:max-w-[400px] p-0 bg-gradient-to-b from-blue-800 to-blue-900 border-none text-white">
          <div className="flex flex-col items-center justify-center p-8 space-y-6">
            {/* Incoming call indicator */}
            <div className="text-center space-y-2">
              <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse mx-auto"></div>
              <p className="text-sm opacity-90">Incoming call</p>
            </div>

            {/* Caller avatar */}

            {/* Caller info */}
            <div className="text-center space-y-1">
              <h2 className="text-2xl font-semibold">{callerName}</h2>
              <p className="text-white/80">{callerNumber}</p>
            </div>

            {/* Action buttons */}
            <div className="flex items-center justify-center space-x-16 pt-8">
              {/* Decline button */}
              <Button
                size="lg"
                variant="destructive"
                className="w-16 h-16 rounded-full bg-red-500 hover:bg-red-600 border-none"
                onClick={handleDecline}
              >
                <PhoneOffIcon className="w-6 h-6" />
              </Button>

              {/* Answer button */}
              <Button
                size="lg"
                className="w-16 h-16 rounded-full bg-green-500 hover:bg-green-600 border-none"
                onClick={handleAnswer}
              >
                <PhoneIcon className="w-6 h-6" />
              </Button>
            </div>

            {/* Quick actions */}
            <div className="flex items-center space-x-8 pt-4">
              <Button
                size="sm"
                variant="ghost"
                className="text-white hover:bg-white/20 rounded-full w-12 h-12"
              >
                <MessageSquareIcon className="w-5 h-5" />
              </Button>
              <Button
                size="sm"
                variant="ghost"
                className="text-white hover:bg-white/20 rounded-full w-12 h-12"
              >
                <MicOffIcon className="w-5 h-5" />
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    );
  }
}

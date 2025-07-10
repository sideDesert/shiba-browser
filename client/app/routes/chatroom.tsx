import React from "react";
import { InteractivityPad } from "@/components/interactivity-pad";
import { get } from "@/lib/utils";
import { Anchor, Phone, PhoneOff } from "lucide-react";
import type { Route } from "./+types/home";

import type { RemoteResponse, User, ClientChatMessage } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { useEffect, useState, useRef } from "react";
import { Chat } from "@/components/chat";
import { DAL, useDAL } from "@/dal";

import { useParams } from "react-router";
import { WS_URL } from "@/root";
import { useQueryClient } from "@tanstack/react-query";
import { useQuery } from "@tanstack/react-query";
import {
  SocketMessageBuilder,
  SocketMessageHeader,
} from "@/lib/socket-message-builder";
import type { ChatMessage, ChatPayload } from "@/lib/schema/chat";
import {
  SfuSignal,
  type IcePayload,
  type SdpPayload,
  SfuType,
} from "@/lib/schema/sfu";
import {
  signal0,
  signal1,
  signal2,
  signal3,
  signal4,
  signal5,
  signal7,
  signal8,
  signal9,
  signal10,
  signal11,
  type Signal11Payload,
  type Signal2Payload,
  type Signal0Payload,
  type Signal3Payload,
} from "@/lib/schema/signal";
import {
  ShibaMessageManager,
  type SideEffects,
} from "@/lib/socket-message-manager.ts/main";
import { IncomingCallDialog } from "@/components/call-dialog";
import { CompactCallStatus } from "@/components/compact-call-status";
import { CallStatusWidget } from "@/components/call-status";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Chatroom" },
    { name: "description", content: "Welcome to Shiba Chatroom!" },
  ];
}

export default function Page() {
  const params = useParams<{ id: string }>();
  const chatroomId = params?.id!;
  // const streamPeerConnection = useRef<RTCPeerConnection | null>(null);
  // const streamIceCandidate = useRef<Array<RTCIceCandidate>>([]);
  const queryClient = useQueryClient();

  // const peerVideoRTCConn = useRef(new SFUConnection());
  // const sfuStreamRTCConn = useRef(new SFUConnection());
  const shiba = useRef<ShibaMessageManager | null>(null);

  const [input, setInput] = useState<string>("");
  const [callStatus, setCallStatus] = useState({
    incoming: false,
    status: "stale",
  });
  const [userData, userDataIsLoading] = useDAL<User>(DAL["auth"]);
  const userId = userData?.user_id;
  const [chatHistoryFn, chatHistoryKey] = DAL["chatroom"]["history"];
  const [remoteQFn, remoteQK] = DAL["remote"]["get"];
  const [showH1, setShowH1] = useState<boolean>(true);
  const [streamConnectionStatus, setStreamConnectionStatus] =
    useState("disconnected");

  const vbrowserStream = useRef<MediaStream | null>(null);
  const vref = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (typeof window !== "undefined") {
      vbrowserStream.current = new MediaStream();
    }
  }, []);

  useEffect(() => {
    if (vref.current) {
      const video = vref.current?.querySelector("video");
      video!.srcObject = vbrowserStream.current;
    }
  }, [vref.current]);

  const { data: remoteData, isLoading: remoteDataIsLoading } =
    useQuery<RemoteResponse>({
      queryFn: () => remoteQFn(chatroomId),
      queryKey: remoteQK,
      enabled: !!chatroomId,
    });

  const chk = [...chatHistoryKey, chatroomId];
  // WEBRTC
  // Track WebRTC setup to prevent infinite loop

  const candidates = useRef<Array<RTCIceCandidate>>([]);

  const [localVideoStream, setLocalVideoStream] =
    useState<MediaStream | null>();
  const [remoteVideoStream, setRemoteVideoStream] =
    useState<MediaStream | null>();

  // FOR VBROWSER STREAMING
  // const [startVirtualBrowser, setStartVirtualBrowser] =
  //   useState<boolean>(false);
  const { data: streamResponse, isLoading: streamResponseIsLoading } = useQuery(
    {
      queryFn: () => get("stream?cid=" + chatroomId),
      queryKey: ["stream", chatroomId],
      enabled: false,
    }
  );

  const localVideoRef = useRef<HTMLVideoElement | null>(null);
  const remoteVideoRef = useRef<HTMLVideoElement | null>(null);

  // function getOrCreateSubPeerConnection() {
  //   if (!subRTCPeerConn.current) {
  //     subRTCPeerConn.current = new RTCPeerConnection();
  //   }

  //   return subRTCPeerConn.current;
  // }

  // function closeRTCPeerConnection() {
  //   if (subRTCPeerConn.current) {
  //     subRTCPeerConn.current.close();
  //     subRTCPeerConn.current = null;
  //   }

  //   if (candidates.current) {
  //     candidates.current = [];
  //   }
  // }

  function closeRemoteVideoStream() {
    if (remoteVideoStream) {
      remoteVideoStream.getTracks().forEach((track) => {
        track.stop();
      });
    }
    setRemoteVideoStream(null);
    remoteVideoRef.current!.srcObject = null;
  }
  function closeLocalVideoStream() {
    if (localVideoStream) {
      localVideoStream.getTracks().forEach((track) => {
        track.stop();
      });
    }
    setLocalVideoStream(null);
    localVideoRef.current!.srcObject = null;
  }

  function endCall() {
    closeLocalVideoStream();
    closeRemoteVideoStream();
    // closeRTCPeerConnection();
  }

  useEffect(() => {
    if (localVideoRef.current && localVideoStream) {
      localVideoRef.current.srcObject = localVideoStream;
      console.log("LOCAL VIDEO", localVideoStream);
    }
  }, [localVideoStream]);

  useEffect(() => {
    if (remoteVideoRef.current && remoteVideoStream) {
      remoteVideoRef.current.srcObject = remoteVideoStream;
      console.log("REMOTE VIDEO", remoteVideoStream);
    }
  }, [remoteVideoStream]);

  const chatHistory = useQuery({
    queryKey: chk,
    queryFn: () => {
      return chatHistoryFn({
        sender: userData?.user_id!,
        cid: chatroomId,
        page: "0",
      });
    },
    enabled: !userDataIsLoading && chatroomId !== "",
  });

  const callbackMap: SideEffects = {
    addChatMessage: (msg: ChatMessage) => {
      const nmsg: ClientChatMessage = {
        chatroom_id: chatroomId,
        content: msg.payload.content,
        created_at: msg.payload.created_at,
        id: crypto.randomUUID(),
        sender: msg.sender,
        sender_name: userData?.username ?? "",
      };
      queryClient.setQueryData(chk, (p: ClientChatMessage[] | undefined) => {
        return [nmsg, ...(p || [])];
      });
    },
    setLocalVideoStream: (stream) => {
      setLocalVideoStream(stream);
    },
    setRemoteVideoStream: (stream) => {
      setRemoteVideoStream(stream);
    },
    setIsIncomingCall: (val: boolean) => {
      setCallStatus((prev) => ({
        ...prev,
        incoming: val,
      }));
    },
    setIsOutgoingCall: (val: boolean) => {
      setCallStatus({
        incoming: false,
        status: "connected",
      });
    },
  };

  // WebSocket connection setup - only runs when userData changes
  useEffect(() => {
    //   // THIS IS JUST FOR VBROWSER
    //   // if (sub.startsWith("stream")) {
    //   //   // THIS USES THE FORMAT - stream.<message-type>.<chatroom-id>.<user-id>
    //   //   if (!streamPeerConnection.current) return;
    //   //   const s = sub.split(".");
    //   //   if (s.length !== 4) {
    //   //     console.error("Invalid message format", s);
    //   //     return;
    //   //   }
    //   //   const msgType = sub.split(".")[1];
    //   //   const _cid = sub.split(".")[2];
    //   //   const _uid = sub.split(".")[3];

    //   //   if (_cid !== chatroomId) {
    //   //     console.error("Invalid chatroom ID From Server");
    //   //     return;
    //   //   }

    //   //   if (_uid !== userData?.user_id) {
    //   //     console.error("Invalid user ID From Server");
    //   //     console.log(msg);
    //   //     return;
    //   //   }

    //   //   if (msgType === "offer") {
    //   //     const offer = msg.payload as string;
    //   //     if (!streamPeerConnection) {
    //   //       return;
    //   //     }
    //   //     console.log("Offer Received:", msg);
    //   //     try {
    //   //       // THIS IS FOR THE STREAM
    //   //       await streamPeerConnection.current?.setRemoteDescription({
    //   //         type: "offer",
    //   //         sdp: offer,
    //   //       });
    //   //     } catch (err) {
    //   //       console.log("OFFER:", offer);
    //   //       console.error("Failed to set remote description", err);
    //   //     }
    //   //     console.log("Stream Remote Offer set as Remote Description");

    //   //     streamPeerConnection.current!.onicecandidate = (event) => {
    //   //       if (event.candidate) {
    //   //         try {
    //   //           socket.current?.send(
    //   //             JSON.stringify({
    //   //               sender: userData?.user_id,
    //   //               subject: "stream.ice." + chatroomId,
    //   //               payload: event.candidate,
    //   //             } as Message<typeof event.candidate>)
    //   //           );
    //   //         } catch (err) {
    //   //           console.error("Couldn't send ICE candidate", err);
    //   //           streamIceCandidate.current.push(event.candidate);
    //   //         }
    //   //       }
    //   //     };

    //   //     streamPeerConnection.current.onicegatheringstatechange = (event) => {
    //   //       const pc = streamPeerConnection.current;
    //   //       if (!pc) return;
    //   //       if (pc.iceGatheringState === "complete") {
    //   //         streamIceCandidate.current.forEach((candidate) => {
    //   //           try {
    //   //             socket.current?.send(
    //   //               JSON.stringify({
    //   //                 sender: userData?.user_id,
    //   //                 subject: "stream.ice." + chatroomId,
    //   //                 payload: candidate,
    //   //               } as Message<typeof candidate>)
    //   //             );
    //   //           } catch (err) {
    //   //             console.error("Couldn't send ICE candidate", err);
    //   //           }
    //   //         });
    //   //       }
    //   //     };

    //   //     streamPeerConnection.current.onconnectionstatechange = () => {
    //   //       console.log(
    //   //         "Connection STATUS: ",
    //   //         streamPeerConnection.current?.connectionState
    //   //       );
    //   //       setStreamConnectionStatus(
    //   //         streamPeerConnection.current?.connectionState ?? "disconnected"
    //   //       );
    //   //       if (streamPeerConnection.current?.connectionState === "connected") {
    //   //         const video = vref.current?.querySelector("video");

    //   //         if (!video) {
    //   //           console.error("no video element in vref");
    //   //           return;
    //   //         }

    //   //         video.srcObject = new MediaStream(
    //   //           streamPeerConnection.current
    //   //             ?.getReceivers()
    //   //             .map((r) => r.track)
    //   //             .filter(Boolean)
    //   //         );
    //   //         setShowH1(false);
    //   //       }
    //   //     };

    //   //     const ans = await streamPeerConnection.current?.createAnswer();
    //   //     console.log("Created Answer to offer for Stream Peer Connection");

    //   //     await streamPeerConnection.current?.setLocalDescription(ans);
    //   //     console.log("Set Answer as Local Description");

    //   //     console.log("Peer Connection Handler", streamPeerConnection);

    //   //     socket.current?.send(
    //   //       JSON.stringify({
    //   //         sender: userData?.user_id,
    //   //         subject: "stream.answer." + chatroomId,
    //   //         payload: ans,
    //   //       })
    //   //     );
    //   //   }

    //   //   if (msgType === "ice") {
    //   //     const ice = msg.payload as RTCIceCandidate;
    //   //     if (!streamPeerConnection) {
    //   //       return;
    //   //     }
    //   //     await streamPeerConnection.current?.addIceCandidate(ice);
    //   //     console.log("Added Stream Remote ICE Candidate");
    //   //   }
    //   // }
    // }

    if (
      userId &&
      userId !== "" &&
      chatroomId &&
      chatroomId !== "" &&
      typeof window !== "undefined"
    ) {
      const wsUrl = WS_URL + "?cid=" + chatroomId;
      shiba.current = new ShibaMessageManager(
        wsUrl,
        userId,
        chatroomId,
        callbackMap
      );

      shiba.current.conns.vc.sub.ontrack = (track) => {
        const stream = track.streams[0];
        setRemoteVideoStream(stream);
      };
    }

    return () => {
      if (shiba.current) {
        shiba.current.close();
      }
    };
  }, [userId, chatroomId, queryClient]);

  function onSendMessage(e: React.MouseEvent<HTMLElement>) {
    e.preventDefault();
    const senderName = userData?.name!;

    if (chatroomId !== "" && shiba.current && userId !== "" && !!userId) {
      // const wsMsg = NewWsChatMessage(senderId, senderName, input, chatroomId);
      const wsMsg = new SocketMessageBuilder();

      wsMsg.setSender(userId);
      wsMsg.setCid(chatroomId);
      wsMsg.setHeader(SocketMessageHeader.chat);
      const payload: ChatPayload = {
        sender_name: senderName,
        content: input,
        created_at: new Date().toISOString(),
        id: crypto.randomUUID(),
      };
      wsMsg.setPayload(payload);
      const localMsg = wsMsg.toChatMessage(chatroomId);

      // const localMsg = NewChatMessage(senderId, senderName, input, chatroomId);

      shiba.current.send(wsMsg.build());

      queryClient.setQueryData(chk, (p: ClientChatMessage[] | undefined) => {
        return [localMsg, ...(p || [])];
      });
    } else {
      console.error("No chatroomId given,", chatroomId);
    }

    setInput("");
  }

  const remoteStream = useRef<MediaStream | null>(null);

  useEffect(() => {
    if (typeof window !== "undefined") {
      remoteStream.current = new MediaStream();
    }
  }, []);

  async function initiateVideoCall() {
    if (shiba.current && userId) {
      // const stream = await getMediaStream();
      // if (!stream) {
      //   console.error("No media stream available");
      //   return;
      // }
      // setLocalVideoStream(stream);
      // stream.getTracks().forEach((track) => {
      //   shiba.current!.conns.vc.pub.addTrack(track, stream);
      // });
      setCallStatus({
        incoming: false,
        status: "calling",
      });

      const initIcMsg = new SocketMessageBuilder();
      initIcMsg.setSender(userId);
      initIcMsg.setHeader(SocketMessageHeader.signal);
      initIcMsg.setCid(chatroomId);
      //TODO: Actually we gotta decide if this is a friend or a groupchat
      initIcMsg.setType(signal0);
      const payload: Signal0Payload = {};
      initIcMsg.setPayload(payload);
      console.log(initIcMsg.build());
      shiba.current.send(initIcMsg.build());

      shiba.current.conns.vc.sub.ontrack = (e) => {
        remoteStream.current!.addTrack(e.track);
        setRemoteVideoStream(remoteStream.current);
      };

      // if (!subRTCPeerConn.current) {
      //   const conn = await initRTCPeerConnection(
      //     userData?.user_id!,
      //     chatroomId,
      //     socket.current,
      //     stream
      //   );
      //   if (!conn) {
      //     console.error("Initiator RTC Peer Connection Could not be created");
      //     return;
      //   }
      //   console.log("RTC Peer Connection Created");

      //   subRTCPeerConn.current = conn;

      //   if (remoteStream.current) {
      //     conn.ontrack = (e) => {
      //       console.log("Initiator received remote track:", e.track.kind);
      //       remoteStream.current!.addTrack(e.track);
      //       setRemoteVideoStream(remoteStream.current);
      //       console.log(
      //         "Remote stream now has tracks:",
      //         remoteStream
      //           .current!.getTracks()
      //           .map((t) => t.kind)
      //           .join(",")
      //       );
      //     };
      //   }
      // }
    }
  }

  function stopVideoCall() {
    endCall();
    if (shiba.current && userData && userId && userId !== "") {
      const msg = new SocketMessageBuilder();
      msg.setCid(chatroomId);
      msg.setSender(userId);
      msg.setHeader(SocketMessageHeader.signal);
      msg.setType(signal3);
      const payload: Signal3Payload = {};
      msg.setPayload(payload);
      shiba.current?.send(msg.build());
    }
    setCallStatus({
      incoming: false,
      status: "stale",
    });
  }

  function IHaveRemote() {
    return remoteData?.user_id === userData?.user_id;
  }

  return (
    <div className="relative flex flex-col mr-8 h-full px-2">
      <IncomingCallDialog
        isOpen={callStatus.incoming}
        onAnswer={() => {
          console.log("Accepted");
          setCallStatus({
            incoming: false,
            status: "connected",
          });
          shiba.current?.acceptCall();
        }}
        onDecline={() => {
          console.log("Declined");
          setCallStatus({
            incoming: false,
            status: "stale",
          });
          shiba.current?.rejectCall();
        }}
      />
      {showH1 && (
        <h1 className="text-3xl block font-semibold py-4">Chat Stream</h1>
      )}
      <div className="relative w-full border-black flex grow-1">
        <div className="w-[80%]">
          <div
            className={`h-full block border ${
              showH1 ? "border-gray-300" : "border-blue-500"
            }`}
          >
            {userData?.user_id && IHaveRemote() && (
              <InteractivityPad
                userId={userId ?? ""}
                chatroomId={chatroomId}
                socket={shiba.current}
                handleStartStream={async () => {
                  setShowH1(false);
                  console.log("Starting Stream...");
                }}
                handleStopStream={async () => {
                  setShowH1(true);
                  console.log("Stopping Stream...");

                  shiba.current?.send({
                    sender: userData?.user_id,
                    subject: "stream.stop-stream." + chatroomId,
                    payload: "",
                  });
                  // setStartVirtualBrowser(false);
                }}
                responseIsLoading={streamResponseIsLoading}
                response={streamResponse}
                streamConnectionStatus={streamConnectionStatus}
                ref={vref}
                // browserStream={vbrowserStream}
                hasRemote={IHaveRemote()}
              />
            )}
          </div>
        </div>
        <div className="w-[20%] h-full">
          <div className="flex flex-col max-h-[800px] overflow-auto">
            <div className="relative shrink-0 inline-block h-[200px] border">
              {/* REMOTE VIDEO */}
              {!IHaveRemote() && (
                <div className="mt-2 ml-2 w-8 h-8 absolute rounded-full flex justify-center items-center bg-gray-200">
                  <Anchor className="w-5 h-5" />
                </div>
              )}
              <video
                className="w-full h-full object-cover"
                ref={remoteVideoRef}
                autoPlay
                playsInline
              />
            </div>

            <div className="relative shrink-0 inline-block h-[200px] border">
              {/* LOCAL VIDEO */}
              {IHaveRemote() && (
                <div className="mt-2 ml-2 w-8 h-8 rounded-full flex justify-center items-center bg-gray-200">
                  <Anchor className="w-6 h-6" />
                </div>
              )}
              <video
                className="w-full h-full object-cover"
                ref={localVideoRef}
                autoPlay
                playsInline
              />
            </div>
          </div>
          {/* <div className="flex justify-center gap-2 mt-2">
            <Button
              className="relative w-28 inlin-block bg-blue-500 hover:bg-blue-600"
              onClick={initiateVideoCall}
              disabled={!!remoteVideoStream}
            >
              <Phone />
              Call
            </Button>
            <Button
              disabled={!remoteVideoStream}
              variant={"destructive"}
              className="w-28"
              onClick={stopVideoCall}
            >
              <PhoneOff />
              Hang up
            </Button>
          </div> */}
          <CallStatusWidget
            onPlaceCall={initiateVideoCall}
            onEndCall={stopVideoCall}
            status={callStatus.status as any}
            className="flex"
            isVisible={true}
          />
        </div>
      </div>
      {!userDataIsLoading && (
        <Chat
          input={input}
          setInput={setInput}
          messages={chatHistory.isLoading ? [] : chatHistory.data}
          onSubmitHandler={onSendMessage}
        />
      )}
    </div>
  );
}

// const initRTCPeerConnection = async (
//   userId: string,
//   chatroomId: string,
//   socket: WebSocket | null,
//   stream: MediaStream
// ) => {
//   if (!socket) {
//     console.error("Socket not initialized");
//     return;
//   }
//   try {
//     // Fetch TURN credentials
//     const response = await fetch(
//       "https://shiba-browser.metered.live/api/v1/turn/credentials?apiKey=01095344344591e468c23ab5e87951baeefc"
//     );
//     const turnServers = await response.json();
//     console.log("TURN Servers", turnServers);

//     // Create peer connection
//     const peerConnection = new RTCPeerConnection({
//       iceServers: [
//         { urls: "stun:stun.l.google.com:19302" },
//         { urls: "stun:stun1.l.google.com:19302" },
//         { urls: "stun:stun2.l.google.com:19302" },
//         ...turnServers,
//       ],
//       iceCandidatePoolSize: 10,
//     });

//     // Set up event handlers
//     peerConnection.onicecandidate = (e) => {
//       if (e.candidate && socket && peerConnection.remoteDescription) {
//         socket.send(
//           JSON.stringify(
//             NewWebrtcMessage(
//               userId,
//               chatroomId,
//               e.candidate.toJSON() as Record<string, unknown>
//             )
//           )
//         );
//       }
//     };

//     // Add local tracks to the connection
//     console.log("STREAM", stream);
//     console.log("Local Tracks", stream.getTracks());

//     stream.getTracks().forEach((track) => {
//       peerConnection.addTrack(track, stream);
//     });

//     // Create and send offer
//     const offer = await peerConnection.createOffer();
//     await peerConnection.setLocalDescription(offer);

//     const sdpMsg = NewWebrtcMessage(userId, chatroomId, offer as any);
//     console.log("Offer:", sdpMsg);
//     socket.send(JSON.stringify(sdpMsg));

//     return peerConnection;
//   } catch (error) {
//     console.error("Error in setupPeerConn:", error);
//     return;
//   }
// };

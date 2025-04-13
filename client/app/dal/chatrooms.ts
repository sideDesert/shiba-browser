import { get } from "@/lib/utils";
export async function getUserChatrooms() {
  return get("chatroom");
}

export async function getChatroomHistory(params: Record<string, string>) {
  const queryString = new URLSearchParams(params).toString();
  return get(`chatroom/history?${queryString}`);
}

export const queryKey = ["chatrooms", "get"];
export const getChatroomHistoryKey = ["chatrooms", "history"]

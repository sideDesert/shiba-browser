import { get, patch } from "@/lib/utils";
export async function getRemote(chatroomId: string) {
  return get("remote?cid=" + chatroomId);
}

export async function patchRemote(data: object) {
  return patch(`remote`, data);
}

export const getQueryKey = ["remote", "get"];
export const patchQueryKey = ["remote", "patch"];

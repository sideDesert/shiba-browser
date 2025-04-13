import { get, patch } from "@/lib/utils";

export async function getUserFriends() {
  return get("friends");
}
export async function patchFriendRequest(body: object) {
  return patch("friends", body)
}

export const queryKey = ["friends", "get"];
export const patchQueryKey = ["friends", "patch"];


import { get, patch } from "@/lib/utils";
export async function getNotifications() {
  return get("notifications");
}


export const getQueryKey = ["notifications", "get"];

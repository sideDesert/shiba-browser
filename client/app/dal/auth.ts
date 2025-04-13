import { get } from "@/lib/utils";

export async function getUserDetails() {
  return get("user");
}

export const queryKey = ["auth", "/user"];

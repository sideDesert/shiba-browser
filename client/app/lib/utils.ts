import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export async function get(endpoint: string) {
  const host = "http://localhost:9000";
  try {
    const res = await fetch(host + "/" + endpoint, {
      method: "GET",
      credentials: "include",
    });
    const body = await res.json();
    return body;
  } catch (err) {
    console.log(err);
    return null;
  }
}

export async function post(endpoint: string, body: object) {
  const host = "http://localhost:9000";
  try {
    const res = await fetch(`${host}/${endpoint}`, {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        Origin: "http://localhost:5173", // optional - only needed for testing
      },
      body: JSON.stringify(body),
    });

    const _body = await res.json();
    return _body;
  } catch (err) {
    console.log(err);
    return err;
  }
}

export async function patch<T extends object>(endpoint: string, body: T) {
  const host = "http://localhost:9000";
  try {
    const res = await fetch(`${host}/${endpoint}`, {
      method: "PATCH",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        Origin: "http://localhost:5173", // optional - only needed for testing
      },
      body: JSON.stringify(body),
    });

    const _body = await res.json();
    return _body;
  } catch (err) {
    console.log(err);
    return err;
  }
}

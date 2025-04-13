import { getUserDetails, queryKey as getUserDetailsKey } from "./auth";
import {
  getUserChatrooms,
  getChatroomHistory,
  queryKey as getUserChatroomsKey,
  getChatroomHistoryKey,
} from "./chatrooms";
import {
  getUserFriends,
  queryKey as getUserFriendsKey,
  patchFriendRequest,
  patchQueryKey as patchFriendRequestKeys,
} from "./friends";
import {
  getNotifications,
  getQueryKey as getNotificationsKey,
} from "./notifications";
import { useQuery } from "@tanstack/react-query";
import {
  getRemote,
  patchRemote,
  getQueryKey as getRemoteKey,
  patchQueryKey as patchRemoteKey,
} from "./remote";

export const DAL = {
  auth: [getUserDetails, getUserDetailsKey],
  chatroom: {
    get: [getUserChatrooms, getUserChatroomsKey],
    history: [getChatroomHistory, getChatroomHistoryKey],
  },
  friends: {
    get: [getUserFriends, getUserFriendsKey],
    patch: [patchFriendRequest, patchFriendRequestKeys],
  },
  notifications: {
    get: [getNotifications, getNotificationsKey],
  },
  remote: {
    get: [getRemote, getRemoteKey],
    patch: [patchRemote, patchRemoteKey],
  },
} as const;

type DALInput = readonly [() => Promise<any>, string[]];
type DALInputQuery = readonly [
  (params: Record<string, string>) => Promise<any>,
  string[]
];

export function useDAL<T>(dalt: DALInput, refetchInterval = Infinity) {
  const [fn, key] = dalt;

  const { data, isLoading } = useQuery<T>({
    queryKey: key,
    queryFn: fn,
    refetchInterval: refetchInterval,
  });

  return [data, isLoading] as const;
}

export function useDALQuery<T>(
  dalt: DALInputQuery,
  queryParams: Record<string, string>,
  refetchInterval = Infinity
) {
  const [fn, key] = dalt;

  const { data, isLoading } = useQuery<T>({
    queryKey: key,
    queryFn: () => {
      return fn(queryParams);
    },
    refetchInterval: refetchInterval,
  });

  return [data, isLoading] as const;
}

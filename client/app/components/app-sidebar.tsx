import { PlusIcon, User2, } from "lucide-react";
import { Label } from "@radix-ui/react-dropdown-menu";
import { useMutation } from "@tanstack/react-query";
import { Input } from "./ui/input";
import { useState } from "react";
import type { Friend, User } from "@/lib/types";
import { DAL } from "@/dal";

import AsyncSelect from 'react-select/async'

import {
  DropdownMenu,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuContent,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

import { Button } from "./ui/button";
import { useDAL } from "@/dal";

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { type UserChatroom } from "@/lib/types";
import { Skeleton } from "./ui/skeleton";
import { useNavigate, useParams } from "react-router";
import { post } from "@/lib/utils";
import type { FriendStatusResponseRequest } from "@/lib/requests";
import { queryClient } from "@/root";


const s = () => {
  return Math.floor(400 * Math.random()) + 200
}
const IMAGE_URL = "https://picsum.photos"

function img() {
  return IMAGE_URL + "/" + String(s())
}



interface SearchUser {
  id: number;
  name: string;
  user_id: string;
  username: string;
  email: string;
  status?: { valid: boolean; string: string };
  profile_picture?: { valid: boolean; string: string };
}

interface SelectOption {
  label: string;
  value: string;
  userId: string;
  username: string;
  email: string;
  status: string | null;
  profilePicture: string | null;
}

interface AsyncMultiSelectProps {
  excludeUserIds?: string[]; // 👈 pass friends' user_ids here
  onChange?: (selected: SelectOption[]) => void;
  isMulti?: boolean
}

const fetchUsers = async (input: string): Promise<SearchUser[]> => {
  const res = await fetch(`http://localhost:9000/search?q=${input}`, {
    credentials: "include",
  });
  if (!res.ok) throw new Error("Failed to fetch users");
  return res.json();
};

const AsyncMultiSelect: React.FC<AsyncMultiSelectProps> = ({ excludeUserIds = [], isMulti = true, onChange }) => {
  const [selected, setSelected] = useState<SelectOption[]>([]);

  const loadOptions = async (inputValue: string): Promise<SelectOption[]> => {
    if (inputValue.length < 1) return [];

    try {
      const data = await fetchUsers(inputValue);

      const filtered = data.filter(
        (user) => !excludeUserIds.includes(user.user_id)
      );

      return filtered.map((item) => ({
        label: item.name,
        value: item.user_id,
        userId: item.user_id,
        username: item.username,
        email: item.email,
        status: item.status?.valid ? item.status.string : null,
        profilePicture: item.profile_picture?.valid ? item.profile_picture.string : null,
      }));
    } catch (err) {
      console.error("Search failed:", err);
      return [];
    }
  };

  const CustomSingleValue = ({ data }: { data: SelectOption }) => (
    <div className="flex items-center gap-2">
      {data.profilePicture && (
        <img src={data.profilePicture ?? img()} alt="Profile" className="w-6 h-6 rounded-full" />
      )}
      <div>
        <div className="font-medium">{data.label}</div>
        <div className="text-blue-500 text-sm">@{data.username}</div>
      </div>
    </div>
  );

  const CustomOption = (props: any) => {
    const { data, innerRef, innerProps } = props;
    return (
      <div
        ref={innerRef}
        {...innerProps}
        className="flex items-center gap-2 p-2 hover:bg-gray-100 cursor-pointer"
      >
        {data.profilePicture && (
          <img src={data.profilePicture ?? img()} alt="Profile" className="w-8 h-8 rounded-full" />
        )}
        <div>
          <div className="font-medium">{data.label}</div>
          <div className="text-blue-500 text-sm">@{data.username}</div>
        </div>
      </div>
    );
  };

  return (
    <AsyncSelect
      cacheOptions
      defaultOptions
      loadOptions={loadOptions}
      value={selected}
      onChange={(selectedOptions) => {
        const values = selectedOptions as SelectOption[];
        setSelected(values);
        onChange?.(values);
      }}
      isMulti={isMulti}
      components={{ SingleValue: CustomSingleValue, Option: CustomOption }}
      getOptionValue={(e) => e.value.toString()}
      placeholder="Search and select..."
    />
  );
};

export default AsyncMultiSelect;


export function AppSidebar() {
  const nav = useNavigate();
  const params = useParams()
  const chatroomId = params.id!
  const [selectedMenuItem, setSelectedMenuItem] = useState<any>()

  const [authData] = useDAL<User>(DAL["auth"])
  // TODO: Handle refetch via nats
  //
  const [chatroomsData, chatroomsDataIsLoading] = useDAL<{ chatrooms: UserChatroom[] }>(DAL["chatroom"]["get"])
  const [friendsData, friendsDataIsLoading] = useDAL<Array<Friend>>(DAL["friends"]["get"], 5000)
  const [notificationsData, notificationsDataIsLoading] = useDAL<Array<Friend>>(DAL["notifications"]["get"], 5000)

  const [friendRequestPatchFn, friendRequestPatchKey] = DAL["friends"]["patch"]

  const notificationMutation = useMutation({
    mutationFn: friendRequestPatchFn,
    mutationKey: friendRequestPatchKey,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: DAL["friends"]["get"][1]
      })
      queryClient.invalidateQueries({
        queryKey: DAL["notifications"]["get"][1]
      });
    },
  })

  function onChatroomButtonClickHandler(chatroomId: string) {
    nav("dashboard/chat/" + chatroomId);
  }

  function acceptFriendRequest(requestId: string) {
    const req: FriendStatusResponseRequest = {
      status: "accepted",
      id: requestId
    }
    notificationMutation.mutate(req)
  }

  function declineFriendRequest(requestId: string) {
    const req: FriendStatusResponseRequest = {
      status: "blocked",
      id: requestId
    }
    notificationMutation.mutate(req)
  }

  return (
    <Sidebar>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Friends</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {
                friendsDataIsLoading ?
                  (
                    <SidebarMenuItem className="">
                      <SidebarMenuButton asChild>
                        <Skeleton className="p-2 h-fit text-left flex justify-start animate-pulse">
                          <div className="h-8 w-8 bg-transparent rounded-full">
                          </div>
                          <div></div>
                        </Skeleton>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  )
                  :

                  Array.isArray(friendsData) && friendsData.length > 0 && friendsData.map((item) => (
                    item.friend_status === 'accepted' &&
                    <SidebarMenuItem key={item.user_id}>
                      <SidebarMenuButton asChild>
                        <Button
                          className={`p-2 h-fit text-left flex justify-start ${chatroomId === item?.chatroom_id ? 'bg-blue-200 hover:bg-blue-300' : 'bg-transparent hover:bg-neutral-200'}`}
                          variant={"ghost"}
                          onClick={() => {
                            setSelectedMenuItem(item.chatroom_id)
                            onChatroomButtonClickHandler(String(item.chatroom_id));
                          }}
                        >
                          <div className="h-8 w-8 bg-neutral-300 rounded-full overflow-hidden">

                            <img src={img()} className='w-full h-full ' />
                          </div>
                          <div className='flex flex-col'>
                            <div>{item.name}</div>
                            <p className='text-blue-500 text-[10px]'>@{item.username}</p>
                          </div>
                        </Button>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  ))

              }
              <SidebarMenuItem className="">

                <SidebarMenuButton asChild>
                  <AddFriendButton />
                </SidebarMenuButton>

              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
        <SidebarGroup>
          <SidebarGroupLabel>Rooms</SidebarGroupLabel>

          <SidebarGroupContent>
            <SidebarMenu>
              {!chatroomsDataIsLoading ? (
                Array.isArray(chatroomsData?.chatrooms) &&
                chatroomsData?.chatrooms.map((room) => !room?.direct_message && (
                  <SidebarMenuItem className="" key={room?.id}>
                    <SidebarMenuButton asChild>
                      <Button
                        className={`p-2 h-fit text-left flex justify-start ${chatroomId === room?.id ? 'bg-blue-300' : 'bg-transparent'}`}
                        variant={"ghost"}
                        onClick={() => {
                          setSelectedMenuItem(room?.id)
                          onChatroomButtonClickHandler(String(room?.id));
                        }}
                      >
                        <div className="h-8 w-8 bg-neutral-200 rounded-full"></div>
                        <div>{room.name}</div>
                      </Button>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))
              ) : (
                <SidebarMenuItem className="">
                  <SidebarMenuButton asChild>
                    <Skeleton className="p-2 h-fit text-left flex justify-start animate-pulse">
                      <div className="h-8 w-8 bg-transparent rounded-full"></div>
                      <div></div>
                    </Skeleton>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              )}
              <SidebarMenuItem className="">
                <SidebarMenuButton asChild>
                  <CreateChatRoomButton />
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        <SidebarGroup>
          <SidebarGroupLabel>Friend Requests</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {
                notificationsDataIsLoading ?
                  (
                    <SidebarMenuItem className="">
                      <SidebarMenuButton asChild>
                        <Skeleton className="p-2 h-fit text-left flex justify-start animate-pulse">
                          <div className="h-8 w-8 bg-transparent rounded-full"></div>
                          <div></div>
                        </Skeleton>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  )
                  :

                  Array.isArray(notificationsData) && notificationsData.length > 0 && notificationsData.map((item) => (item?.request_id &&
                    <SidebarMenuItem key={item.user_id}>
                      <div className="p-3 flex flex-col items-center gap-4 bg-white rounded-lg shadow-md hover:shadow-lg transition">
                        {/* Avatar */}
                        <div className='flex flex-row gap-2 justify-start items-center'>
                          <div className="h-10 w-10 bg-neutral-200 rounded-full"></div>

                          {/* Name & Action Buttons */}
                          <div className="flex-1">
                            <div className="text-lg font-semibold">{item.name}</div>
                          </div>
                        </div>

                        {/* Buttons */}
                        <div className="flex flex-row gap-2">
                          <Button
                            onClick={e => {
                              e.preventDefault()
                              declineFriendRequest(item?.request_id!)
                            }}
                            variant='destructive'>
                            Decline
                          </Button>
                          <Button onClick={(e) => {
                            e.preventDefault()
                            acceptFriendRequest(item?.request_id!)
                          }} className="bg-blue-500 hover:bg-blue-600 text-white px-4 !py-1 rounded-md transition">
                            Accept
                          </Button>
                        </div>
                      </div>
                    </SidebarMenuItem>
                  ))

              }
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton className="flex flex-col h-[5rem]">
                  <div className="flex justify-center gap-2 items-center">
                    <User2 strokeWidth={3} />
                    <div className="text-xl font-semibold">
                      {authData?.name}
                    </div>
                  </div>
                  <div className="flex justify-center items-center">
                    <h2 className="text-sm text-blue-600">
                      @{authData?.username}
                    </h2>
                  </div>
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-[120px]">
                <DropdownMenuItem>
                  <span>Settings</span>
                </DropdownMenuItem>
                <DropdownMenuItem>
                  <span>Sign out</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
  );
}

export function CreateChatRoomButton() {
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button className="p-2 w-full h-fit text-left flex" variant={"outline"}>
          <div className="flex justify-center items-center rounded-full">
            <PlusIcon />
          </div>
          <div>Create Chat Room</div>
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Create Chat Room</DialogTitle>
          <DialogDescription>
            Create Chat Room with Participants
          </DialogDescription>
        </DialogHeader>

        <main className="flex flex-col gap-6">
          <div className="">
            <div className="flex flex-col flex-1 justify-center gap-2">
              <Label className="">Name</Label>

              <Input id="link" defaultValue="" />
            </div>
          </div>
          <div className="flex flex-col flex-1 justify-center gap-2">
            <h2 className="text-xl font-bold mt-3">Add People</h2>
            <div className="relative">
              <AsyncMultiSelect />
            </div>
          </div>
        </main>

        <DialogFooter className="sm:justify-start">
          <Button type="button">
            Create Room
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}


const sendFriendRequest = async (userData: any) => {
  return post("friends", userData)
}

export function AddFriendButton() {
  const [friendsData] = useDAL<Array<Friend>>(DAL["friends"]["get"])
  const [selected, setSelected] = useState<{ label: string, value: string } | null>(null)
  const mutation = useMutation({ mutationKey: ["friends", "sendFr"], mutationFn: sendFriendRequest })

  function onClickHandler(e: React.MouseEvent<HTMLButtonElement, MouseEvent>) {
    e.preventDefault()
    if (selected && selected?.value) {

      const userData = { "friend_id": selected.value }
      mutation.mutate(userData)
    } else {
      console.warn("NULL Selected")
    }
    setSelected(null)
  }

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button className="p-2 w-full h-fit text-left flex" variant={"outline"}>
          <div className="flex justify-center items-center rounded-full">
            <PlusIcon />
          </div>
          <div>Add Friend</div>
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Add Friend</DialogTitle>
          <DialogDescription>
            Search by name, username
          </DialogDescription>
        </DialogHeader>

        <main className="flex flex-col gap-6">
          <div className="flex flex-row flex-1 justify-center gap-2">
            <div className="relative w-full">
              <Label className="sr-only">Search People</Label>
              <AsyncMultiSelect
                isMulti={false}
                onChange={(val) => {
                  setSelected(val as any)
                }}
                excludeUserIds={friendsData?.map(f => f.user_id)}
              />
            </div>
            <Button
              onClick={onClickHandler}
              disabled={mutation.isPending}
              className={
                mutation.isSuccess
                  ? 'bg-green-600'
                  : mutation.isError
                    ? 'bg-red-600'
                    : ''
              }
            >
              {mutation.isPending
                ? 'Sending...'
                : mutation.isSuccess
                  ? 'Sent!'
                  : 'Send Request'}
            </Button>
          </div>
        </main>

      </DialogContent>
    </Dialog>
  );
}

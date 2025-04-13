import { z } from "zod";

export const UserChatroomSchema = z.object({
  id: z.string(),
  name: z.string(),
  profile_picture: z.object({
    String: z.string(),
    Valid: z.boolean(),
  }),
  direct_message: z.boolean(),
  created_at: z.date(),
});

export type UserChatroom = z.infer<typeof UserChatroomSchema>;

export type Message<T> = {
  subject: string;
  sender: string;
  payload: T;
};

export type ChatMessage = {
  chatroom_id: string;
  sender: string;
  id: string;
  sender_name: string;
  content: string;
  created_at: string;
};

export type ChatMessagePayload = {
  id: string;
  sender_name: string;
  content: string;
  created_at: string;
};

export const UserSchema = z.object({
  user_id: z.string(),
  name: z.string(),
  username: z.string(),
  email: z.string().email(), // Added email validation
  profile_picture: z.string().nullable(), // profile_picture can be null
  created_at: z.string().datetime().nullable(), // created_at can be null and is a datetime string.
});

// Example usage (type inference):
export type User = z.infer<typeof UserSchema>;

const StringValidSchema = z.object({
  String: z.string().nullable(), // String can be null
  Valid: z.boolean(),
});

export const FriendSchema = z.object({
  user_id: z.string(),
  name: z.string(),
  username: z.string(),
  profile_picture: StringValidSchema,
  user_status: StringValidSchema,
  friend_status: z.string().nullable(),
  request_id: z.string().nullable(),
  chatroom_id: z.string().nullable(),
});

export type Friend = z.infer<typeof FriendSchema>;

export const PatchChatroomRemoteRequestSchema = z.object({
  chatroom_id: z.string().uuid(),
  user_id: z.string().uuid(),
});
export type PatchChatroomRemoteRequest = z.infer<
  typeof PatchChatroomRemoteRequestSchema
>;

export const RemoteResponseSchema = z.object({
  name: z.string(),
  username: z.string(),
  user_id: z.string(),
  status: z.string(),
});

// TypeScript Type (using Zod inference)
export type RemoteResponse = z.infer<typeof RemoteResponseSchema>;

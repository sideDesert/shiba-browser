export type SocketMessage<T> = {
  // Subject can we - chatroom.chat.<cid>, chatroom.sfu.<cid>, chatroom.signal.<cid>
  subject: string;
  // Sender is always userId
  sender: string;
  // Payload depends on different type of messages
  payload: T;
};

export type SocketMessage<T> = {
  subject: string;
  sender: string;
  payload: T;
};

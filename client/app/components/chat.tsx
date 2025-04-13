import { Button } from "./ui/button";
import { useState } from "react";
import { type ChatMessage, type Message } from "@/lib/types";
import { queryClient } from "@/root";
import { DAL } from "@/dal";

type ChatProps = {
  messages: ChatMessage[];
  onSubmitHandler: React.MouseEventHandler<HTMLButtonElement>;
  input: string;
  setInput: (t: string) => void;
};

export function Chat({
  messages,
  onSubmitHandler,
  input,
  setInput,
}: ChatProps) {
  const [open, useOpen] = useState(false);
  const data: any = queryClient.getQueryData(DAL["auth"][1])

  const userId = data?.user_id ?? ""
  return (
    <>
      <Button
        className="fixed bottom-4 right-4 flex items-center justify-center text-sm font-medium disabled:pointer-events-none disabled:opacity-50 border rounded-full w-16 h-16 bg-black hover:bg-gray-700 cursor-pointer border-gray-200 p-0 leading-5 hover:text-gray-900"
        type="button"
        onClick={() => {
          useOpen(!open);
        }}
        aria-haspopup="dialog"
        aria-expanded="false"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="30"
          height="40"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="text-white"
        >
          <path d="m3 21 1.9-5.7a8.5 8.5 0 1 1 3.8 3.8z" />
        </svg>
      </Button>

      {open && (
        <div className="fixed flex flex-col bottom-20 right-4 bg-white p-6 rounded-lg border border-gray-200 w-[440px] h-[634px]">
          <div className="pb-6">
            <h2 className="font-semibold text-lg">Messages</h2>
          </div>
          <div className="grow flex flex-col-reverse overflow-auto pr-4 pb-2">
            {userId !== "" && messages.map((msg) => {
              return <Message key={msg.id} orientation={`${msg.sender === userId ? 'right' : 'left'}`} sender={msg?.sender_name} children={msg?.content} />
            })}
          </div>

          <div className="flex items-center pt-2">
            <form className="flex w-full space-x-2">
              <input
                onChange={(e) => {
                  setInput(e.target.value);

                }}
                value={input}
                className="h-10 w-full rounded-md border border-gray-300 px-3 py-2 text-sm placeholder-gray-500 focus:ring-2 focus:ring-gray-400"
                placeholder="Type your message"
              />
              <Button
                className="rounded-md text-sm font-medium text-white bg-black hover:bg-gray-800 h-10 px-4"
                type="submit"
                onClick={onSubmitHandler}
              >
                Send
              </Button>
            </form>
          </div>
        </div>
      )}
    </>
  );
}

const ppurl =
  "https://gratisography.com/wp-content/uploads/2025/03/gratisography-funny-dog-1036x780.jpg";

function Message({
  sender,
  children,
  orientation,
  profilePictureUrl = ppurl,
}: {
  sender: string;
  children: React.ReactNode;
  orientation: "left" | "right";
  profilePictureUrl?: string;
}) {
  const isRight = orientation === "right";

  return (
    <div className={`flex ${isRight ? "justify-end" : ""}`}>
      <div
        className={`flex items-end gap-3 my-2 max-w-xs ${isRight ? "bg-blue-500 text-white p-3 rounded-lg" : "text-gray-600"
          }`}
      >
        {!isRight && (
          <span className="w-10 h-10 rounded-full overflow-hidden bg-gray-100 border shrink-0">
            <img
              className="w-full h-full object-cover"
              src={profilePictureUrl}
            />
          </span>
        )}
        <div>
          <span
            className={`block font-bold ${isRight ? "text-white" : "text-gray-700"
              }`}
          >
            {sender}
          </span>
          <p>{children}</p>
        </div>
      </div>
    </div>
  );
}

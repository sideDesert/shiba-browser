import type { Route } from "../+types/home";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Dashboard" },
    { name: "description", content: "Welcome to Shiba Dashboard!" },
  ];
}

export default function Page() {
  return (
    <div className="m-4">
      <h2 className="text-xl font-semibold">
        Hello! Welcome to Shiba Dashboard
      </h2>
    </div>
  );
}

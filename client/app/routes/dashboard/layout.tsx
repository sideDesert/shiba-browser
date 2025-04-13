import type { Route } from "../+types/home";
import { useAuth } from "@/hooks/use-auth";
import { useEffect } from "react";
import { Outlet, useNavigate } from "react-router";
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/app-sidebar";

export function meta({}: Route.MetaArgs) {
  return [
    { title: "Dashboard" },
    { name: "description", content: "Welcome to Shiba Dashboard!" },
  ];
}

export default function Page() {
  const authorized = useAuth();
  const nav = useNavigate();

  useEffect(() => {
    if (!authorized) {
      nav("login");
    }
  }, []);

  if (authorized) {
    return <DashboardLayout />;
  }

  return (
    <div className="m-2">
      <h2>Unauthorized</h2>
    </div>
  );
}

function DashboardLayout() {
  return (
    <>
      <SidebarProvider>
        <AppSidebar />
        <>
          <SidebarTrigger />
        </>
        <div className="grow">
          <Outlet />
        </div>
      </SidebarProvider>
    </>
  );
}

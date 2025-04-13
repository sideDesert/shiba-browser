import type { Route } from "./+types/home";
import { useNavigate } from "react-router";
import { useQuery } from "@tanstack/react-query";
import { useEffect } from "react";
import { DAL } from "@/dal";

export function meta({ }: Route.MetaArgs) {
  return [
    { title: "Shiba Browser" },
    { name: "description", content: "Welcome to Shiba Browser!" },
  ];
}

async function fetchHealth() {
  const res = await fetch("http://localhost:9000/health")
  const body = await res.json()
  return body
}


export default function Home() {
  const test = useQuery({
    queryKey: ['test'],
    queryFn: fetchHealth
  })

  // TODO: Make the types use zod
  const [userQfn, userQk] = DAL["auth"]

  const user = useQuery<{ error: string } | { [key: string]: unknown }>({
    queryKey: [...userQk],
    queryFn: userQfn
  })

  useEffect(() => {
    if (user.data?.error) {
      navigate("/login")
    } else {
      navigate("/dashboard")
    }
  }, [user.data])

  const navigate = useNavigate()
  return <h2>Shiba Browser</h2>
}

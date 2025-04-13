import { DAL } from "@/dal"
import { useQuery } from "@tanstack/react-query"

export function useAuth() {
  const [fn, keys] = DAL["auth"]
  const user = useQuery({
    queryKey: keys,
    queryFn: fn
  })

  if (user.data?.error) {
    return false
  }

  //TODO: Better Auth Check
  return true
}

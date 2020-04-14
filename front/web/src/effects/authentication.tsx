import React, { createContext, useState, useEffect, FC, Dispatch, SetStateAction } from "react"
import { AuthenticationStatus, statusInitialized, statusUnauthenticated, newSignedInStatus, User } from "../auth"
import { getItem, setItem } from "../local-storage"

const AuthenticationContext = createContext<AuthenticationStatus>(statusInitialized)

export const useAuthentication = (): [AuthenticationStatus, Dispatch<SetStateAction<AuthenticationStatus>>] => {
  const [status, setStatus] = useState<AuthenticationStatus>(statusInitialized)

  useEffect(() => {
    const get = (): void => {
      const user = getItem("github-user-access-token")
      if (user === null) {
        setStatus(statusUnauthenticated)
        return
      }
      setStatus(newSignedInStatus(user))
    }

    get()
  }, [])

  return [status, setStatus]
}

export const AuthenticationProvider: FC = ({ children }) => {
  const [status] = useAuthentication()
  return <AuthenticationContext.Provider value={status}>{children}</AuthenticationContext.Provider>
}

export const useSignIn = (): ((user: User) => void) => {
  const [_, setStatus] = useAuthentication()
  return (user: User): void => {
    setItem("github-user-access-token", user)
    setStatus(newSignedInStatus(user))
  }
}

export const useSignOut = (): (() => void) => {
  const [_, setStatus] = useAuthentication()
  return (): void => {
    setStatus(statusUnauthenticated)
  }
}

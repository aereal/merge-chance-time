import React, { FC, useEffect, useState } from "react"
import Grid from "@material-ui/core/Grid"
import Typography from "@material-ui/core/Typography"
import { isSignedIn } from "../auth"
import { useAuthentication } from "../effects/authentication"

interface Token {
  readonly auth_time: number
}

interface Claims {
  readonly name: string
  readonly picture: string
}

interface Repository {
  readonly full_name: string
}

interface Data {
  readonly token: Token
  readonly claims: Claims
  readonly repositories: Repository[]
}

interface ErrorPayload {
  readonly Error: string
}

type ResponsePayload = Data | ErrorPayload

export const RootPage: FC = () => {
  const [authStatus] = useAuthentication()
  const [data, setData] = useState<ResponsePayload>()
  useEffect((): void => {
    if (!isSignedIn(authStatus)) {
      return
    }

    const fetchData = async () => {
      const resp = await fetch("http://localhost:8000/api/me", {
        headers: {
          authorization: `Bearer ${authStatus.user.token}`,
        },
      })
      const payload = await resp.json()
      setData(payload)
    }
    fetchData()
  }, [authStatus.type])

  return (
    <>
      <Grid item xs={12}>
        <Typography variant="h1">Merge Chance Time</Typography>
        <pre>{JSON.stringify(data, undefined, "  ")}</pre>
      </Grid>
    </>
  )
}

import React, { FC, useEffect, useState } from "react"
import Grid from "@material-ui/core/Grid"
import Typography from "@material-ui/core/Typography"
import { useAuthentication, isSignedIn } from "../effects/authentication"

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
  const authStatus = useAuthentication()
  const [data, setData] = useState<ResponsePayload>()
  useEffect((): void => {
    if (!isSignedIn(authStatus)) {
      return
    }

    const fetchData = async () => {
      const resp = await fetch("http://localhost:8000/api/me", {
        headers: {
          authorization: `Bearer ${authStatus.user.idToken}`,
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
        <Payload payload={data} />
        <pre>{JSON.stringify(data, undefined, "  ")}</pre>
      </Grid>
    </>
  )
}

const Payload: FC<{ readonly payload?: ResponsePayload }> = ({ payload }) => {
  if (payload === undefined) {
    return <>empty payload</>
  }

  if ("Error" in payload) {
    return <>Error: {payload.Error}</>
  }

  return (
    <>
      {payload.repositories.map((repo) => (
        <RepositoryItem repo={repo} key={repo.full_name} />
      ))}
    </>
  )
}

import Paper from "@material-ui/core/Paper"
const RepositoryItem: FC<{ readonly repo: Repository }> = ({ repo: { full_name: fullName } }) => (
  <Paper>
    <Typography variant="body1">{fullName}</Typography>
  </Paper>
)

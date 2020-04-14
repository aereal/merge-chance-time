import React, { FC, useEffect, useState } from "react"
import Grid from "@material-ui/core/Grid"
import Typography from "@material-ui/core/Typography"
import { useAuthentication } from "../effects/authentication"
import { isSignedIn } from "../auth"

export const RootPage: FC = () => {
  const [authStatus] = useAuthentication()
  const [data, setData] = useState()

  useEffect(() => {
    if (!isSignedIn(authStatus)) {
      return
    }

    const fetchData = async () => {
      const resp = await fetch("http://localhost:8000/api/user/installed_repos", {
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

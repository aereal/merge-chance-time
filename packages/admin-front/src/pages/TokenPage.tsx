import React, { FC, useState, useEffect } from "react"
import Typography from "@material-ui/core/Typography"
import Grid from "@material-ui/core/Grid"
import TextField from "@material-ui/core/TextField"
import { useAuthentication } from "../effects/authentication"
import { isSignedIn } from "../auth"

const shouldShow = process.env.NODE_ENV === "development"

export const TokenPage: FC = () => {
  const [authStatus] = useAuthentication()
  const [authz, setAuthz] = useState<string>()
  useEffect(() => {
    if (isSignedIn(authStatus)) {
      setAuthz(`Bearer ${authStatus.user.token}`)
    }
  }, [authStatus.type])

  if (!shouldShow) {
    return null
  }

  return (
    <Grid item xs={12}>
      <Typography variant="subtitle1">My Token</Typography>
      <TextField value={authz} />
    </Grid>
  )
}

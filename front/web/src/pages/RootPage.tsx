import React, { FC } from "react"
import Grid from "@material-ui/core/Grid"
import Typography from "@material-ui/core/Typography"
import { useAuthentication, isSignedIn } from "../effects/authentication"

export const RootPage: FC = () => {
  const authStatus = useAuthentication()
  console.log(`auth status = ${JSON.stringify(authStatus)}`)

  if (!isSignedIn(authStatus)) {
    return null
  }

  return (
    <>
      <Grid item xs={12}>
        <Typography variant="h1">Merge Chance Time</Typography>
      </Grid>
    </>
  )
}

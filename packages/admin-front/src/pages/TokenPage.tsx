import React, { FC, useState, useEffect } from "react"
import Typography from "@material-ui/core/Typography"
import Grid from "@material-ui/core/Grid"
import TextField from "@material-ui/core/TextField"
import { auth } from "firebase"
import { useAuthentication } from "../effects/authentication"
import { isSignedIn } from "../auth"

// https://merge-chance-time.firebaseapp.com/__/auth/handler
const shouldShow = process.env.NODE_ENV === "development"

export const TokenPage: FC = () => {
  const [authStatus] = useAuthentication()
  const [token, setToken] = useState<string>()
  useEffect(() => {
    if (isSignedIn(authStatus)) {
      setToken(authStatus.user.token)
    }
  }, [authStatus.type])

  if (!shouldShow) {
    return null
  }
  const authz = token ? `Bearer ${token}` : undefined
  const cred = token ? auth.GithubAuthProvider.credential(token) : undefined

  return (
    <Grid item xs={12}>
      <Typography variant="subtitle1">My Token</Typography>
      <div>
        {token ? (
          <TextField label="Authorization Header" value={authz} />
        ) : (
          <TextField label="Authorization Header" disabled />
        )}
      </div>
      {cred ? (
        <>
          <div>
            <TextField label="Access Token" value={cred.accessToken} fullWidth />
          </div>
          <div>
            <TextField label="ID Token" value={cred.idToken} fullWidth />
          </div>
          <pre>{JSON.stringify(cred.toJSON(), null, "  ")}</pre>
        </>
      ) : (
        <>
          <div>
            <TextField label="Access Token" disabled />
          </div>
          <div>
            <TextField label="ID Token" disabled />
          </div>
        </>
      )}
    </Grid>
  )
}

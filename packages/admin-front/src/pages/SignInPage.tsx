import React, { FC, FormEvent } from "react"
import Button from "@material-ui/core/Button"
import makeStyles from "@material-ui/core/styles/makeStyles"
import Avatar from "@material-ui/core/Avatar"
import LockOutlinedIcon from "@material-ui/icons/LockOutlined"
import Typography from "@material-ui/core/Typography"
import { Helmet } from "react-helmet"
import { apiOrigin } from "../api-origin"

const useStyles = makeStyles((theme) => ({
  paper: {
    marginTop: theme.spacing(8),
    display: "flex",
    flexDirection: "column",
    alignItems: "center",
  },
  avatar: {
    margin: theme.spacing(1),
    backgroundColor: theme.palette.secondary.main,
  },
  form: {
    width: "100%",
    marginTop: theme.spacing(1),
  },
  submit: {
    margin: theme.spacing(3, 0, 2),
  },
}))

export const SignInPage: FC = () => {
  const { paper, avatar, form, submit } = useStyles()

  const onSubmit = (event: FormEvent): void => {
    event.preventDefault()
    const params = new URLSearchParams()
    params.set("initiator_url", window.location.origin + "/auth/callback")
    window.location.href = `${apiOrigin()}/auth/start?${params.toString()}`
  }

  return (
    <>
      <Helmet>
        <title>Sign In - Merge Chance Time</title>
      </Helmet>
      <div className={paper}>
        <Avatar className={avatar}>
          <LockOutlinedIcon />
        </Avatar>
        <Typography variant="h5" component="h1">
          Sign in
        </Typography>
        <form className={form} noValidate onSubmit={onSubmit}>
          <Button className={submit} type="submit" fullWidth variant="contained" color="primary">
            Sign in with GitHub
          </Button>
        </form>
      </div>
    </>
  )
}

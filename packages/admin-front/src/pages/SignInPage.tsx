import React, { FC, FormEvent } from "react"
import Grid from "@material-ui/core/Grid"
import Button from "@material-ui/core/Button"
import makeStyles from "@material-ui/core/styles/makeStyles"
import Avatar from "@material-ui/core/Avatar"
import LockOutlinedIcon from "@material-ui/icons/LockOutlined"
import { Helmet } from "react-helmet"
import { apiOrigin } from "../api-origin"

const useStyles = makeStyles((theme) => ({
  paper: {
    marginTop: theme.spacing(8),
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
  const classes = useStyles()

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
      <Grid item xs={12}>
        <div className={classes.paper}>
          <Avatar className={classes.paper}>
            <LockOutlinedIcon />
          </Avatar>
          <form className={classes.form} noValidate onSubmit={onSubmit}>
            <Button className={classes.submit} type="submit" fullWidth variant="contained" color="primary">
              Sign in with GitHub
            </Button>
          </form>
        </div>
      </Grid>
    </>
  )
}

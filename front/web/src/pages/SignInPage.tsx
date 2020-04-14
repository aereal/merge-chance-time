import React, { FC, FormEvent } from "react"
import Grid from "@material-ui/core/Grid"
import Button from "@material-ui/core/Button"
import makeStyles from "@material-ui/core/styles/makeStyles"
import Avatar from "@material-ui/core/Avatar"
import LockOutlinedIcon from "@material-ui/icons/LockOutlined"
import { signIn } from "../effects/authentication"
import { routes } from "../routes"

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

const redirect = async (): Promise<void> => {
  await routes.root.push()
  return
}

export const SignInPage: FC = () => {
  const classes = useStyles()

  const onSubmit = async (event: FormEvent): Promise<void> => {
    event.preventDefault()
    await signIn()
    await routes.root.push()
  }

  return (
    <>
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
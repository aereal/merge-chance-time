import React, { FC, ComponentType } from "react"
import Snackbar, { SnackbarOrigin } from "@material-ui/core/Snackbar"
import SnackbarContent from "@material-ui/core/SnackbarContent"
import makeStyles from "@material-ui/core/styles/makeStyles"
import { Theme as DefaultTheme } from "@material-ui/core/styles/createMuiTheme"
import green from "@material-ui/core/colors/green"
import red from "@material-ui/core/colors/red"
import CheckCircleOutlineIcon from "@material-ui/icons/CheckCircleOutline"
import ErrorOutlineIcon from "@material-ui/icons/ErrorOutline"

type Kind = "success" | "error"

interface StylesProps {
  readonly kind: Kind
}

interface NotificationProps {
  readonly message: string
  readonly open: boolean
  readonly onClose: () => void
}

interface BaseNotificationProps extends StylesProps, NotificationProps {
  readonly icon: ComponentType<{ className?: string }>
}

const colors: Record<Kind, string> = {
  success: green[600],
  error: red[600],
}

const useStyles = makeStyles<DefaultTheme, StylesProps>((theme) => ({
  content: {
    backgroundColor: ({ kind }) => colors[kind],
  },
  message: {
    display: "flex",
    alignItems: "center",
  },
  icon: {
    fontSize: 20,
    opacity: 0.9,
    marginRight: theme.spacing(1),
  },
}))

const anchorOrigin: SnackbarOrigin = { vertical: "top", horizontal: "center" }

export const Notification: FC<BaseNotificationProps> = ({ kind, icon: Icon, message, open, onClose }) => {
  const classes = useStyles({ kind })
  return (
    <Snackbar anchorOrigin={anchorOrigin} autoHideDuration={3000} open={open} onClose={onClose}>
      <SnackbarContent
        className={classes.content}
        message={
          <span className={classes.message}>
            <Icon className={classes.icon} /> {message}
          </span>
        }
      ></SnackbarContent>
    </Snackbar>
  )
}

export const ErrorNotification: FC<NotificationProps> = (props) => (
  <Notification kind="error" icon={ErrorOutlineIcon} {...props} />
)

export const SuccessNotification: FC<NotificationProps> = (props) => (
  <Notification kind="success" icon={CheckCircleOutlineIcon} {...props} />
)

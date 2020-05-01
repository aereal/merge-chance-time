import React, { FC, useState, FormEventHandler, FormEvent } from "react"
import TextField from "@material-ui/core/TextField"
import Grid from "@material-ui/core/Grid"
import makeStyles from "@material-ui/core/styles/makeStyles"
import gql from "graphql-tag"
import { useMutation } from "@apollo/react-hooks"
import { useFetchState } from "../effects/fetch-state"
import { RepoDetailFragment } from "./__generated__/RepoDetailFragment"
import { UpdateRepoConfig, UpdateRepoConfigVariables } from "./__generated__/UpdateRepoConfig"
import { PrimaryButton } from "./PrimaryButton"
import { ErrorNotification, SuccessNotification } from "./Notification"
import { ScheduleRange } from "./ScheduleRange"

export const REPO_DETAIL_FRAGMENT = gql`
  fragment RepoDetailFragment on Repository {
    owner {
      login
    }
    name
    config {
      startSchedule
      stopSchedule
      mergeAvailable
    }
  }
`

const UPDATE_REPO_CONFIG = gql`
  mutation UpdateRepoConfig($owner: String!, $name: String!, $startSchedule: String, $stopSchedule: String) {
    updateRepositoryConfig(
      owner: $owner
      name: $name
      config: { startSchedule: $startSchedule, stopSchedule: $stopSchedule }
    )
  }
`

const useStyles = makeStyles((theme) => ({
  actions: {
    marginTop: theme.spacing(4),
  },
}))

interface RepoDetailProps {
  readonly repo: RepoDetailFragment
}

export const RepoDetail: FC<RepoDetailProps> = ({ repo }) => {
  const classes = useStyles()
  const [startSchedule, setStartSchedule] = useState(repo.config?.startSchedule)
  const [stopSchedule, setStopSchedule] = useState(repo.config?.stopSchedule)
  const { fetchState, start, succeed, fail, ready } = useFetchState()
  const [doUpdate] = useMutation<UpdateRepoConfig, UpdateRepoConfigVariables>(UPDATE_REPO_CONFIG)

  const handleChangeStart: FormEventHandler<HTMLTextAreaElement | HTMLInputElement> = (event) => {
    setStartSchedule(event.currentTarget.value)
  }
  const handleChangeStop: FormEventHandler<HTMLTextAreaElement | HTMLInputElement> = (event) => {
    setStopSchedule(event.currentTarget.value)
  }
  const handleSubmit = async (event: FormEvent): Promise<void> => {
    event.preventDefault()
    start()
    const { errors } = await doUpdate({
      variables: {
        owner: repo.owner.login,
        name: repo.name,
        startSchedule,
        stopSchedule,
      },
    })
    const errn = errors?.length ?? 0
    if (errn > 0) {
      fail(errors![0])
      return
    }
    succeed()
  }
  const handleErrorClose = () => {
    ready()
  }
  const handleCompleteClose = () => {
    ready()
  }

  return (
    <>
      <form noValidate onSubmit={handleSubmit}>
        <ScheduleRange />
        <div className={classes.actions}>
          <PrimaryButton disabled={fetchState.kind === "started"} type="submit">
            Save
          </PrimaryButton>
        </div>
      </form>
      {<SuccessNotification message="Success" open={fetchState.kind === "successful"} onClose={handleCompleteClose} />}
      {fetchState.kind === "failure" && (
        <ErrorNotification
          message={fetchState.error.message}
          open={fetchState.kind === "failure"}
          onClose={handleErrorClose}
        />
      )}
    </>
  )
}

import React, { FC, useState, FormEvent } from "react"
import makeStyles from "@material-ui/core/styles/makeStyles"
import gql from "graphql-tag"
import { useMutation } from "@apollo/react-hooks"
import { GraphQLError } from "graphql"
import FormControlLabel from "@material-ui/core/FormControlLabel"
import Switch from "@material-ui/core/Switch"
import ExpansionPanel from "@material-ui/core/ExpansionPanel"
import ExpansionPanelDetails from "@material-ui/core/ExpansionPanelDetails"
import ExpansionPanelSummary from "@material-ui/core/ExpansionPanelSummary"
import Typography from "@material-ui/core/Typography"
import ExpandMoreIcon from "@material-ui/icons/ExpandMore"
import { useFetchState } from "../effects/fetch-state"
import { MergeChanceSchedulesToUpdate, MergeChanceScheduleToUpdate } from "../globalTypes"
import {
  RepoDetailFragment,
  RepoDetailFragment_config_schedules as Schedules,
} from "./__generated__/RepoDetailFragment"
import { UpdateRepoConfig, UpdateRepoConfigVariables } from "./__generated__/UpdateRepoConfig"
import { PrimaryButton } from "./PrimaryButton"
import { ErrorNotification, SuccessNotification } from "./Notification"
import { ScheduleRange, SCHEDULES_FRAGMENT } from "./ScheduleRange"
import { MergeChanceScheduleFragment } from "./__generated__/MergeChanceScheduleFragment"

export const REPO_DETAIL_FRAGMENT = gql`
  fragment RepoDetailFragment on Repository {
    owner {
      login
    }
    name
    config {
      mergeAvailable
      schedules {
        ...SchedulesFragment
      }
    }
  }
  ${SCHEDULES_FRAGMENT}
`

const UPDATE_REPO_CONFIG = gql`
  mutation UpdateRepoConfig($owner: String!, $name: String!, $config: RepositoryConfigToUpdate!) {
    updateRepositoryConfig(owner: $owner, name: $name, config: $config)
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
  const [schedules, setSchedules] = useState<Schedules>(
    repo.config?.schedules ?? {
      __typename: "MergeChanceSchedules",
      sunday: null,
      monday: null,
      tuesday: null,
      wednesday: null,
      thursday: null,
      friday: null,
      saturday: null,
    }
  )
  const [mergeAvailable, setMergeAvailable] = useState(repo.config?.mergeAvailable ?? false)
  const { fetchState, start, succeed, fail, ready } = useFetchState()
  const [doUpdate] = useMutation<UpdateRepoConfig, UpdateRepoConfigVariables>(UPDATE_REPO_CONFIG)

  const handleMergeAvailabilityChanged = (): void => {
    setMergeAvailable((prev) => !prev)
  }
  const handleChanged = (updated: Schedules): void => {
    setSchedules(updated)
  }
  const handleSubmit = async (event: FormEvent): Promise<void> => {
    event.preventDefault()
    start()
    let errors: GraphQLError[] | undefined
    try {
      const ret = await doUpdate({
        variables: {
          owner: repo.owner.login,
          name: repo.name,
          config: {
            schedules: toInput(schedules),
            mergeAvailable,
          },
        },
      })
      errors = ret.errors
    } catch (e) {
      errors = [e]
    }
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
        <ScheduleRange schedules={schedules} onChanged={handleChanged} />
        <ExpansionPanel>
          <ExpansionPanelSummary expandIcon={<ExpandMoreIcon />}>
            <Typography>Extended actions</Typography>
          </ExpansionPanelSummary>
          <ExpansionPanelDetails>
            <FormControlLabel
              label="Merge Availability"
              control={<Switch color="primary" checked={mergeAvailable} onChange={handleMergeAvailabilityChanged} />}
            />
          </ExpansionPanelDetails>
        </ExpansionPanel>
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

const scheduleToInput = (schedule: MergeChanceScheduleFragment | null): MergeChanceScheduleToUpdate | null =>
  schedule ? { startHour: schedule.startHour, stopHour: schedule.stopHour } : null

const toInput = (schedules: Schedules): MergeChanceSchedulesToUpdate => ({
  sunday: scheduleToInput(schedules.sunday),
  monday: scheduleToInput(schedules.monday),
  tuesday: scheduleToInput(schedules.tuesday),
  wednesday: scheduleToInput(schedules.wednesday),
  thursday: scheduleToInput(schedules.thursday),
  friday: scheduleToInput(schedules.friday),
  saturday: scheduleToInput(schedules.saturday),
})

import React, { FC } from "react"
import gql from "graphql-tag"
import { weekdays, Weekday } from "../schedule"
import {
  WeekdayRangeSlider,
  MergeChanceScheduleRange,
  OnUpdateValue,
  MERGE_CHANCE_SCHEDULE_FRAGMENT,
} from "./WeekdayRangeSlider"
import { SchedulesFragment } from "./__generated__/SchedulesFragment"

interface ScheduleRangeProps {
  readonly schedules: SchedulesFragment
  readonly onChanged: (updated: SchedulesFragment) => void
}

export const SCHEDULES_FRAGMENT = gql`
  fragment SchedulesFragment on MergeChanceSchedules {
    sunday {
      ...MergeChanceScheduleFragment
    }
    monday {
      ...MergeChanceScheduleFragment
    }
    tuesday {
      ...MergeChanceScheduleFragment
    }
    wednesday {
      ...MergeChanceScheduleFragment
    }
    thursday {
      ...MergeChanceScheduleFragment
    }
    friday {
      ...MergeChanceScheduleFragment
    }
    saturday {
      ...MergeChanceScheduleFragment
    }
  }
  ${MERGE_CHANCE_SCHEDULE_FRAGMENT}
`

const normalizeValues = (value: number | number[]): MergeChanceScheduleRange => {
  if (typeof value === "number") {
    return { startHour: value, stopHour: value }
  }
  if (value.length !== 2) {
    throw new Error("Invalid value length")
  }
  const [from, to] = value
  return { startHour: from, stopHour: to }
}

export const ScheduleRange: FC<ScheduleRangeProps> = ({ schedules, onChanged }) => {
  const sliderHandler = (weekday: Weekday): OnUpdateValue => (value) => {
    onChanged({ ...schedules, [weekday]: value === null ? null : normalizeValues(value) })
  }
  const handlers = weekdays.reduce(
    (handlers, wd) => ({
      ...handlers,
      [wd]: sliderHandler(wd),
    }),
    {} as { [k: string]: OnUpdateValue }
  )

  return (
    <>
      {weekdays.map((wd) => (
        <WeekdayRangeSlider weekday={wd} scheduleRange={schedules[wd]} onUpdateValue={handlers[wd]} key={`wd${wd}`} />
      ))}
    </>
  )
}

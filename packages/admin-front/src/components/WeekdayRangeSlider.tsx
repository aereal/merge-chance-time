import gql from "graphql-tag"
import React, { FC, ChangeEvent } from "react"
import Slider from "@material-ui/core/Slider"
import red from "@material-ui/core/colors/red"
import indigo from "@material-ui/core/colors/indigo"
import makeStyles from "@material-ui/core/styles/makeStyles"
import Switch from "@material-ui/core/Switch"
import FormGroup from "@material-ui/core/FormGroup"
import FormControlLabel from "@material-ui/core/FormControlLabel"
import { Weekday, sunday, saturday, wholeDay } from "../schedule"
import { MergeChanceScheduleFragment } from "./__generated__/MergeChanceScheduleFragment"

export type OnUpdateValue = (value: number | number[] | null) => void
export interface MergeChanceScheduleRange {
  startHour: number
  stopHour: number
}

interface WeekdayRangeSliderProps {
  readonly weekday: Weekday
  readonly scheduleRange: MergeChanceScheduleFragment | null
  readonly onUpdateValue: OnUpdateValue
}

export const MERGE_CHANCE_SCHEDULE_FRAGMENT = gql`
  fragment MergeChanceScheduleFragment on MergeChanceSchedule {
    startHour
    stopHour
  }
`

const useSyles = makeStyles({
  sundayLabel: {
    color: red[500],
  },
  saturdayLabel: {
    color: indigo[500],
  },
})

export const WeekdayRangeSlider: FC<WeekdayRangeSliderProps> = ({ weekday, onUpdateValue, scheduleRange }) => {
  const { sundayLabel, saturdayLabel } = useSyles()

  const available = scheduleRange !== null
  const { startHour, stopHour } = scheduleRange ?? wholeDay
  const range = [startHour, stopHour]
  const colors: Partial<Record<Weekday, string>> = {
    [saturday]: saturdayLabel,
    [sunday]: sundayLabel,
  }
  const handleRangeChange = (_: ChangeEvent<{}>, value: number | number[]): void => {
    onUpdateValue(value)
  }
  const handleSwitchChange = (_: ChangeEvent<HTMLInputElement>, checked: boolean): void => {
    onUpdateValue(checked ? range : null)
  }

  return (
    <>
      <FormGroup row>
        <FormControlLabel
          className={colors[weekday]}
          label={weekday}
          control={<Switch color="primary" checked={available} onChange={handleSwitchChange} />}
        />
      </FormGroup>
      <Slider
        disabled={!available}
        marks
        valueLabelDisplay="auto"
        step={1}
        min={wholeDay.startHour}
        max={wholeDay.stopHour}
        value={range}
        onChange={handleRangeChange}
      />
    </>
  )
}

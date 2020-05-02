import React, { FC, ChangeEvent } from "react"
import Slider from "@material-ui/core/Slider"
import red from "@material-ui/core/colors/red"
import indigo from "@material-ui/core/colors/indigo"
import makeStyles from "@material-ui/core/styles/makeStyles"
import Switch from "@material-ui/core/Switch"
import FormGroup from "@material-ui/core/FormGroup"
import FormControlLabel from "@material-ui/core/FormControlLabel"
import { Weekday, sunday, saturday, wholeDay } from "../schedule"

export type OnUpdateValue = (value: number | number[] | null) => void
export type MergeChanceScheduleRange = [number, number]

interface WeekdayRangeSliderProps {
  readonly weekday: Weekday
  readonly scheduleRange: MergeChanceScheduleRange | null
  readonly onUpdateValue: OnUpdateValue
}

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
  const colors: Partial<Record<Weekday, string>> = {
    [saturday]: saturdayLabel,
    [sunday]: sundayLabel,
  }
  const handleRangeChange = (_: ChangeEvent<{}>, value: number | number[]): void => {
    onUpdateValue(value)
  }
  const handleSwitchChange = (_: ChangeEvent<HTMLInputElement>, checked: boolean): void => {
    onUpdateValue(checked ? scheduleRange ?? wholeDay : null)
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
        min={wholeDay[0]}
        max={wholeDay[1]}
        value={scheduleRange ?? wholeDay}
        onChange={handleRangeChange}
      />
    </>
  )
}

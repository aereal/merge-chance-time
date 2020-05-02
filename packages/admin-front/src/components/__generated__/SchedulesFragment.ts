/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL fragment: SchedulesFragment
// ====================================================

export interface SchedulesFragment_sunday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface SchedulesFragment_monday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface SchedulesFragment_tuesday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface SchedulesFragment_wednesday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface SchedulesFragment_thursday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface SchedulesFragment_friday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface SchedulesFragment_saturday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface SchedulesFragment {
  readonly __typename: "MergeChanceSchedules";
  readonly sunday: SchedulesFragment_sunday | null;
  readonly monday: SchedulesFragment_monday | null;
  readonly tuesday: SchedulesFragment_tuesday | null;
  readonly wednesday: SchedulesFragment_wednesday | null;
  readonly thursday: SchedulesFragment_thursday | null;
  readonly friday: SchedulesFragment_friday | null;
  readonly saturday: SchedulesFragment_saturday | null;
}

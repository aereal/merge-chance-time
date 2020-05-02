/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL fragment: RepoDetailFragment
// ====================================================

export interface RepoDetailFragment_owner {
  readonly __typename: "User" | "Organization";
  readonly login: string;
}

export interface RepoDetailFragment_config_schedules_sunday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface RepoDetailFragment_config_schedules_monday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface RepoDetailFragment_config_schedules_tuesday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface RepoDetailFragment_config_schedules_wednesday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface RepoDetailFragment_config_schedules_thursday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface RepoDetailFragment_config_schedules_friday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface RepoDetailFragment_config_schedules_saturday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface RepoDetailFragment_config_schedules {
  readonly __typename: "MergeChanceSchedules";
  readonly sunday: RepoDetailFragment_config_schedules_sunday | null;
  readonly monday: RepoDetailFragment_config_schedules_monday | null;
  readonly tuesday: RepoDetailFragment_config_schedules_tuesday | null;
  readonly wednesday: RepoDetailFragment_config_schedules_wednesday | null;
  readonly thursday: RepoDetailFragment_config_schedules_thursday | null;
  readonly friday: RepoDetailFragment_config_schedules_friday | null;
  readonly saturday: RepoDetailFragment_config_schedules_saturday | null;
}

export interface RepoDetailFragment_config {
  readonly __typename: "RepositoryConfig";
  readonly mergeAvailable: boolean;
  readonly schedules: RepoDetailFragment_config_schedules;
}

export interface RepoDetailFragment {
  readonly __typename: "Repository";
  readonly owner: RepoDetailFragment_owner;
  readonly name: string;
  readonly config: RepoDetailFragment_config | null;
}

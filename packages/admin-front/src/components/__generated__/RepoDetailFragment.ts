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

export interface RepoDetailFragment_config {
  readonly __typename: "RepositoryConfig";
  readonly startSchedule: string;
  readonly stopSchedule: string;
  readonly mergeAvailable: boolean;
}

export interface RepoDetailFragment {
  readonly __typename: "Repository";
  readonly owner: RepoDetailFragment_owner;
  readonly name: string;
  readonly config: RepoDetailFragment_config | null;
}

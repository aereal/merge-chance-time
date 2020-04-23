/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL fragment: RepoSummary
// ====================================================

export interface RepoSummary_owner {
  readonly __typename: "User" | "Organization";
  readonly login: string;
}

export interface RepoSummary {
  readonly __typename: "Repository";
  readonly fullName: string;
  readonly name: string;
  readonly owner: RepoSummary_owner;
}

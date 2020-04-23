/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: GetRepoDetail
// ====================================================

export interface GetRepoDetail_repository_owner {
  readonly __typename: "User" | "Organization";
  readonly login: string;
}

export interface GetRepoDetail_repository {
  readonly __typename: "Repository";
  readonly owner: GetRepoDetail_repository_owner;
  readonly name: string;
}

export interface GetRepoDetail {
  readonly repository: GetRepoDetail_repository | null;
}

export interface GetRepoDetailVariables {
  readonly owner: string;
  readonly name: string;
}

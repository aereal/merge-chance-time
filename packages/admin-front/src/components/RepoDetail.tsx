import React, { FC } from "react"
import gql from "graphql-tag"
import { RepoDetailFragment } from "./__generated__/RepoDetailFragment"

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

interface RepoDetailProps {
  readonly repo: RepoDetailFragment
}

export const RepoDetail: FC<RepoDetailProps> = ({ repo }) => (
  <>
    <pre>{JSON.stringify(repo.config)}</pre>
  </>
)

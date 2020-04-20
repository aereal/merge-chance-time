import React, { FC } from "react"
import List from "@material-ui/core/List"
import { Repo, RepoSummary } from "./RepoSummary"

interface ReposListProps {
  readonly repos: Repo[]
}

export const ReposList: FC<ReposListProps> = ({ repos }) => (
  <List>
    {repos.map((repo) => (
      <RepoSummary repo={repo} key={repo.fullName} />
    ))}
  </List>
)

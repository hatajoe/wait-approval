# wait-approval

wait-approval pause GitHub Action workflow and wait for approval from someone.

- approval keyword is `/approve`
  - approve step and subsequent steps are going to run
- deny keyword is `/deny`
  - revoke step and subsequent steps are not going to run.

# Usage

```
steps:
  - uses: hatajoe/wait-approval@v1
    with:
      github-pull-request-number: ${{ github.event.pull_request.number }}
    env:
      GITHUB_TOKEN: ${{ secret.GITHUB_TOKEN }}
```


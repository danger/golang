# Contribution Guidelines

We welcome contributions from the community! Whether you're fixing a bug, adding a feature, or improving documentation,
your contributions help us improve this project for everyone!

## Discussions

If you're unsure about anything, or are not sure if a change should be made, feel free to ask questions by opening a
discussion.

## Before You Start

* Before submitting a PR, verify that an issue exists that describes the bug fix or feature you want to contribute. If
  there's no issue yet, please create one.
* If you don't have access to the repo, fork to your on your GitHub user account, make code changes there, and then
  create a PR against the `main` branch.
* Add tests for any new features or bug fixes. Ideally, each PR increases the test coverage.
* Follow the existing code style (follow [Effective Go](https://go.dev/doc/effective_go) where no obvious convention
  exists).
* Use the pull request template provided, and fill in all the details you can.
* Document newly introduced methods and classes with godoc-style comments, and add inline comments to code that is not
  self-documenting.
* Separate unrelated changes into multiple PRs.
* Commit messages should follow the conventions outlined at https://www.conventionalcommits.org/.

## Writing Tests

- Where possible, we try to write unit tests in a [table-driven](https://go.dev/wiki/TableDrivenTests) way (something
  you’ll likely be familiar with if you’re familiar with Go).
- This doesn’t mean a test _has_ to be written in this way but
  if you find yourself in a situation where you are testing various inputs to a function that lead to different outputs, a
  table-driven test will work well!



## Creating the PR

### PR Title

Please follow the [Conventional Commits](https://www.conventionalcommits.org/) guidelines for PR titles (and commit
titles).

### PR Description

Please split the PR description into the categories defined by https://keepachangelog.com, for example:

```
### Added
- v1.1 Brazilian Portuguese translation.
### Changed
- Use frontmatter title & description in each language version template
```

If you are looking for help, or aren’t sure if something is an issue or not, please create a post under [Discussions](../discussions)
instead.

## Feedback

Lastly, **don’t let review comments discourage you**! We appreciate any and all public contributions. Luno has an
internal set of standards and guidelines we are committed to upholding on all of our repos, so we may request changes to
bring any PRs in line with those.
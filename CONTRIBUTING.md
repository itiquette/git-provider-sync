<!--
SPDX-FileCopyrightText: 2025 The Git Provider Sync Authors

SPDX-License-Identifier: CC0-1.0

-->

# Contributing

[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-%23FE5196?style=for-the-badge&logo=conventionalcommits&logoColor=white)](https://conventionalcommits.org)
[![Code of Conduct](https://img.shields.io/badge/Simple%20Code%20of%20Conduct-1.0-4baaaa.svg?style=for-the-badge)](CODE_OF_CONDUCT.md)
[![DCO - developer certificate of origin](https://img.shields.io/badge/DCO-Developer%20Certificate%20of%20Origin-lightyellow?style=for-the-badge)](docs/dco.txt)

Welcome! We are excited that you are interested in contributing to the project!
However, there are some things you might need to know, so please browse the following:

## Ways to Contribute

There are multiple ways of getting involved:

As a new contributor, you are in an excellent position to give us feedback to our project.
For example, you could:

- Fix or report a bug.

- Suggest improvements to code, tests and documentation.

- Report/fix problems found during installing or developer environments.

- Add suggestions for something else that is missing.

## Community Guideline

Be nice and respectful to each other.

We follow the [Simple Contributor Code Of Conduct](CODE_OF_CONDUCT.md).

## File an Issue

Please check briefly if there already exists an Issue with your topic.
If so, you can just add a comment to that with your information instead of creating a new Issue.

### Report a bug

Reporting bugs is a good and easy way to contribute.

To do this, open an Issue that summarizes the bug and set the label to "bug".

### Suggest a feature

To request a new feature you should summarize the desired functionality and its use case.
Set the Issue label to "feature".

## Contribute Code, Documentation and more

You want to contribute code, documentation or other improvements.
Great, however, there are some practical points to check to make sure that everything runs as smoothly as possible.

- Discuss your plans beforehand to check that your contribution fits the project goals.

- Check the list of open Issues. Either assign an existing issue to yourself, or create a new one that you would like to work on, and discuss your ideas and use cases.

- Follow the project convention and style regarding test, code and documentation, commit style etc.

- The project can decide to decline a contribution not following the general project guidelines, or deemed to not fit into the general project goal/architecture.

- Make sure you have an understanding of the [Pull Request Process](#pull-request-process)

- You agree to that in general, all contributions to this project will be released under the **inbound=outbound** norm, that is, contributions are submitted under the same terms as the project licenses.

In a more formal way - \'Unless You explicitly state otherwise, any Contribution intentionally submitted for inclusion in the Work by You to the Licensor shall be under the terms and conditions of the projects License, without any additional terms or conditions.\'

- [Sign your commits](#commit-guideline).

## Issues and Pull Request Feedback

The listed project maintainers of the project are doing their best to review and/or respond to Issues.
**If the project is not listed as archived, it is maintained.**.
You should expect feedback, for an Issue or a Pull Request, at some point.

## Note

Feedback might mean a lot of things depending on the scope of your
Issue/Pull Request. What you should expect is at least a comment on your
Issue.

For non trivial and proposed bigger contributions, please discuss the contribution with the project first.
There might be occasions where the cost of maintenance and review have to be agreed upon before it can be merged and reviewed.

The quality of the given information in your Pull Request/Issue will affect the feedback loop.
So please keep the focus and topic, and be sure to give as much relevant information as possible.
It is highly recommended that you fill in the requested fields when submitting a contribution.

## Pull Request Process

We use the [Fork-and-Pull model](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/getting-started/about-collaborative-development-models#fork-and-pull-model):

### Steps to Submit a PR:

1. **Fork and clone** the repository
2. **Create a feature branch** from main: `git checkout -b feature/my-feature`
3. **Make your changes** following our code style guidelines
4. **Test thoroughly**: `just verify` must pass
5. **Commit** with sign-off and GPG signature: `git commit --signoff --gpg-sign -m "feat: add new feature"`
6. **Push** to your fork: `git push origin feature/my-feature`
7. **Open a Pull Request** with:

- Clear description of changes
- Reference to related issues
- Screenshots/examples if applicable

### PR Requirements:

- [ ] All tests pass (`just test`)
- [ ] Linting passes (`just lint-go`)
- [ ] Commits are signed and follow conventional commits
- [ ] Documentation updated if needed
- [ ] Breaking changes are clearly documented

## Commit Guideline

### DCO - Signoff and Signing Required

All commits must be both signed-off (DCO agreement) and cryptographically signed.

**Required flags**: `git commit --signoff --gpg-sign -m "your message"`

- **Signoff** (`--signoff`): Confirms you have rights to contribute ([DCO](https://developercertificate.org/))
- **Sign** (`--gpg-sign`): Cryptographic verification of commit author

**Setup**: Configure GPG or SSH signing - see [GitHub's guide](https://github.blog/changelog/2022-08-23-ssh-commit-verification-now-supported/)

### Commit Standard

Aim for a clear human readable commit history:

- Make sure you [Sign-Off](#commit-guideline) your commits.

- In general

  - Use the [Conventional Commit standard](https://www.conventionalcommits.org).

  - Group relevant changes in commits, avoid scope creep and keep focus on the relevant issue.

  - Your commit messages should tell a human reader what will it do when the commit is applied.

  - Make your commit message/s easily human readable in a expected way, use imperative verb form:\

    - A Conventional Commit example:\
            *fix: add a null pointer check to MyMethod parameter*\
            Would be read as \'When this **fix** is applied it will...​
            **add a null pointer check to MyMethod parameter**\'.

## Reporting security issues

If you discover a security issue, please bring it to our attention.

If the vulnerability is a widely known issue, detected by various Vulnerability Scanning sources it might be okay to file an public Issue.

However, if any uncertainty around this, please **DO NOT** file a public issue, see [Security information](SECURITY.md) for how to handle this.

Security reports are **greatly** appreciated.

## Development Guidelines

For development setup, build instructions, and testing guidelines, see the [DEVELOPMENT Guide](./DEVELOPMENT.adoc).

## Writing style and Translations

Here are a few guidelines regarding text and documentation.

- Aim to keep the documentation EASY to read, and avoid the official "agency authority" style.

- Don't be too verbose, bullet points are good in this context.

- Be concise, in terminology, and avoid longer explanations, link instead.

- Strive to use [one-sentence-per-line](https://sembr.org/) when writing in MarkDown or AsciiDoc.

English is the projects primary language, and any translations are done on a best effort basis.
This implies that for any contributions to the translated version, make sure that the English primary version contains the corresponding change.

## FOSS Standards

This project follows the principles outlined in the following standards:

- License compliance with the [REUSE specification](https://reuse.software/)

- Commits in the [Conventional Commits format](https://www.conventionalcommits.org/en/v1.0.0/)

- Changelog in the [Keep-A-Changelog format](https://keepachangelog.com/en/1.1.0/)

***Happy contributing!***

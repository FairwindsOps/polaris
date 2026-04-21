#!/usr/bin/env sh
# Wrap goreleaser with branch/tag-specific env (see .goreleaser.yml templates)
# and, on non-tag CI runs, a temporary git tag.

cleanup() {
  if [ "${CIRCLE_TAG}" == "" ] ; then
    echo "${this_script} deleting git tag ${temporary_git_tag} for goreleaser"
    unset GORELEASER_CURRENT_TAG
    git tag -d ${temporary_git_tag}
  fi
}

set -eE # errexit and errtrace
trap 'cleanup' ERR
this_script="$(basename $0)"
hash goreleaser
if [ "${TMPDIR}" == "" ] ; then
  export TMPDIR="/tmp"
  echo "${this_script} temporarily set the TMPDIR environment variable to ${TMPDIR}, used for a temporary GOBIN environment variable"
fi

export GORELEASER_SKIP_FEATURE_DOCKER_TAGS=false
export GORELEASER_SKIP_RELEASE=true
if [ "${CIRCLE_TAG}" == "" ] ; then
  # Create a temporary tag for goreleaser, incrementing the last tag.
  last_git_tag="$(git describe --tags --abbrev=0 2>/dev/null)"
  if [ "${last_git_tag}" == "" ] ; then
    echo "${this_script} is unable to determine the last git tag so a temporary tag can be created, using: git describe --tags --abbrev=0"
    exit 1
  fi
  if [ "$(git config user.email)" == "" ] ; then
    # git will use this env var as its user.email.
    # git tag -m is used in case tags are manually pushed by accident,
    # however git tag -m requires an email.
    export EMAIL='goreleaser_ci@fairwinds.com'
    echo "${this_script} using ${EMAIL} temporarily as the git user.email"
  fi
  temporary_git_tag=$(echo "${last_git_tag}" | awk -F. '{$NF = $NF + 1;} 1' | sed 's/ /./g')-rc
  echo "${this_script} creating temporary git tag ${temporary_git_tag} for goreleaser, the last real tag is ${last_git_tag}"
  # The -f is included to overwrite existing tags, perhaps from previous CI jobs.
  git tag -f -m "temporary local tag for goreleaser" ${temporary_git_tag}
  export GORELEASER_CURRENT_TAG=${temporary_git_tag}
  # Use an adjusted git feature branch name as a docker tag; export so goreleaser receives .Env.FEATURE_DOCKER_TAG.
  export FEATURE_DOCKER_TAG=$(echo "${CIRCLE_BRANCH:0:26}" | sed 's/[^a-zA-Z0-9]/-/g' | sed 's/-\+$//')
  echo "${this_script} also using docker tag ${FEATURE_DOCKER_TAG} since ${CIRCLE_BRANCH} is a feature branch"
else
  export GORELEASER_CURRENT_TAG=${CIRCLE_TAG}
  echo "${this_script} setting GORELEASER_SKIP_RELEASE to false, and GORELEASER_SKIP_FEATURE_DOCKER_TAGS to true, because CIRCLE_TAG is set"
  export GORELEASER_SKIP_FEATURE_DOCKER_TAGS=true
  export GORELEASER_SKIP_RELEASE=false
fi

echo "${this_script} using git tag ${GORELEASER_CURRENT_TAG}"
goreleaser --skip=sign "$@"
cleanup
